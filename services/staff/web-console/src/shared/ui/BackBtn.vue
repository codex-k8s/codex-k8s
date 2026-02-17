<template>
  <VBtn
    v-bind="attrs"
    icon
    variant="tonal"
    class="back-btn"
    :title="label"
    :aria-label="label"
  >
    <VIcon icon="mdi-arrow-left" size="22" />
  </VBtn>
</template>

<script setup lang="ts">
import { useAttrs } from "vue";

defineOptions({ inheritAttrs: false });

withDefaults(
  defineProps<{
    label?: string;
  }>(),
  { label: "Back" },
);

const attrs = useAttrs();
</script>

<style scoped>
.back-btn {
  width: 32px;
  height: 32px;
  min-width: 32px;
  border-radius: 8px;
  padding: 0;
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
}
</style>
