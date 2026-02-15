<template>
  <div>
    <PageHeader :title="title">
      <template #actions>
        <CopyChip :label="t('cluster.namespace')" :value="uiContext.namespace || '-'" icon="mdi-kubernetes" />
        <CopyChip :label="t('common.id')" :value="name" icon="mdi-identifier" />

        <VBtn variant="tonal" prepend-icon="mdi-arrow-left" :to="backTo">
          {{ t("common.back") }}
        </VBtn>
        <VBtn variant="tonal" prepend-icon="mdi-refresh" @click="mockReload">
          {{ t("common.refresh") }}
        </VBtn>
        <VBtn
          color="error"
          variant="tonal"
          prepend-icon="mdi-delete-outline"
          :disabled="uiContext.clusterMode === 'view-only'"
          @click="previewOpen = true"
        >
          {{ t("common.delete") }}
        </VBtn>
      </template>
    </PageHeader>

    <AdminClusterContextBar />

    <VTabs v-model="tab" class="mt-4">
      <VTab value="overview">{{ t("cluster.details.tabs.overview") }}</VTab>
      <VTab value="yaml">{{ t("cluster.details.tabs.yaml") }}</VTab>
      <VTab value="events">{{ t("cluster.details.tabs.events") }}</VTab>
      <VTab value="related">{{ t("cluster.details.tabs.related") }}</VTab>
      <VTab v-if="supportsLogs" value="logs">{{ t("cluster.details.tabs.logs") }}</VTab>
    </VTabs>

    <VWindow v-model="tab" class="mt-2">
      <VWindowItem value="overview">
        <VRow density="compact">
          <VCol cols="12" md="4">
            <VCard variant="outlined">
              <VCardText>
                <div class="text-caption text-medium-emphasis">{{ t("cluster.details.metrics.status") }}</div>
                <div class="text-h6 font-weight-bold">{{ mockStatus }}</div>
              </VCardText>
            </VCard>
          </VCol>
          <VCol cols="12" md="4">
            <VCard variant="outlined">
              <VCardText>
                <div class="text-caption text-medium-emphasis">{{ t("cluster.details.metrics.age") }}</div>
                <div class="text-h6 font-weight-bold">10d</div>
              </VCardText>
            </VCard>
          </VCol>
          <VCol cols="12" md="4">
            <VCard variant="outlined">
              <VCardText>
                <div class="text-caption text-medium-emphasis">{{ t("cluster.details.metrics.mode") }}</div>
                <div class="text-h6 font-weight-bold">{{ uiContext.clusterMode }}</div>
              </VCardText>
            </VCard>
          </VCol>
        </VRow>

        <VAlert type="info" variant="tonal" class="mt-4">
          <div class="text-body-2">
            {{ t("cluster.details.scaffoldNote") }}
          </div>
        </VAlert>
      </VWindowItem>

      <VWindowItem value="yaml">
        <VCard variant="outlined">
          <VCardText>
            <MonacoEditor v-model="yaml" language="yaml" height="520px" read-only />
          </VCardText>
        </VCard>
      </VWindowItem>

      <VWindowItem value="events">
        <VCard variant="outlined">
          <VCardText>
            <VList density="compact">
              <VListItem v-for="e in mockEvents" :key="e.at + ':' + e.type" :title="e.type" :subtitle="e.at" prepend-icon="mdi-bell-outline" />
            </VList>
          </VCardText>
        </VCard>
      </VWindowItem>

      <VWindowItem value="related">
        <VCard variant="outlined">
          <VCardText>
            <VList density="compact">
              <VListItem v-for="r in mockRelated" :key="r" :title="r" prepend-icon="mdi-link-variant" />
            </VList>
          </VCardText>
        </VCard>
      </VWindowItem>

      <VWindowItem v-if="supportsLogs" value="logs">
        <LogsViewer :lines="mockLogLines" :file-name="`${kind}-${name}.log`" @refresh="noopRefresh" />
      </VWindowItem>
    </VWindow>
  </div>

  <VDialog v-model="previewOpen" max-width="720">
    <VCard>
      <VCardTitle class="text-subtitle-1">{{ t("cluster.details.previewTitle") }}</VCardTitle>
      <VCardText>
        <VAlert
          v-if="uiContext.clusterMode === 'view-only'"
          type="info"
          variant="tonal"
          class="mb-3"
        >
          {{ t("cluster.details.viewOnlyWarning") }}
        </VAlert>
        <VAlert
          v-else-if="uiContext.clusterMode === 'dry-run'"
          type="warning"
          variant="tonal"
          class="mb-3"
        >
          {{ t("cluster.details.dryRunWarning") }}
        </VAlert>

        <MonacoEditor v-model="yaml" language="yaml" height="360px" read-only />
      </VCardText>
      <VCardActions>
        <VSpacer />
        <VBtn variant="text" @click="previewOpen = false">{{ t("common.cancel") }}</VBtn>
        <VBtn
          color="error"
          variant="tonal"
          :disabled="uiContext.clusterMode === 'view-only'"
          @click="confirmDelete"
        >
          {{ t("common.delete") }}
        </VBtn>
      </VCardActions>
    </VCard>
  </VDialog>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные k8s ресурсы через backend (RBAC + audit), YAML и Events, а также безопасные режимы view-only/dry-run.
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";

