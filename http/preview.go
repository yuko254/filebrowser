//go:generate go-enum --sql --marshal --names --file $GOFILE
package fbhttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/img"
	"github.com/gorilla/mux"
	"github.com/spf13/afero"
)

/*
ENUM(
thumb
big
)
*/
type PreviewSize int

type ImgService interface {
	FormatFromExtension(ext string) (img.Format, error)
	Resize(ctx context.Context, in io.Reader, width, height int, out io.Writer, options ...img.Option) error
}

type FileCache interface {
	Store(ctx context.Context, key string, value []byte) error
	Load(ctx context.Context, key string) ([]byte, bool, error)
	Delete(ctx context.Context, key string) error
}

func previewHandler(imgSvc ImgService, fileCache FileCache, enableThumbnails, resizePreview bool) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Download {
			return http.StatusAccepted, nil
		}
		vars := mux.Vars(r)

		previewSize, err := ParsePreviewSize(vars["size"])
		if err != nil {
			return http.StatusBadRequest, err
		}

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       "/" + vars["path"],
			Modify:     d.user.Perm.Modify,
			Expand:     true,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		if err != nil {
			return errToStatus(err), err
		}

		setContentDisposition(w, r, file)

		switch file.Type {
		case "image":
			return handleImagePreview(w, r, imgSvc, fileCache, file, previewSize, enableThumbnails, resizePreview)
		case "video":
			return handleVideoPreview(w, r, imgSvc, fileCache, file, previewSize, enableThumbnails, resizePreview)
		default:
			return http.StatusNotImplemented, fmt.Errorf("can't create preview for %s type", file.Type)
		}
	})
}

var sem = make(chan int, 3)

func handleVideoPreview(
	w http.ResponseWriter,
	r *http.Request,
	_ ImgService,
	fileCache FileCache,
	file *files.FileInfo,
	previewSize PreviewSize,
	_, _ bool,
) (int, error) {
	path := afero.FullBaseFsPath(file.Fs.(*afero.BasePathFs), file.Path)

	cacheKey := previewCacheKey(file, previewSize)
	resizedImage, ok, err := fileCache.Load(r.Context(), cacheKey)
	if err != nil {
		return errToStatus(err), err
	}
	if !ok {
		sem <- 1
		defer func() { <-sem }()

		// Define size and quality based on previewSize
		var (
			width, height int
			quality       string
		)

		quality = "90"
		switch previewSize {
		case PreviewSizeThumb:
			width, height = 256, 256
		case PreviewSizeBig:
			width, height = 1280, 720
		default:
			return http.StatusBadRequest, fmt.Errorf("unknown preview size")
		}

		// Build filter strings for GPU and software fallback
		var gpuFilter, swFilter string
		if previewSize == PreviewSizeThumb {
			// Thumbnail: crop to square then scale
			swFilter = fmt.Sprintf("thumbnail=n=300,crop=w='min(iw\\,ih)':h='min(iw\\,ih)',scale=%d:%d", width, height)
			// GPU: select frame with thumbnail_cuda, download to CPU, then crop & scale (software)
			gpuFilter = fmt.Sprintf("thumbnail_cuda=n=300,scale_cuda=%d:%d,hwdownload,format=nv12", width, height)
		} else {
			// Big preview: scale to fit within dimensions, preserving aspect ratio
			swFilter = fmt.Sprintf("thumbnail=n=300,scale='min(%d\\,iw)':'min(%d\\,ih)'", width, height)
			// GPU: pure GPU pipeline – select and scale on GPU, then download
			gpuFilter = fmt.Sprintf("thumbnail_cuda=n=300,scale_cuda=%d:%d:force_original_aspect_ratio=decrease,hwdownload,format=nv12", width, height)
		}

		var stderr bytes.Buffer
		var GPU bool
		var CPU bool

		// Attempt 1: GPU‑accelerated command
		cmd := exec.Command("ffmpeg",
			"-y",
			"-hwaccel", "cuda",
			"-hwaccel_output_format", "cuda",
			"-skip_frame", "nokey",
			"-ss", "5",
			"-i", path,
			"-vf", gpuFilter,
			"-quality", quality,
			"-frames:v", "1",
			"-c:v", "webp",
			"-f", "webp",
			"-",
		)
		cmd.Stderr = &stderr
		stdout, err := cmd.Output()
		if err != nil {
			// GPU failed – fall back to software
			lines := bytes.Split(stderr.Bytes(), []byte{'\n'})
			lines = lines[len(lines)-10:]
			cleanError := strings.TrimSpace(string(bytes.Join(lines, []byte{'\n'})))
			fmt.Printf("GPU preview failed for %s : %v, used software fallback.\n\n", file.Path, cleanError)
			GPU = false
			stderr.Reset()
		} else {
			GPU = true
		}

		if !GPU {
			// Attempt 2: Software fallback
			cmd = exec.Command("ffmpeg",
				"-y",
				"-skip_frame", "nokey",
				"-ss", "5",
				"-i", path,
				"-vf", swFilter,
				"-quality", quality,
				"-frames:v", "1",
				"-c:v", "webp",
				"-f", "webp",
				"-",
			)
			cmd.Stderr = &stderr
			stdout, err = cmd.Output()
			if err != nil {
				// software fallback failed - fallback to generous ffmpeg settings
				lines := bytes.Split(stderr.Bytes(), []byte{'\n'})
				lines = lines[len(lines)-10:]
				cleanError := strings.TrimSpace(string(bytes.Join(lines, []byte{'\n'})))
				fmt.Printf("CPU preview failed for %s : %v, used last fallback.\n\n", file.Path, cleanError)
				CPU = false
				stderr.Reset()
			} else {
				CPU = true
			}
		}
		
		if !CPU && !GPU {
			// Attempt 3
			filter := "thumbnail,crop=w='min(iw\\,ih)':h='min(iw\\,ih)',scale=256:256"
			cmd = exec.Command("ffmpeg",
				"-y",
				"-i", path,
				"-vf", filter,
				"-quality", quality,
				"-frames:v", "1",
				"-c:v", "webp",
				"-f", "webp",
				"-",
			)
			cmd.Stderr = &stderr
			stdout, err = cmd.Output()
			if err != nil {
				return errToStatus(err), err
			}
		}

		resizedImage = stdout

		go func() {
			if err := fileCache.Store(context.Background(), cacheKey, resizedImage); err != nil {
				fmt.Printf("failed to cache resized image: %v", err)
			}
		}()
	}

	w.Header().Set("Cache-Control", "private")
	w.Header().Set("Content-Type", "image/webp")
	http.ServeContent(w, r, "", file.ModTime, bytes.NewReader(resizedImage))
	return 0, nil
}

