<template>
  <TableScaffoldPage
    titleKey="pages.cluster.namespaces.title"
    hintKey="pages.cluster.namespaces.hint"
    :headers="headers"
    :items="rows"
  >
    <template #below-header>
      <AdminClusterContextBar />
    </template>
    <template #row-actions="{ item }">
      <VBtn size="small" variant="text" :to="{ name: 'cluster-namespaces-details', params: { name: String(item.name) } }">
        {{ t("scaffold.rowActions.view") }}
      </VBtn>
    </template>
  </TableScaffoldPage>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные namespaces из backend (k8s API + RBAC + audit) и правила режимов:
// - ai-staging/prod для platform resources (app.kubernetes.io/part-of=codex-k8s) = view-only
// - ai env = destructive actions через backend dry-run, с явным feedback
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";

type NamespaceRow = {
  name: string;
  status: "Active" | "Terminating";
  age: string;
  part_of: "codex-k8s" | "-";
};

const { t } = useI18n({ useScope: "global" });

const headers = [
  { title: "name", key: "name" },
  { title: "status", key: "status", width: 160 },
  { title: "age", key: "age", width: 120 },
  { title: "part-of", key: "part_of", width: 140 },
  { title: "", key: "actions", sortable: false, width: 48 },
] as const;

const rows: NamespaceRow[] = [
  { name: "codex-k8s-dev-1", status: "Active", age: "3h", part_of: "-" },
  { name: "codex-k8s-dev-2", status: "Active", age: "1d", part_of: "-" },
  { name: "codex-k8s-ai-staging", status: "Active", age: "12d", part_of: "codex-k8s" },
  { name: "codex-k8s-prod", status: "Active", age: "45d", part_of: "codex-k8s" },
];
</script>
