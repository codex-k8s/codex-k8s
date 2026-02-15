<template>
  <TableScaffoldPage titleKey="pages.agents.title" hintKey="pages.agents.hint" :headers="headers" :items="rows">
    <template #row-actions="{ item }">
      <VBtn size="small" variant="text" :to="{ name: 'agent-details', params: { agentName: String(item.name) } }">
        {{ t("scaffold.rowActions.view") }}
      </VBtn>
    </template>
  </TableScaffoldPage>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные агенты (list/details), режимы (full-env/code-only), лимиты, статус, а также tabs:
// Settings / Prompt templates / History-Audit.
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";

type AgentRow = {
  name: string;
  kind: "system" | "custom";
  mode: "full-env" | "code-only";
  limits: string;
  status: "ready" | "disabled";
};

const { t } = useI18n({ useScope: "global" });

const headers = [
  { title: "name", key: "name" },
  { title: "kind", key: "kind", width: 140 },
  { title: "mode", key: "mode", width: 160 },
  { title: "limits", key: "limits", width: 200 },
  { title: "status", key: "status", width: 140 },
  { title: "", key: "actions", sortable: false, width: 48 },
] as const;

const rows: AgentRow[] = [
  { name: "developer", kind: "system", mode: "full-env", limits: "cpu=2, mem=4Gi", status: "ready" },
  { name: "reviewer", kind: "system", mode: "code-only", limits: "cpu=1, mem=2Gi", status: "ready" },
  { name: "ops", kind: "system", mode: "full-env", limits: "cpu=2, mem=4Gi", status: "disabled" },
  { name: "custom-ui", kind: "custom", mode: "code-only", limits: "cpu=1, mem=2Gi", status: "ready" },
];
</script>
