<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.replace") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.replaceMessage") }}</p>
    </div>

    <div class="card-action">
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
        class="button button--flat button--blue"
        :aria-label="$t('buttons.rename')"
        :title="$t('buttons.rename')"
        tabindex="2"
        @click="(event) => currentPrompt.confirm(event, 'rename')"
      >
        {{ $t("buttons.rename") }}
      </button>
      <button
        id="focus-prompt"
        class="button button--flat button--red"
        :aria-label="$t('buttons.replace')"
        :title="$t('buttons.replace')"
        tabindex="1"
        @click="(event) => currentPrompt.confirm(event, 'overwrite')"
      >
        {{ $t("buttons.replace") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState } from "pinia";
import { useLayoutStore } from "@/stores/layout";

export default {
  name: "ReplaceRename",
  computed: {
    ...mapState(useLayoutStore, ["currentPrompt"]),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
  },
};
</script>
