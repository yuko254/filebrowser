<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.addPath") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.addPathMessage") }}</p>
      <label>{{ $t("prompts.name") }}</label>
      <input
        v-model="name"
        class="input"
      >
      <file-list
        ref="fileList"
        :only-dirs="true"
        :system="true"
        :initial-from-store="false"
        tabindex="1"
        @update:selected="(val) => (selected = val)"
      />
    </div>

    <div
      class="card-action"
      style="display: flex; justify-content: flex-end; gap: 8px"
    >
      <button
        class="button button--flat button--grey"
        @click="closeHovers"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        id="focus-prompt"
        class="button"
        :disabled="!selected"
        @click="add"
      >
        {{ $t("buttons.add") }}
      </button>
    </div>
  </div>
</template>

<script>
import FileList from "./FileList.vue";
import { mapState } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { files as api } from "@/api";
import { users as usersApi } from "@/api";

export default {
  name: "AddPath",
  components: { FileList },
  inject: ["$showError"],
  data() {
    return {
      name: "",
      selected: null,
    };
  },
  computed: {
    ...mapState(useAuthStore, ["user"]),
  },
  mounted() {
    // initialize the file-list with system root (this PC) when admin
    const root = this.user?.perm?.admin ? "/files/?system=true" : "/files/";
    api
      .fetch(root)
      .then((res) => {
        const fl = this.$refs.fileList;
        if (fl && typeof fl.fillOptions === "function") {
          fl.fillOptions(res);
        }
      })
      .catch((e) => this.$showError(e));
  },
  methods: {
    add: async function () {
      try {
        const auth = useAuthStore();
        const user = auth.user;
        if (!this.selected) return;

        // derive the underlying path from the selected value (strip query and /files prefix)
        let underlyingPath = (this.selected || "").split("?")[0];
        if (underlyingPath.startsWith("/files")) {
          underlyingPath = underlyingPath.replace(/^\/files/, "");
        }
        underlyingPath = underlyingPath.replace(/\\/g, "/");
        if (!underlyingPath.startsWith("/"))
          underlyingPath = "/" + underlyingPath;

        const newEntry = {
          name: this.name || "",
          path: underlyingPath,
          system: !!user?.perm?.admin,
        };
        const newShortcuts = (user.shortcuts || []).concat(newEntry);
        await usersApi.update({ id: user.id, shortcuts: newShortcuts }, [
          "shortcuts",
        ]);
        auth.updateUser({ shortcuts: newShortcuts });
        this.closeHovers();
      } catch (e) {
        this.$showError(e);
      }
    },
    closeHovers() {
      const layout = useLayoutStore();
      layout.closeHovers();
    },
  },
};
</script>

<style scoped>
.card-content input.input {
  width: 100%;
  margin-bottom: 0.5em;
}
</style>