func handleImagePreview(
	w http.ResponseWriter,
	r *http.Request,
	imgSvc ImgService,
	fileCache FileCache,
	file *files.FileInfo,
	previewSize PreviewSize,
	enableThumbnails, resizePreview bool,
) (int, error) {
	if (previewSize == PreviewSizeBig && !resizePreview) ||
		(previewSize == PreviewSizeThumb && !enableThumbnails) {
		return rawFileHandler(w, r, file)
	}

	format, err := imgSvc.FormatFromExtension(file.Extension)
	// Unsupported extensions directly return the raw data
	if errors.Is(err, img.ErrUnsupportedFormat) || format == img.FormatGif {
		return rawFileHandler(w, r, file)
	}
	if err != nil {
		return errToStatus(err), err
	}

	cacheKey := previewCacheKey(file, previewSize)
	resizedImage, ok, err := fileCache.Load(r.Context(), cacheKey)
	if err != nil {
		return errToStatus(err), err
	}
	if !ok {
		resizedImage, err = createPreview(imgSvc, fileCache, file, previewSize)
		if err != nil {
			return errToStatus(err), err
		}
	}

	w.Header().Set("Cache-Control", "private")
	http.ServeContent(w, r, file.Name, file.ModTime, bytes.NewReader(resizedImage))

	return 0, nil
}

func createPreview(imgSvc ImgService, fileCache FileCache,
	file *files.FileInfo, previewSize PreviewSize) ([]byte, error) {
	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	var (
		width   int
		height  int
		options []img.Option
	)

	switch previewSize {
	case PreviewSizeBig:
		width = 1080
		height = 1080
		options = append(options, img.WithMode(img.ResizeModeFit), img.WithQuality(img.QualityMedium))
	case PreviewSizeThumb:
		width = 256
		height = 256
		options = append(options, img.WithMode(img.ResizeModeFill), img.WithQuality(img.QualityLow), img.WithFormat(img.FormatJpeg))
	default:
		return nil, img.ErrUnsupportedFormat
	}

	buf := &bytes.Buffer{}
	if err := imgSvc.Resize(context.Background(), fd, width, height, buf, options...); err != nil {
		return nil, err
	}

	go func() {
		cacheKey := previewCacheKey(file, previewSize)
		if err := fileCache.Store(context.Background(), cacheKey, buf.Bytes()); err != nil {
			fmt.Printf("failed to cache resized image: %v", err)
		}
	}()

	return buf.Bytes(), nil
}

func previewCacheKey(f *files.FileInfo, previewSize PreviewSize) string {
	return fmt.Sprintf("%x%x%x", f.RealPath(), f.ModTime.Unix(), previewSize)
}
