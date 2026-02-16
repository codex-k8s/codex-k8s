<template>
  <div class="mt-4">
    <VRow density="compact" class="align-center">
      <VCol cols="12" md="6">
        <VSelect
          v-model="namespace"
          :items="namespaces"
          :label="t('cluster.namespace')"
          density="comfortable"
          hide-details
        />
      </VCol>
      <VCol cols="12" md="6" class="d-flex justify-end">
        <VChip :color="modeColor" variant="tonal" class="font-weight-bold">
          {{ modeLabel }}
        </VChip>
      </VCol>
    </VRow>

    <VAlert :type="modeAlertType" variant="tonal" class="mt-3">
      <div class="text-body-2">
        {{ modeHint }}
      </div>
    </VAlert>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import { useUiContextStore } from "../../features/ui-context/store";

const { t } = useI18n({ useScope: "global" });
const uiContext = useUiContextStore();

type SelectItem = { title: string; value: string };

const namespaces = computed<SelectItem[]>(() => {
  const project = "codex-k8s";
  const all = { title: t("context.allObjects"), value: "" };

  const ai = [`${project}-dev-1`, `${project}-dev-2`, `${project}-dev-3`];
  const production = [`${project}-production`];
  const prod = [`${project}-prod`];

  const values =
    uiContext.env === "ai"
      ? ai
      : uiContext.env === "production"
        ? production
        : uiContext.env === "prod"
          ? prod
          : [...ai, ...production, ...prod];

  return [all, ...values.map((v) => ({ title: v, value: v }))];
});

const namespace = computed({
  get: () => uiContext.namespace,
  set: (v) => uiContext.setNamespace(v),
});

const modeColor = computed(() => {
  switch (uiContext.clusterMode) {
    case "view-only":
      return "info";
    case "dry-run":
      return "warning";
    default:
      return "success";
  }
});

const modeAlertType = computed(() => {
  switch (uiContext.clusterMode) {
    case "view-only":
      return "info";
    case "dry-run":
      return "warning";
    default:
      return "success";
  }
});

const modeLabel = computed(() => {
  switch (uiContext.clusterMode) {
    case "view-only":
      return t("cluster.mode.viewOnly");
    case "dry-run":
      return t("cluster.mode.dryRun");
    default:
      return t("cluster.mode.normal");
  }
});

const modeHint = computed(() => {
  switch (uiContext.clusterMode) {
    case "view-only":
      return t("cluster.modeHint.viewOnly");
    case "dry-run":
      return t("cluster.modeHint.dryRun");
    default:
      return t("cluster.modeHint.normal");
  }
});
</script>
