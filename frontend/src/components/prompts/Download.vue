<template>
  <div
    id="download"
    class="card floating"
  >
    <div class="card-title">
      <h2>{{ t("prompts.download") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ t("prompts.downloadMessage") }}</p>

      <button
        v-for="(ext, format) in formats"
        id="focus-prompt"
        :key="format"
        class="button button--block"
        @click="layoutStore.currentPrompt?.confirm(format)"
      >
        {{ ext }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useLayoutStore } from "@/stores/layout";

const layoutStore = useLayoutStore();

const { t } = useI18n();

const formats = {
  zip: "zip",
  tar: "tar",
  targz: "tar.gz",
  tarbz2: "tar.bz2",
  tarxz: "tar.xz",
  tarlz4: "tar.lz4",
  tarsz: "tar.sz",
  tarbr: "tar.br",
  tarzst: "tar.zst",
};
</script>
