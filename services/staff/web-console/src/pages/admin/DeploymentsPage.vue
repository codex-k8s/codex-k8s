<template>
  <TableScaffoldPage
    titleKey="pages.cluster.deployments.title"
    hintKey="pages.cluster.deployments.hint"
    :headers="headers"
    :items="rows"
  >
    <template #below-header>
      <AdminClusterContextBar />
    </template>
    <template #row-actions="{ item }">
      <VBtn size="small" variant="text" :to="{ name: 'cluster-deployments-details', params: { name: String(item.name) } }">
        {{ t("scaffold.rowActions.view") }}
      </VBtn>
    </template>
  </TableScaffoldPage>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные Deployments (list/get), табы Overview/YAML/Events/Related, и guardrails:
// - platform deployments (app.kubernetes.io/part-of=codex-k8s) в ai-staging/prod = view-only
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";

type DeploymentRow = {
  name: string;
  namespace: string;
  ready: string;
  updated: string;
  age: string;
};

const { t } = useI18n({ useScope: "global" });

const headers = [
  { title: "name", key: "name" },
  { title: "namespace", key: "namespace", width: 220 },
  { title: "ready", key: "ready", width: 120 },
  { title: "updated", key: "updated", width: 160 },
  { title: "age", key: "age", width: 120 },
  { title: "", key: "actions", sortable: false, width: 48 },
] as const;

const rows: DeploymentRow[] = [
  { name: "codex-k8s", namespace: "codex-k8s-ai-staging", ready: "1/1", updated: "2026-02-15", age: "10d" },
  { name: "codex-k8s-worker", namespace: "codex-k8s-ai-staging", ready: "1/1", updated: "2026-02-15", age: "10d" },
  { name: "api-gateway", namespace: "codex-k8s-ai-staging", ready: "1/1", updated: "2026-02-15", age: "10d" },
  { name: "web-console", namespace: "codex-k8s-ai-staging", ready: "1/1", updated: "2026-02-15", age: "10d" },
];
</script>
