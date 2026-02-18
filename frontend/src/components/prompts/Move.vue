<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.move") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.moveMessage") }}</p>
      <file-list
        ref="fileList"
        :exclude="excludedFolders"
        tabindex="1"
        @update:selected="(val) => (dest = val)"
      />
    </div>

    <div
      class="card-action"
      style="display: flex; align-items: center; justify-content: space-between"
    >
      <template v-if="user.perm.create">
        <button
          class="button button--flat"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
          style="justify-self: left"
          @click="$refs.fileList.createDir()"
        >
          <span>{{ $t("sidebar.newFolder") }}</span>
        </button>
      </template>
      <div>
        <button
          class="button button--flat button--grey"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
          tabindex="3"
          @click="closeHovers"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          id="focus-prompt"
          class="button button--flat"
          :disabled="$route.path === dest"
          :aria-label="$t('buttons.move')"
          :title="$t('buttons.move')"
          tabindex="2"
          @click="move"
        >
          {{ $t("buttons.move") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState, mapWritableState } from "pinia";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import FileList from "./FileList.vue";
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import * as upload from "@/utils/upload";
import { removePrefix } from "@/api/utils";

export default {
  name: "Move",
  components: { FileList },
  inject: ["$showError"],
  data: function () {
    return {
      current: window.location.pathname,
      dest: null,
    };
  },
  computed: {
    ...mapState(useFileStore, ["req", "selected"]),
    ...mapState(useAuthStore, ["user"]),
    ...mapWritableState(useFileStore, ["reload", "preselect"]),
    excludedFolders() {
      return this.selected
        .filter((idx) => this.req.items[idx].isDir)
        .map((idx) => this.req.items[idx].url);
    },
  },
  methods: {
    ...mapActions(useLayoutStore, ["showHover", "closeHovers"]),
    move: async function (event) {
      event.preventDefault();
      const items = [];

      for (const item of this.selected) {
        items.push({
          from: this.req.items[item].url,
          to: this.dest + encodeURIComponent(this.req.items[item].name),
          name: this.req.items[item].name,
        });
      }

      const action = async (overwrite, rename) => {
        buttons.loading("move");

        await api
          .move(items, overwrite, rename)
          .then(() => {
            buttons.success("move");
            this.preselect = removePrefix(items[0].to);
            if (this.user.redirectAfterCopyMove)
              this.$router.push({ path: this.dest });
            else this.reload = true;
          })
          .catch((e) => {
            buttons.done("move");
            this.$showError(e);
          });
      };

      const dstItems = (await api.fetch(this.dest)).items;
      const conflict = upload.checkConflict(items, dstItems);

      let overwrite = false;
      let rename = false;

      if (conflict) {
        this.showHover({
          prompt: "replace-rename",
          confirm: (event, option) => {
            overwrite = option == "overwrite";
            rename = option == "rename";

            event.preventDefault();
            this.closeHovers();
            action(overwrite, rename);
          },
        });

        return;
      }

      action(overwrite, rename);
    },
  },
};
</script>
