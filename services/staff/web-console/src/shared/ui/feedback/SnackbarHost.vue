<template>
  <VSnackbar v-model="openModel" :timeout="timeout" :color="color" location="bottom right">
    {{ text }}
    <template #actions>
      <VBtn variant="text" @click="snackbar.close">{{ t("common.close") }}</VBtn>
    </template>
  </VSnackbar>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import { useSnackbarStore } from "./snackbar-store";

const snackbar = useSnackbarStore();
const { t } = useI18n({ useScope: "global" });

const openModel = computed({
  get: () => snackbar.open,
  set: (v: boolean) => {
    if (v) return;
    snackbar.close();
  },
});

const text = computed(() => snackbar.current?.text || "");
const timeout = computed(() => snackbar.current?.timeoutMs ?? 0);
const color = computed(() => {
  switch (snackbar.current?.kind) {
    case "success":
      return "success";
    case "error":
      return "error";
    default:
      return "info";
  }
});
</script>

