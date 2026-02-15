<template>
  <TableScaffoldPage
    titleKey="pages.cluster.pvc.title"
    hintKey="pages.cluster.pvc.hint"
    :headers="headers"
    :items="rows"
  >
    <template #below-header>
      <AdminClusterContextBar />
    </template>
    <template #row-actions="{ item }">
      <VBtn size="small" variant="text" :to="{ name: 'cluster-pvc-details', params: { name: String(item.name) } }">
        {{ t("scaffold.rowActions.view") }}
      </VBtn>
    </template>
  </TableScaffoldPage>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные PVC (list/get) и правила view-only/dry-run для окружений.
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";

type PvcRow = {
  name: string;
  namespace: string;
  status: string;
  volume: string;
  capacity: string;
  accessModes: string;
  storageClass: string;
  age: string;
};

const { t } = useI18n({ useScope: "global" });

const headers = [
  { title: "name", key: "name" },
  { title: "namespace", key: "namespace", width: 220 },
  { title: "status", key: "status", width: 140 },
  { title: "volume", key: "volume", width: 180 },
  { title: "capacity", key: "capacity", width: 120 },
  { title: "access", key: "accessModes", width: 120 },
  { title: "storageClass", key: "storageClass", width: 160 },
  { title: "age", key: "age", width: 120 },
  { title: "", key: "actions", sortable: false, width: 48 },
] as const;

const rows: PvcRow[] = [
  {
    name: "postgres-data",
    namespace: "codex-k8s-ai-staging",
    status: "Bound",
    volume: "pvc-7b2f...",
    capacity: "20Gi",
    accessModes: "RWO",
    storageClass: "standard",
    age: "10d",
  },
  {
    name: "pgvector-data",
    namespace: "codex-k8s-ai-staging",
    status: "Bound",
    volume: "pvc-3c19...",
    capacity: "50Gi",
    accessModes: "RWO",
    storageClass: "standard",
    age: "10d",
  },
];
</script>
