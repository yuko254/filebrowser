<template>
  <div v-show="active" @click="closeHovers" class="overlay"></div>
  <nav :class="{ active }">
    <template v-if="isLoggedIn">
      <button @click="toAccountSettings" class="action">
        <i class="material-icons">person</i>
        <span>{{ user.username }}</span>
      </button>
      <button
        class="action my-files-action"
        @click="toggleMyFiles"
        :aria-label="$t('sidebar.myFiles')"
        :title="$t('sidebar.myFiles')"
      >
        <i class="material-icons">folder</i>
        <span>{{ $t("sidebar.myFiles") }}</span>
        <i class="material-icons expand-icon">{{ myFilesOpen ? 'expand_less' : 'expand_more' }}</i>
      </button>

      <div v-show="myFilesOpen" class="my-files-list">
        <button class="action" @click="toRoot">
          <i class="material-icons">home</i>
          <span>{{ $t('files.home') }}</span>
        </button>

        <button class="action" @click="addPath">
          <i class="material-icons">add</i>
          <span>{{ $t('prompts.addPath') }}</span>
        </button>

        <template v-for="(s, idx) in user.shortcuts || []" :key="idx">
          <div class="shortcut-row">
            <button class="action shortcut" @click="openShortcut(s)">
              <i class="material-icons">chevron_right</i>
              <span>{{ displayShortcut(s) }}</span>
            </button>

            <div class="shortcut-actions">
              <button
                class="icon-btn"
                @click.stop="showHover({ prompt: 'editShortcut', props: { idx } })"
                :aria-label="$t('buttons.edit')"
                :title="$t('buttons.edit')"
              >
                <i class="material-icons">more_vert</i>
              </button>
            </div>
          </div>
        </template>
      </div>

      <div v-if="user.perm.create">
        <button
          @click="showHover('newDir')"
          class="action"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
        >
          <i class="material-icons">create_new_folder</i>
          <span>{{ $t("sidebar.newFolder") }}</span>
        </button>

        <button
          @click="showHover('newFile')"
          class="action"
          :aria-label="$t('sidebar.newFile')"
          :title="$t('sidebar.newFile')"
        >
          <i class="material-icons">note_add</i>
          <span>{{ $t("sidebar.newFile") }}</span>
        </button>
      </div>

      <div v-if="user.perm.admin">
        <button
          class="action"
          @click="toGlobalSettings"
          :aria-label="$t('sidebar.settings')"
          :title="$t('sidebar.settings')"
        >
          <i class="material-icons">settings_applications</i>
          <span>{{ $t("sidebar.settings") }}</span>
        </button>
      </div>
      <button
        v-if="canLogout"
        @click="logout"
        class="action"
        id="logout"
        :aria-label="$t('sidebar.logout')"
        :title="$t('sidebar.logout')"
      >
        <i class="material-icons">exit_to_app</i>
        <span>{{ $t("sidebar.logout") }}</span>
      </button>
    </template>
    <template v-else>
      <router-link
        v-if="!hideLoginButton"
        class="action"
        to="/login"
        :aria-label="$t('sidebar.login')"
        :title="$t('sidebar.login')"
      >
        <i class="material-icons">exit_to_app</i>
        <span>{{ $t("sidebar.login") }}</span>
      </router-link>

      <router-link
        v-if="signup"
        class="action"
        to="/login"
        :aria-label="$t('sidebar.signup')"
        :title="$t('sidebar.signup')"
      >
        <i class="material-icons">person_add</i>
        <span>{{ $t("sidebar.signup") }}</span>
      </router-link>
    </template>

    <div
      class="credits"
      v-if="isFiles && !disableUsedPercentage"
      style="width: 90%; margin: 2em 2.5em 3em 2.5em"
    >
      <progress-bar :val="usage.usedPercentage" size="small"></progress-bar>
      <br />
      {{ usage.used }} of {{ usage.total }} used
    </div>

    <p class="credits">
      <span>
        <span v-if="disableExternal">File Browser</span>
        <a
          v-else
          rel="noopener noreferrer"
          target="_blank"
          href="https://github.com/filebrowser/filebrowser"
          >File Browser</a
        >
        <span> {{ " " }} {{ version }}</span>
      </span>
      <span>
        <a @click="help">{{ $t("sidebar.help") }}</a>
      </span>
    </p>
  </nav>
</template>

<script>
import { reactive } from "vue";
import { mapActions, mapState } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import * as auth from "@/utils/auth";
import url from "@/utils/url";
import {
  version,
  signup,
  hideLoginButton,
  disableExternal,
  disableUsedPercentage,
  noAuth,
  logoutPage,
  loginPage,
} from "@/utils/constants";
import { files as api } from "@/api";
import * as usersApi from "@/api/users";
import ProgressBar from "@/components/ProgressBar.vue";
import prettyBytes from "pretty-bytes";

const USAGE_DEFAULT = { used: "0 B", total: "0 B", usedPercentage: 0 };

