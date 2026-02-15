<template>
  <TableScaffoldPage
    titleKey="pages.cluster.pods.title"
    hintKey="pages.cluster.pods.hint"
    :headers="headers"
    :items="rows"
  >
    <template #below-header>
      <AdminClusterContextBar />
    </template>
    <template #row-actions="{ item }">
      <VBtn size="small" variant="text" :to="{ name: 'cluster-pods-details', params: { name: String(item.name) } }">
        {{ t("scaffold.rowActions.view") }}
      </VBtn>
    </template>
  </TableScaffoldPage>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные Pods + отдельный экран/таб Logs (Monaco не нужен; для логов использовать logs viewer),
// а также action preview перед destructive действиями.
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";

type PodRow = {
  name: string;
  namespace: string;
  ready: string;
  status: string;
  restarts: number;
  age: string;
};

const { t } = useI18n({ useScope: "global" });

const headers = [
  { title: "name", key: "name" },
  { title: "namespace", key: "namespace", width: 220 },
  { title: "ready", key: "ready", width: 120 },
  { title: "status", key: "status", width: 160 },
  { title: "restarts", key: "restarts", width: 120, align: "end" },
  { title: "age", key: "age", width: 120 },
  { title: "", key: "actions", sortable: false, width: 48 },
] as const;

const rows: PodRow[] = [
  { name: "codex-k8s-6b7c79d7b9-2rj6q", namespace: "codex-k8s-ai-staging", ready: "1/1", status: "Running", restarts: 0, age: "2h" },
  { name: "api-gateway-7f7cc8dbb7-7d2s9", namespace: "codex-k8s-ai-staging", ready: "1/1", status: "Running", restarts: 1, age: "2h" },
  { name: "web-console-6d6fd8c8d6-h9v2p", namespace: "codex-k8s-ai-staging", ready: "1/1", status: "Running", restarts: 0, age: "2h" },
  { name: "codex-k8s-worker-6bfbd7b8f9-6k1z2", namespace: "codex-k8s-ai-staging", ready: "1/1", status: "Running", restarts: 0, age: "2h" },
  { name: "agent-runner-27184120-zzr8m", namespace: "codex-k8s-dev-1", ready: "1/1", status: "Running", restarts: 0, age: "15m" },
];
</script>
