<template>
  <div>
    <PageHeader :title="t('pages.mcpTools.title')" :hint="t('pages.mcpTools.hint')" />

    <VTabs v-model="tab" class="mt-4">
      <VTab value="catalog">{{ t("pages.mcpTools.tabs.catalog") }}</VTab>
      <VTab value="approvals">{{ t("pages.mcpTools.tabs.approvals") }}</VTab>
      <VTab value="history">{{ t("pages.mcpTools.tabs.history") }}</VTab>
    </VTabs>

    <VWindow v-model="tab" class="mt-2">
      <VWindowItem value="catalog">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-2">{{ t("pages.mcpTools.catalogTitle") }}</VCardTitle>
          <VCardText>
            <VDataTable :headers="catalogHeaders" :items="catalogRows" :items-per-page="10" density="comfortable" />
          </VCardText>
        </VCard>
      </VWindowItem>

      <VWindowItem value="approvals">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-2">{{ t("pages.mcpTools.approvalsTitle") }}</VCardTitle>
          <VCardText>
            <VDataTable :headers="approvalHeaders" :items="approvalRows" :items-per-page="10" density="comfortable" />
          </VCardText>
        </VCard>
      </VWindowItem>

      <VWindowItem value="history">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-2">{{ t("pages.mcpTools.historyTitle") }}</VCardTitle>
          <VCardText>
            <VDataTable :headers="historyHeaders" :items="historyRows" :items-per-page="10" density="comfortable">
              <template #item.created_at="{ item }">
                <span class="text-medium-emphasis">{{ formatDateTime(String(item.created_at || ''), locale) }}</span>
              </template>
              <template #item.correlation_id="{ item }">
                <span class="mono text-medium-emphasis">{{ item.correlation_id }}</span>
              </template>
            </VDataTable>
          </VCardText>
        </VCard>
      </VWindowItem>
    </VWindow>
  </div>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные MCP инструменты/апрувы/историю вызовов из backend и связать с audit log/correlation_id.
import { ref } from "vue";
import { useI18n } from "vue-i18n";

import PageHeader from "../../shared/ui/PageHeader.vue";
import { formatDateTime } from "../../shared/lib/datetime";

const { t, locale } = useI18n({ useScope: "global" });

const tab = ref<"catalog" | "approvals" | "history">("catalog");

const catalogHeaders = [
  { title: t("table.fields.tool"), key: "tool", width: 220, align: "start" },
  { title: t("table.fields.action"), key: "action", width: 220, align: "center" },
  { title: t("table.fields.scope"), key: "scope", width: 160, align: "center" },
  { title: t("table.fields.description"), key: "description", align: "center" },
] as const;
const catalogRows = [
  { tool: "github", action: "search.pull_requests", scope: "repo", description: "Search pull requests" },
  { tool: "kubernetes", action: "get.pods", scope: "cluster", description: "List pods in namespace" },
  { tool: "mcp", action: "approval.request", scope: "platform", description: "Request external approval" },
];

const approvalHeaders = [
  { title: t("table.fields.tool"), key: "tool", width: 220, align: "start" },
  { title: t("table.fields.action"), key: "action", width: 220, align: "center" },
  { title: t("table.fields.mode"), key: "mode", width: 140, align: "center" },
  { title: t("table.fields.approver"), key: "approver", align: "center" },
] as const;
const approvalRows = [
  { tool: "github", action: "create.pull_request", mode: "owner", approver: "Owner" },
  { tool: "kubernetes", action: "delete.namespace", mode: "delegated", approver: "Ops" },
  { tool: "mcp", action: "approval.request", mode: "owner", approver: "Owner" },
];

const historyHeaders = [
  { title: t("table.fields.created_at"), key: "created_at", width: 160, align: "start" },
  { title: t("table.fields.tool"), key: "tool", width: 180, align: "center" },
  { title: t("table.fields.action"), key: "action", width: 220, align: "center" },
  { title: t("table.fields.outcome"), key: "outcome", width: 140, align: "center" },
  { title: t("table.fields.correlation_id"), key: "correlation_id", align: "center" },
] as const;
const historyRows = [
  { created_at: "2026-02-15T10:12:09Z", tool: "github", action: "search.pull_requests", outcome: "ok", correlation_id: "c0d3x-9a3d" },
  { created_at: "2026-02-15T10:05:31Z", tool: "kubernetes", action: "get.pods", outcome: "ok", correlation_id: "c0d3x-1f21" },
  { created_at: "2026-02-15T09:49:44Z", tool: "mcp", action: "approval.request", outcome: "pending", correlation_id: "c0d3x-5d3c" },
];
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
