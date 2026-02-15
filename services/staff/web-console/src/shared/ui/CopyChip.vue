<template>
  <VTooltip :text="tooltip">
    <template #activator="{ props: tipProps }">
      <VChip
        v-bind="tipProps"
        class="font-weight-bold"
        size="small"
        variant="tonal"
        :prepend-icon="icon"
        @click="copy"
      >
        {{ label }}: <span class="ml-1 value">{{ value }}</span>
      </VChip>
    </template>
  </VTooltip>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import { useSnackbarStore } from "./feedback/snackbar-store";

const props = withDefaults(
  defineProps<{
    label: string;
    value: string;
    icon?: string;
  }>(),
  { icon: "mdi-content-copy" },
);

const snackbar = useSnackbarStore();
const { t } = useI18n({ useScope: "global" });

const tooltip = computed(() => (props.value ? t("common.copy") : ""));

async function copy(): Promise<void> {
  if (!props.value) return;
  try {
    await navigator.clipboard.writeText(props.value);
    snackbar.success(t("common.copied"));
  } catch {
    snackbar.error(t("errors.copyFailed"));
  }
}
</script>

<style scoped>
.value {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
