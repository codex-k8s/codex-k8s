<template>
  <div class="page-header">
    <div>
      <div class="text-h5 font-weight-bold">
        {{ title }}
      </div>
      <div v-if="hint" class="text-body-2 text-medium-emphasis mt-1">
        {{ hint }}
      </div>
    </div>

    <div class="page-header-actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string;
  hint?: string;
}>();
</script>

<style scoped>
.page-header {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  justify-content: space-between;
}
.page-header-actions {
  display: inline-flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
  justify-content: flex-end;
}
</style>

