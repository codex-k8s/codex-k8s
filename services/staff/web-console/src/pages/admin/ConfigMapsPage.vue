<template>
  <TableScaffoldPage
    titleKey="pages.cluster.configMaps.title"
    hintKey="pages.cluster.configMaps.hint"
    :headers="headers"
    :items="rows"
  >
    <template #below-header>
      <AdminClusterContextBar />
    </template>
    <template #row-actions="{ item }">
      <VBtn size="small" variant="text" :to="{ name: 'cluster-configmaps-details', params: { name: String(item.name) } }">
        {{ t("scaffold.rowActions.view") }}
      </VBtn>
    </template>
  </TableScaffoldPage>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные ConfigMaps (list/get/apply/delete), YAML view/edit на Monaco и action preview перед destructive действиями.
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";

type ConfigMapRow = {
  name: string;
  namespace: string;
  keys: number;
  age: string;
};

const { t } = useI18n({ useScope: "global" });

const headers = [
  { title: "name", key: "name" },
  { title: "namespace", key: "namespace", width: 220 },
  { title: "keys", key: "keys", width: 120, align: "end" },
  { title: "age", key: "age", width: 120 },
  { title: "", key: "actions", sortable: false, width: 48 },
] as const;

const rows: ConfigMapRow[] = [
  { name: "codex-k8s-web-console-config", namespace: "codex-k8s-ai-staging", keys: 3, age: "10d" },
  { name: "codex-k8s-api-gateway-config", namespace: "codex-k8s-ai-staging", keys: 5, age: "10d" },
  { name: "runner-env", namespace: "codex-k8s-dev-1", keys: 12, age: "2h" },
  { name: "feature-flags", namespace: "codex-k8s-dev-2", keys: 4, age: "1d" },
];
</script>
