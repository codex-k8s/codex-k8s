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
            <VDataTable :headers="historyHeaders" :items="historyRows" :items-per-page="10" density="comfortable" />
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

const { t } = useI18n({ useScope: "global" });

const tab = ref<"catalog" | "approvals" | "history">("catalog");

const catalogHeaders = [
  { title: "tool", key: "tool", width: 220 },
  { title: "action", key: "action", width: 220 },
  { title: "scope", key: "scope", width: 160 },
  { title: "description", key: "description" },
] as const;
const catalogRows = [
  { tool: "github", action: "search.pull_requests", scope: "repo", description: "Search pull requests" },
  { tool: "kubernetes", action: "get.pods", scope: "cluster", description: "List pods in namespace" },
  { tool: "mcp", action: "approval.request", scope: "platform", description: "Request external approval" },
];

const approvalHeaders = [
  { title: "tool", key: "tool", width: 220 },
  { title: "action", key: "action", width: 220 },
  { title: "mode", key: "mode", width: 140 },
  { title: "approver", key: "approver" },
] as const;
const approvalRows = [
  { tool: "github", action: "create.pull_request", mode: "owner", approver: "Owner" },
  { tool: "kubernetes", action: "delete.namespace", mode: "delegated", approver: "Ops" },
  { tool: "mcp", action: "approval.request", mode: "owner", approver: "Owner" },
];

const historyHeaders = [
  { title: "created_at", key: "created_at", width: 160 },
  { title: "tool", key: "tool", width: 180 },
  { title: "action", key: "action", width: 220 },
  { title: "outcome", key: "outcome", width: 140 },
  { title: "correlation_id", key: "correlation_id" },
] as const;
const historyRows = [
  { created_at: "2026-02-15T10:12:09Z", tool: "github", action: "search.pull_requests", outcome: "ok", correlation_id: "c0d3x-9a3d" },
  { created_at: "2026-02-15T10:05:31Z", tool: "kubernetes", action: "get.pods", outcome: "ok", correlation_id: "c0d3x-1f21" },
  { created_at: "2026-02-15T09:49:44Z", tool: "mcp", action: "approval.request", outcome: "pending", correlation_id: "c0d3x-5d3c" },
];
</script>