export default {
  name: "sidebar",
  setup() {
    const usage = reactive(USAGE_DEFAULT);
    return { usage, usageAbortController: new AbortController() };
  },
  components: {
    ProgressBar,
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useAuthStore, ["user", "isLoggedIn"]),
    ...mapState(useFileStore, ["isFiles", "reload"]),
    ...mapState(useLayoutStore, ["currentPromptName"]),
    active() {
      return this.currentPromptName === "sidebar";
    },
    signup: () => signup,
    hideLoginButton: () => hideLoginButton,
    version: () => version,
    disableExternal: () => disableExternal,
    disableUsedPercentage: () => disableUsedPercentage,
    canLogout: () => !noAuth && (loginPage || logoutPage !== "/login"),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers", "showHover"]),
    abortOngoingFetchUsage() {
      this.usageAbortController.abort();
    },
    async fetchUsage() {
      const path = this.$route.path.endsWith("/")
        ? this.$route.path
        : this.$route.path + "/";
      let usageStats = USAGE_DEFAULT;
      if (this.disableUsedPercentage) {
        return Object.assign(this.usage, usageStats);
      }
      try {
        this.abortOngoingFetchUsage();
        this.usageAbortController = new AbortController();
        const usage = await api.usage(path, this.usageAbortController.signal);
        usageStats = {
          used: prettyBytes(usage.used, { binary: true }),
          total: prettyBytes(usage.total, { binary: true }),
          usedPercentage: Math.round((usage.used / usage.total) * 100),
        };
      } finally {
        return Object.assign(this.usage, usageStats);
      }
    },
    toRoot() {
      this.$router.push({ path: "/files" });
      this.closeHovers();
    },
    toAccountSettings() {
      this.$router.push({ path: "/settings/profile" });
      this.closeHovers();
    },
    toGlobalSettings() {
      this.$router.push({ path: "/settings/global" });
      this.closeHovers();
    },
    help() {
      this.showHover("help");
    },
    logout: auth.logout,
    toggleMyFiles() {
      this.myFilesOpen = !this.myFilesOpen;
    },
    addPath() {
      this.showHover({ prompt: "addPath" });
    },
    openShortcut(s) {
      // strip any query params from stored path
      const raw = (s.path || "").split("?")[0];
      let targetPath = "";
      if (raw.startsWith("/files")) {
        targetPath = raw;
      } else {
        targetPath = `/files${url.encodePath(raw)}`;
      }

      if (s.system) {
        this.$router.push({ path: targetPath, query: { system: "true" } });
      } else {
        this.$router.push({ path: targetPath });
      }
      this.closeHovers();
    },
    async deleteShortcut(idx) {
      if (!confirm(this.$t("prompts.confirmDeleteShortcut"))) return;
      try {
        const shortcuts = (this.user.shortcuts || []).slice();
        shortcuts.splice(idx, 1);
        await usersApi.update({ id: this.user.id, shortcuts }, ["shortcuts"]);
        const authStore = useAuthStore();
        authStore.updateUser({ shortcuts });
      } catch (e) {
        this.$showError(e);
      }
    },
    displayShortcut(s) {
      if (!s) return "";
      if (s.name && s.name.trim() !== "") return s.name;
      let p = s.path || "";
      // strip query string
      const qidx = p.indexOf("?");
      if (qidx !== -1) p = p.slice(0, qidx);
      // remove leading /files if present
      if (p.startsWith("/files")) p = p.replace(/^\/files/, "");
      // normalize trailing slash for display
      if (p.endsWith("/") && p !== "/") p = p.slice(0, -1);
      return p || "/";
    },
    startRename(idx) {
      this.editingIndex = idx;
      const s = (this.user.shortcuts || [])[idx];
      this.editingName = s?.name || s?.path || "";
    },
    cancelRename() {
      this.editingIndex = -1;
      this.editingName = "";
    },
    async saveRename(idx) {
      const name = (this.editingName || "").trim();
      if (!name) return;
      try {
        const shortcuts = (this.user.shortcuts || []).slice();
        shortcuts[idx] = { ...shortcuts[idx], name };
        await usersApi.update({ id: this.user.id, shortcuts }, ["shortcuts"]);
        const authStore = useAuthStore();
        authStore.updateUser({ shortcuts });
        this.cancelRename();
      } catch (e) {
        this.$showError(e);
      }
    },
  },
  data() {
    return {
      myFilesOpen: false,
      editingIndex: -1,
      editingName: "",
    };
  },
  watch: {
    $route: {
      handler(to) {
        if (to.path.includes("/files")) {
          this.fetchUsage();
        }
      },
      immediate: true,
    },
  },
  unmounted() {
    this.abortOngoingFetchUsage();
  },
};
</script>

<style scoped>
.shortcut-row { display: flex; align-items: center; gap: 8px; }
.shortcut-row .action.shortcut { flex: 1; text-align: left; }
.shortcut-actions { margin-left: auto; display: flex; gap: 6px; }
.icon-btn { background: transparent; border: none; cursor: pointer; padding: 4px; }
.shortcut-edit-input { width: 100%; }

.icon-btn i {
  color: rgba(102,102,102,0.9) !important;
  font-size: 18px;
}
.icon-btn:hover i {
  color: rgba(30,30,30,0.95) !important;
}
.icon-btn:hover {
  background: rgba(0,0,0,0.04);
  border-radius: 4px;
}
.my-files-list { margin-top: 0.5em; max-height: calc(100vh - 200px); overflow-y: auto; overflow-x: hidden; padding-right: 6px; position: relative; }
</style>
