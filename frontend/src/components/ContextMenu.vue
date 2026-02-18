<template>
  <div
    v-show="show"
    ref="contextMenu"
    class="context-menu"
    :style="{
      top: `${props.pos.y}px`,
      left: `${left}px`,
    }"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, onUnmounted } from "vue";

const emit = defineEmits(["hide"]);
const props = defineProps<{ show: boolean; pos: { x: number; y: number } }>();
const contextMenu = ref<HTMLElement | null>(null);

const left = computed(() => {
  return Math.min(
    props.pos.x,
    window.innerWidth - (contextMenu.value?.clientWidth ?? 0)
  );
});

const hideContextMenu = () => {
  emit("hide");
};

watch(
  () => props.show,
  (val) => {
    if (val) {
      document.addEventListener("click", hideContextMenu);
    } else {
      document.removeEventListener("click", hideContextMenu);
    }
  }
);

onUnmounted(() => {
  document.removeEventListener("click", hideContextMenu);
});
</script>