import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import CopyChip from "../../shared/ui/CopyChip.vue";
import LogsViewer from "../../shared/ui/LogsViewer.vue";
import PageHeader from "../../shared/ui/PageHeader.vue";
import MonacoEditor from "../../shared/ui/monaco/MonacoEditor.vue";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { useUiContextStore } from "../../features/ui-context/store";

const props = defineProps<{
  kind: "namespaces" | "configmaps" | "secrets" | "deployments" | "pods" | "jobs" | "pvc";
  name: string;
}>();

const { t } = useI18n({ useScope: "global" });
const uiContext = useUiContextStore();
const snackbar = useSnackbarStore();

const tab = ref<"overview" | "yaml" | "events" | "related" | "logs">("overview");
const previewOpen = ref(false);

const kind = computed(() => props.kind);
const name = computed(() => props.name);

const supportsLogs = computed(() => kind.value === "pods" || kind.value === "jobs");
const title = computed(() => `${t("cluster.details.title")} · ${kind.value}/${name.value}`);

const backTo = computed(() => {
  switch (kind.value) {
    case "namespaces":
      return { name: "cluster-namespaces" };
    case "configmaps":
      return { name: "cluster-configmaps" };
    case "secrets":
      return { name: "cluster-secrets" };
    case "deployments":
      return { name: "cluster-deployments" };
    case "pods":
      return { name: "cluster-pods" };
    case "jobs":
      return { name: "cluster-jobs" };
    default:
      return { name: "cluster-pvc" };
  }
});

const mockStatus = computed(() => (kind.value === "pods" ? "Running" : kind.value === "jobs" ? "Active" : "Ready"));

const yaml = ref(
  `apiVersion: v1\nkind: ${kind.value}\nmetadata:\n  name: ${name.value}\n  namespace: ${uiContext.namespace || "default"}\n`,
);

const mockEvents = [
  { type: "Scheduled", at: "2026-02-15T10:05:31Z" },
  { type: "Pulled", at: "2026-02-15T10:05:44Z" },
  { type: "Started", at: "2026-02-15T10:05:48Z" },
];

const mockRelated = ["Service/web-console", "ConfigMap/web-console-config", "Secret/github-oauth"];

const mockLogLines = [
  "[info] starting server on :8080",
  "[info] ready",
  "[warn] dry-run mode enabled",
  "[info] GET /health 200",
];

function mockReload(): void {
  snackbar.info(t("common.refresh"));
}

function noopRefresh(_tailLines: number): void {
  // scaffold: logs are static
  snackbar.info(t("common.refresh"));
}

function confirmDelete(): void {
  previewOpen.value = false;
  if (uiContext.clusterMode === "dry-run") {
    snackbar.success(t("cluster.details.dryRunOk"));
    return;
  }
  snackbar.error(t("cluster.details.deleteNotImplemented"));
}
</script>

