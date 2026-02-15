<template>
  <TableScaffoldPage
    titleKey="pages.labelsStages.title"
    hintKey="pages.labelsStages.hint"
    :headers="headers"
    :items="rows"
  >
    <template #below-header>
      <DismissibleWarningAlert
        alert-id="labels_stages"
        :title="t('warnings.labelsStages.title')"
        :text="t('warnings.labelsStages.text')"
      />
    </template>
  </TableScaffoldPage>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальный stage/label policy (OpenAPI контракт + store) и режимы редактирования с аудитом.
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import DismissibleWarningAlert from "../../shared/ui/DismissibleWarningAlert.vue";
import { useI18n } from "vue-i18n";

type PolicyRow = {
  kind: "label" | "stage";
  key: string;
  description: string;
  status: "active" | "planned";
};

const { t } = useI18n({ useScope: "global" });

const headers = [
  { key: "kind", width: 120 },
  { key: "key", width: 220 },
  { key: "description" },
  { key: "status", width: 140 },
  { key: "actions", sortable: false, width: 48 },
] as const;

const rows: PolicyRow[] = [
  { kind: "stage", key: "intake", description: "Issue intake and validation", status: "active" },
  { kind: "stage", key: "plan", description: "Work planning and decomposition", status: "active" },
  { kind: "stage", key: "impl", description: "Implementation (agent-run)", status: "active" },
  { kind: "stage", key: "review", description: "PR review / owner review", status: "active" },
  { kind: "stage", key: "ops", description: "Apply to cluster / smoke checks", status: "planned" },
  { kind: "label", key: "run:dev", description: "Run dev flow", status: "active" },
  { kind: "label", key: "need:owner_review", description: "Waiting for owner review", status: "active" },
  { kind: "label", key: "need:mcp_approval", description: "Waiting for MCP approval", status: "active" },
  { kind: "label", key: "state:blocked", description: "Execution blocked", status: "planned" },
];
</script>
