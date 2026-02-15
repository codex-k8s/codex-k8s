<template>
  <TableScaffoldPage titleKey="pages.agents.title" hintKey="pages.agents.hint" :headers="headers" :items="rows">
    <template #item.name="{ item }">
      <RouterLink
        class="text-primary font-weight-bold text-decoration-none"
        :to="{ name: 'agent-details', params: { agentName: String(item.name) } }"
      >
        {{ item.name }}
      </RouterLink>
    </template>
    <template #row-actions="{ item }">
      <VTooltip :text="t('scaffold.rowActions.view')">
        <template #activator="{ props: tipProps }">
          <VBtn
            v-bind="tipProps"
            size="small"
            variant="text"
            icon="mdi-open-in-new"
            :to="{ name: 'agent-details', params: { agentName: String(item.name) } }"
          />
        </template>
      </VTooltip>
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
  { key: "name" },
  { key: "kind", width: 140 },
  { key: "mode", width: 160 },
  { key: "limits", width: 200 },
  { key: "status", width: 140 },
  { key: "actions", sortable: false, width: 48 },
] as const;

const rows: AgentRow[] = [
  { name: "developer", kind: "system", mode: "full-env", limits: "cpu=2, mem=4Gi", status: "ready" },
  { name: "reviewer", kind: "system", mode: "code-only", limits: "cpu=1, mem=2Gi", status: "ready" },
  { name: "ops", kind: "system", mode: "full-env", limits: "cpu=2, mem=4Gi", status: "disabled" },
  { name: "custom-ui", kind: "custom", mode: "code-only", limits: "cpu=1, mem=2Gi", status: "ready" },
];
</script>
