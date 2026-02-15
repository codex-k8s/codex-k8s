<template>
  <TableScaffoldPage
    titleKey="pages.cluster.jobs.title"
    hintKey="pages.cluster.jobs.hint"
    :headers="headers"
    :items="rows"
  >
    <template #below-header>
      <AdminClusterContextBar />
    </template>
    <template #row-actions="{ item }">
      <VBtn size="small" variant="text" :to="{ name: 'cluster-jobs-details', params: { name: String(item.name) } }">
        {{ t("scaffold.rowActions.view") }}
      </VBtn>
    </template>
  </TableScaffoldPage>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные Jobs + логи контейнеров (logs viewer), табы Overview/YAML/Events/Related/Logs и action preview.
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";

type JobRow = {
  name: string;
  namespace: string;
  completions: string;
  duration: string;
  age: string;
};

const { t } = useI18n({ useScope: "global" });

const headers = [
  { title: "name", key: "name" },
  { title: "namespace", key: "namespace", width: 220 },
  { title: "completions", key: "completions", width: 140 },
  { title: "duration", key: "duration", width: 140 },
  { title: "age", key: "age", width: 120 },
  { title: "", key: "actions", sortable: false, width: 48 },
] as const;

const rows: JobRow[] = [
  { name: "agent-runner-27184120", namespace: "codex-k8s-dev-1", completions: "0/1", duration: "15m", age: "15m" },
  { name: "db-migrate-27184001", namespace: "codex-k8s-ai-staging", completions: "1/1", duration: "32s", age: "10d" },
  { name: "smoke-check-27183012", namespace: "codex-k8s-ai-staging", completions: "1/1", duration: "58s", age: "10d" },
];
</script>
