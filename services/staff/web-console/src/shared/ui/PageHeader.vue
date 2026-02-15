<template>
  <div class="page-header">
    <div class="page-header-main">
      <div v-if="hasLeading" class="page-header-leading">
        <slot name="leading" />
      </div>

      <div class="page-header-text">
        <div class="text-h5 font-weight-bold">
          {{ title }}
        </div>
        <div v-if="hint" class="text-body-2 text-medium-emphasis mt-1">
          {{ hint }}
        </div>
      </div>
    </div>

    <div class="page-header-actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, useSlots } from "vue";

defineProps<{
  title: string;
  hint?: string;
}>();

const slots = useSlots();
const hasLeading = computed(() => Boolean(slots.leading));
</script>

<style scoped>
.page-header {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  justify-content: space-between;
}
.page-header-main {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  min-width: 0;
}
.page-header-leading {
  display: inline-flex;
  align-items: center;
  padding-top: 2px;
}
.page-header-text {
  min-width: 0;
}
.page-header-actions {
  display: inline-flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
  justify-content: flex-end;
}
</style>
