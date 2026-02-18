package users

import (
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/rules"
)

// ViewMode describes a view mode.
type ViewMode string

const (
	ListViewMode   ViewMode = "list"
	MosaicViewMode ViewMode = "mosaic"
)

// User describes a user.
type User struct {
	ID                    uint          `storm:"id,increment" json:"id"`
	Username              string        `storm:"unique" json:"username"`
	Password              string        `json:"password"`
	Scope                 string        `json:"scope"`
	Locale                string        `json:"locale"`
	LockPassword          bool          `json:"lockPassword"`
	ViewMode              ViewMode      `json:"viewMode"`
	SingleClick           bool          `json:"singleClick"`
	RedirectAfterCopyMove bool          `json:"redirectAfterCopyMove"`
	Perm                  Permissions   `json:"perm"`
	Commands              []string      `json:"commands"`
	Sorting               files.Sorting `json:"sorting"`
	Fs                    afero.Fs      `json:"-" yaml:"-"`
	Rules                 []rules.Rule  `json:"rules"`
	HideDotfiles          bool          `json:"hideDotfiles"`
	DateFormat            bool          `json:"dateFormat"`
	AceEditorTheme        string        `json:"aceEditorTheme"`
	Shortcuts             []Shortcut    `json:"shortcuts"`
}

// GetRules implements rules.Provider.
func (u *User) GetRules() []rules.Rule {
	return u.Rules
}

var checkableFields = []string{
	"Username",
	"Password",
	"Scope",
	"ViewMode",
	"Commands",
	"Sorting",
	"Shortcuts",
	"Rules",
}

// Clean cleans up a user and verifies if all its fields
// are alright to be saved.
func (u *User) Clean(baseScope string, fields ...string) error {
	if len(fields) == 0 {
		fields = checkableFields
	}

	for _, field := range fields {
		switch field {
		case "Username":
			if u.Username == "" {
				return fberrors.ErrEmptyUsername
			}
		case "Password":
			if u.Password == "" {
				return fberrors.ErrEmptyPassword
			}
		case "ViewMode":
			if u.ViewMode == "" {
				u.ViewMode = ListViewMode
			}
		case "Commands":
			if u.Commands == nil {
				u.Commands = []string{}
			}
		case "Sorting":
			if u.Sorting.By == "" {
				u.Sorting.By = "name"
			}

		case "Shortcuts":
			if u.Shortcuts == nil {
				u.Shortcuts = []Shortcut{}
			}

			// sanitize shortcut paths: remove query strings, normalize slashes,
			// remove leading /files prefix and handle Windows drive literal prefixed with '/'.
			for i := range u.Shortcuts {
				p := u.Shortcuts[i].Path
				// strip query/hash
				for j := 0; j < len(p); j++ {
					if p[j] == '?' || p[j] == '#' {
						p = p[:j]
						break
					}
				}
				// normalize backslashes to slashes
				p = strings.ReplaceAll(p, "\\\\", "/")
				// remove /files prefix if present
				p = strings.TrimPrefix(p, "/files")
				// handle Windows absolute paths that may have been saved as "/C:/path"
				if len(p) >= 3 && p[0] == '/' && ((p[1] >= 'A' && p[1] <= 'Z') || (p[1] >= 'a' && p[1] <= 'z')) && p[2] == ':' {
					p = strings.TrimPrefix(p, "/")
				}
				if p == "" {
					p = "/"
				}
				u.Shortcuts[i].Path = p
			}
		case "Rules":
			if u.Rules == nil {
				u.Rules = []rules.Rule{}
			}
		}
	}

	if u.Fs == nil {
		scope := u.Scope
		// If user provided an absolute scope (e.g. on Windows: E:/dir),
		// use it directly instead of joining with baseScope to avoid
		// producing paths like <base><abs-path> which are invalid.
		if filepath.IsAbs(scope) {
			scope = filepath.Clean(scope)
			u.Fs = afero.NewBasePathFs(afero.NewOsFs(), scope)
		} else {
			scope = filepath.Join(baseScope, filepath.Join("/", scope))
			u.Fs = afero.NewBasePathFs(afero.NewOsFs(), scope)
		}
	}

	return nil
}

// FullPath gets the full path for a user's relative path.
func (u *User) FullPath(path string) string {
	return afero.FullBaseFsPath(u.Fs.(*afero.BasePathFs), path)
}
