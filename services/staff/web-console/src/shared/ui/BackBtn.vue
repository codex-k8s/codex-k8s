<template>
  <VBtn
    v-bind="attrs"
    icon
    size="large"
    variant="tonal"
    class="back-btn"
    :title="label"
    :aria-label="label"
  >
    <VIcon icon="mdi-arrow-left" size="26" />
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
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
}
</style>

