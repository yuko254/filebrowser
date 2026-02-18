<template>
  <div>
    <ul class="file-list">
      <li
        v-for="item in items"
        :key="item.name"
        role="button"
        tabindex="0"
        :aria-label="item.name"
        :aria-selected="selected == item.url"
        :data-url="item.url"
        :data-isdir="item.isDir"
        @click="itemClick"
        @touchstart="touchstart"
        @dblclick="next"
      >
        {{ item.name }}
      </li>
    </ul>

    <p>
      {{ $t("prompts.currentlyNavigating") }} <code>{{ nav }}</code>.
    </p>
  </div>
</template>

<script>
import { mapState, mapActions } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import url from "@/utils/url";
import { files } from "@/api";
import { StatusError } from "@/api/utils.js";

export default {
  name: "FileList",
  inject: ["$showError"],
  props: {
    exclude: {
      type: Array,
      default: () => [],
    },
    onlyDirs: {
      type: Boolean,
      default: false,
    },
    system: {
      type: Boolean,
      default: false,
    },
    initialFromStore: {
      type: Boolean,
      default: true,
    },
  },
  data: function () {
    return {
      items: [],
      touches: {
        id: "",
        count: 0,
      },
      selected: null,
      current: window.location.pathname,
      nextAbortController: new AbortController(),
      isSystem: false,
    };
  },
  computed: {
    ...mapState(useAuthStore, ["user"]),
    ...mapState(useFileStore, ["req"]),
    nav() {
      return decodeURIComponent(this.current);
    },
  },
  mounted() {
    if (this.initialFromStore) {
      this.fillOptions(this.req);
    }
  },
  unmounted() {
    this.abortOngoingNext();
  },
  methods: {
    ...mapActions(useLayoutStore, ["showHover"]),
    abortOngoingNext() {
      this.nextAbortController.abort();
    },
    fillOptions(req) {
      // Sets the current path and resets
      // the current items.
      this.current = req.url;
      this.items = [];

      this.$emit("update:selected", this.current);

      // If the path isn't the root path, show a button to navigate to the previous directory.
      // Prefer using the underlying resource `path` when present (used for system browsing).
      const isSystem = !!req.system;
      if (isSystem) {
        this.isSystem = true;
        const curPath = req.path || "/";
        if (curPath !== "/") {
          // compute parent underlying path and map it to a files URL
          const parent = url.removeLastDir(curPath);
          let parentUrl = `/files${parent}`;
          if (!parentUrl.endsWith("/")) parentUrl += "/";
          // append system query so backend serves OS FS
          parentUrl =
            parentUrl +
            (parentUrl.includes("?") ? "&system=true" : "?system=true");
          this.items.push({
            name: "..",
            url: parentUrl,
            isDir: true,
            system: true,
          });
        }
      } else {
        this.isSystem = false;
        if (req.url !== "/files/") {
          this.items.push({
            name: "..",
            url: url.removeLastDir(req.url) + "/",
          });
        }
      }

      // track system browsing mode
      this.isSystem = !!req.system;

      // If this folder is empty, finish here.
      if (req.items === null) return;

      // Otherwise we add every directory to the
      // move options.
      for (const item of req.items) {
        if (this.onlyDirs && !item.isDir) continue;
        if (this.exclude?.includes(item.url)) continue;

        this.items.push({
          name: item.name,
          url: item.url,
          isDir: item.isDir,
          system: !!item.system || false,
        });
      }
    },
    next: function (event) {
      // Retrieves the URL of the directory the user
      // just clicked in and fill the options with its
      // content.
      const uri = event.currentTarget.dataset.url;
      const isDir = event.currentTarget.dataset.isdir === "true";
      if (!isDir) return; // only navigate into directories
      this.abortOngoingNext();
      this.nextAbortController = new AbortController();
      let fetchUri = uri;
      // if this listing was returned as system browse, append query as a query param (not part of path)
      if (this.isSystem || this.system) {
        fetchUri = uri + (uri.includes("?") ? "&system=true" : "?system=true");
      }

      files
        .fetch(fetchUri, this.nextAbortController.signal)
        .then(this.fillOptions)
        .catch((e) => {
          if (e instanceof StatusError && e.is_canceled) {
            return;
          }
          this.$showError(e);
        });
    },
    touchstart(event) {
      const url = event.currentTarget.dataset.url;

      // In 300 milliseconds, we shall reset the count.
      setTimeout(() => {
        this.touches.count = 0;
      }, 300);

      // If the element the user is touching
      // is different from the last one he touched,
      // reset the count.
      if (this.touches.id !== url) {
        this.touches.id = url;
        this.touches.count = 1;
        return;
      }

      this.touches.count++;

      // If there is more than one touch already,
      // open the next screen.
      if (this.touches.count > 1) {
        this.next(event);
      }
    },
    itemClick: function (event) {
      if (this.user.singleClick) this.next(event);
      else this.select(event);
    },
    select: function (event) {
      // If the element is already selected, unselect it.
      if (this.selected === event.currentTarget.dataset.url) {
        this.selected = null;
        this.$emit("update:selected", this.current);
        return;
      }

      // Otherwise select the element.
      this.selected = event.currentTarget.dataset.url;
      this.$emit("update:selected", this.selected);
    },
    createDir: async function () {
      this.showHover({
        prompt: "newDir",
        action: null,
        confirm: null,
        props: {
          redirect: false,
          base: this.current === this.$route.path ? null : this.current,
        },
      });
    },
  },
};
</script>
