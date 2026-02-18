<template>
  <div class="card floating small-edit-shortcut">
    <div class="card-content">
      <label>{{ $t("prompts.name") }}</label>
      <input
        id="focus-prompt"
        v-model="name"
      >
      <p style="margin-top: 8px; font-size: 0.9em; color: var(--muted)">
        {{ path }}
      </p>
    </div>
    <div class="card-action">
      <button
        class="button button--flat button--grey"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
        @click="closeHovers"
      >
        {{ $t("buttons.cancel") }}
      </button>

      <button
        class="button button--flat"
        :aria-label="$t('buttons.delete')"
        :title="$t('buttons.delete')"
        @click="doDelete"
      >
        {{ $t("buttons.delete") }}
      </button>

      <button
        class="button button--flat button--primary"
        :aria-label="$t('buttons.save')"
        :title="$t('buttons.save')"
        @click="doSave"
      >
        {{ $t("buttons.save") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState } from "pinia";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import * as usersApi from "@/api/users";

export default {
  name: "EditShortcut",
  computed: {
    ...mapState(useLayoutStore, ["currentPrompt"]),
  },
  data() {
    return {
      name: "",
      path: "",
      idx: -1,
    };
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    async doSave() {
      try {
        const auth = useAuthStore();
        const shortcuts = (auth.user?.shortcuts || []).slice();
        if (this.idx >= 0 && this.idx < shortcuts.length) {
          shortcuts[this.idx] = { ...shortcuts[this.idx], name: this.name };
          await usersApi.update({ id: auth.user?.id, shortcuts }, [
            "shortcuts",
          ]);
          auth.updateUser({ shortcuts });
        }
        this.closeHovers();
      } catch (e) {
        this.$showError(e);
      }
    },
    async doDelete() {
      if (!confirm(this.$t("prompts.confirmDeleteShortcut"))) return;
      try {
        const auth = useAuthStore();
        const shortcuts = (auth.user?.shortcuts || []).slice();
        if (this.idx >= 0 && this.idx < shortcuts.length) {
          shortcuts.splice(this.idx, 1);
          await usersApi.update({ id: auth.user?.id, shortcuts }, [
            "shortcuts",
          ]);
          auth.updateUser({ shortcuts });
        }
        this.closeHovers();
      } catch (e) {
        this.$showError(e);
      }
    },
  },
  mounted() {
    const props = this.currentPrompt?.props || {};
    this.idx = props.idx ?? -1;
    const auth = useAuthStore();
    const s = auth.user?.shortcuts?.[this.idx];
    this.name = s?.name || "";
    this.path = s?.path || "";
  },
};
</script>

<style scoped>
.small-edit-shortcut {
  width: 320px;
}
.small-edit-shortcut .card-content input {
  width: 100%;
}
</style>
