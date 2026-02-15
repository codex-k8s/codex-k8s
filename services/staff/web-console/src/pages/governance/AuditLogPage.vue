<template>
  <TableScaffoldPage
    titleKey="pages.auditLog.title"
    hintKey="pages.auditLog.hint"
    :headers="headers"
    :items="rows"
  />
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальный audit log (endpoint + модель данных), фильтры (actor/object/env/correlation_id) и пагинацию.
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";

type AuditRow = {
  created_at: string;
  actor: string;
  action: string;
  object: string;
  env: string;
  correlation_id: string;
};

const headers = [
  { title: "created_at", key: "created_at", width: 140 },
  { title: "actor", key: "actor", width: 180 },
  { title: "action", key: "action", width: 160 },
  { title: "object", key: "object" },
  { title: "env", key: "env", width: 140 },
  { title: "correlation_id", key: "correlation_id" },
  { title: "", key: "actions", sortable: false, width: 48 },
] as const;

const rows: AuditRow[] = [
  {
    created_at: "2026-02-15T10:12:09Z",
    actor: "owner@codex-k8s.local",
    action: "approval.resolve",
    object: "approval_request#418",
    env: "ai",
    correlation_id: "c0d3x-9a3d",
  },
  {
    created_at: "2026-02-15T10:05:31Z",
    actor: "system",
    action: "run.created",
    object: "run#4d2a",
    env: "ai",
    correlation_id: "c0d3x-1f21",
  },
  {
    created_at: "2026-02-15T09:58:02Z",
    actor: "admin@codex-k8s.local",
    action: "project.upsert",
    object: "project#codex-k8s",
    env: "ai-staging",
    correlation_id: "c0d3x-7b11",
  },
  {
    created_at: "2026-02-15T09:52:18Z",
    actor: "system",
    action: "mcp.call",
    object: "tool=github.search/action=pull_requests",
    env: "ai",
    correlation_id: "c0d3x-5d3c",
  },
  {
    created_at: "2026-02-15T09:49:44Z",
    actor: "developer@codex-k8s.local",
    action: "run.wait",
    object: "wait_state=mcp",
    env: "ai",
    correlation_id: "c0d3x-5d3c",
  },
];
</script>

