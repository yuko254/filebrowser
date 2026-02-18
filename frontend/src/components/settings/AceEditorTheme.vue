<template>
  <select
    name="selectAceEditorTheme"
    :value="aceEditorTheme"
    @change="change"
  >
    <option
      v-for="theme in themes"
      :key="theme.theme"
      :value="theme.theme"
    >
      {{ theme.name }}
    </option>
  </select>
</template>

<script setup lang="ts">
import { type SelectHTMLAttributes } from "vue";
import { themes } from "ace-builds/src-noconflict/ext-themelist";

defineProps<{
  aceEditorTheme: string;
}>();

const emit = defineEmits<{
  (e: "update:aceEditorTheme", val: string | null): void;
}>();

const change = (event: Event) => {
  emit("update:aceEditorTheme", (event.target as SelectHTMLAttributes)?.value);
};
</script>
