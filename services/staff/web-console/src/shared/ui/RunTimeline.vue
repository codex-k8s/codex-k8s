<template>
  <VCard variant="outlined">
    <VCardTitle class="text-subtitle-1 d-flex align-center ga-2">
      <VIcon icon="mdi-timeline-clock-outline" />
      {{ t("runs.timeline.title") }}
    </VCardTitle>
    <VCardText>
      <VTimeline density="compact" align="start" line-thickness="2">
        <VTimelineItem
          v-for="s in steps"
          :key="s.key"
          :dot-color="s.color"
          :icon="s.icon"
          size="small"
        >
          <div class="d-flex align-center justify-space-between ga-4 flex-wrap">
            <div class="font-weight-bold">{{ s.title }}</div>
            <div class="text-body-2 text-medium-emphasis mono">{{ s.atLabel }}</div>
          </div>
          <div v-if="s.subtitle" class="text-body-2 text-medium-emphasis mt-1">
            {{ s.subtitle }}
          </div>
        </VTimelineItem>
      </VTimeline>
    </VCardText>
  </VCard>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import { formatDateTime } from "../lib/datetime";
import type { FlowEvent, Run } from "../../features/runs/types";

const props = defineProps<{
  run: Run | null;
  events: FlowEvent[];
  locale: string;
}>();

const { t } = useI18n({ useScope: "global" });

type Step = {
  key: string;
  title: string;
  at: string | null;
  orderAt: number;
  atLabel: string;
  subtitle?: string;
  color: string;
  icon: string;
};

type AgentStatusPayload = {
  statusText: string;
  agentKey: string;
};

function parseAgentStatusPayload(raw: string): AgentStatusPayload {
  const value = String(raw || "").trim();
  if (!value) {
    return { statusText: "", agentKey: "" };
  }
  try {
    const parsed = JSON.parse(value) as unknown;
    const normalized = typeof parsed === "string" ? JSON.parse(parsed) as unknown : parsed;
    if (!normalized || typeof normalized !== "object") {
      return { statusText: "", agentKey: "" };
    }
    const record = normalized as Record<string, unknown>;
    return {
      statusText: String(record.status_text || "").trim(),
      agentKey: String(record.agent_key || "").trim(),
    };
  } catch {
    return { statusText: "", agentKey: "" };
  }
}

function dateOrder(value: string | null | undefined): number {
  if (!value) return 0;
  const parsed = Date.parse(value);
  if (Number.isNaN(parsed)) return 0;
  return parsed;
}

const steps = computed<Step[]>(() => {
  const r = props.run;
  if (!r) return [];

  const created: Step = {
    key: "created",
    title: t("runs.timeline.created"),
    at: r.created_at,
    orderAt: dateOrder(r.created_at),
    atLabel: formatDateTime(r.created_at, props.locale),
    color: "info",
    icon: "mdi-calendar-plus",
  };

  const started: Step = {
    key: "started",
    title: t("runs.timeline.started"),
    at: r.started_at ?? null,
    orderAt: dateOrder(r.started_at),
    atLabel: formatDateTime(r.started_at, props.locale),
    color: "primary",
    icon: "mdi-play",
  };

  const waiting: Step | null = r.wait_state
      ? {
        key: "waiting",
        title: t("runs.timeline.waiting"),
        at: r.wait_since ?? null,
        orderAt: dateOrder(r.wait_since),
        atLabel: formatDateTime(r.wait_since, props.locale),
        subtitle: `${t("runs.timeline.waitState")}: ${r.wait_state}`,
        color: "warning",
        icon: "mdi-timer-sand",
      }
    : null;

  const finished: Step = {
    key: "finished",
    title: t("runs.timeline.finished"),
    at: r.finished_at ?? null,
    orderAt: dateOrder(r.finished_at),
    atLabel: formatDateTime(r.finished_at, props.locale),
    subtitle: `${t("runs.timeline.status")}: ${r.status}`,
    color: r.status === "succeeded" ? "success" : r.status === "failed" ? "error" : "secondary",
    icon: r.status === "succeeded" ? "mdi-check" : r.status === "failed" ? "mdi-alert-octagon-outline" : "mdi-flag-outline",
  };

  const out: Step[] = [created];
  if (r.started_at) out.push(started);
  if (waiting) out.push(waiting);
  if (r.finished_at) out.push(finished);

  const statusSteps: Step[] = [];
  for (const eventItem of props.events) {
    if (eventItem.event_type !== "run.agent.status_reported") continue;
    const payload = parseAgentStatusPayload(eventItem.payload_json || "");
    if (!payload.statusText) continue;
    statusSteps.push({
      key: `agent-status:${eventItem.created_at}:${payload.statusText}`,
      title: payload.statusText,
      at: eventItem.created_at,
      orderAt: dateOrder(eventItem.created_at),
      atLabel: formatDateTime(eventItem.created_at, props.locale),
      subtitle: payload.agentKey ? `${t("runs.timeline.agent")}: ${payload.agentKey}` : undefined,
      color: "secondary",
      icon: "mdi-robot-outline",
    });
  }

  out.push(...statusSteps);
  out.sort((a, b) => {
    if (a.orderAt === b.orderAt) return a.key < b.key ? 1 : -1;
    return b.orderAt - a.orderAt;
  });
  return out;
});
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
:deep(.v-timeline-item) {
  min-height: 44px;
}
:deep(.v-timeline-item__body) {
  padding-bottom: 6px;
}
</style>
