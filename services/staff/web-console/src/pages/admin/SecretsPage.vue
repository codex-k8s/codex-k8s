<template>
  <TableScaffoldPage
    titleKey="pages.cluster.secrets.title"
    hintKey="pages.cluster.secrets.hint"
    :headers="headers"
    :items="rows"
  >
    <template #below-header>
      <AdminClusterContextBar />
    </template>
    <template #row-actions="{ item }">
      <VBtn size="small" variant="text" :to="{ name: 'cluster-secrets-details', params: { name: String(item.name) } }">
        {{ t("scaffold.rowActions.view") }}
      </VBtn>
    </template>
  </TableScaffoldPage>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные Secrets (metadata-only по умолчанию), reveal как отдельное осознанное действие, YAML view/edit на Monaco.
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";

type SecretRow = {
  name: string;
  namespace: string;
  type: string;
  keys: number;
  age: string;
};

const { t } = useI18n({ useScope: "global" });

const headers = [
  { title: "name", key: "name" },
  { title: "namespace", key: "namespace", width: 220 },
  { title: "type", key: "type", width: 180 },
  { title: "keys", key: "keys", width: 120, align: "end" },
  { title: "age", key: "age", width: 120 },
  { title: "", key: "actions", sortable: false, width: 48 },
] as const;

const rows: SecretRow[] = [
  { name: "postgres-credentials", namespace: "codex-k8s-ai-staging", type: "Opaque", keys: 2, age: "10d" },
  { name: "github-oauth", namespace: "codex-k8s-ai-staging", type: "Opaque", keys: 2, age: "10d" },
  { name: "runner-token", namespace: "codex-k8s-dev-1", type: "Opaque", keys: 1, age: "2h" },
  { name: "tls-cert", namespace: "codex-k8s-prod", type: "kubernetes.io/tls", keys: 2, age: "45d" },
];
</script>
