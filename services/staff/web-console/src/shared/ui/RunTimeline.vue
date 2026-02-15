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
import type { Run } from "../../features/runs/types";

const props = defineProps<{
  run: Run | null;
  locale: string;
}>();

const { t } = useI18n({ useScope: "global" });

type Step = {
  key: string;
  title: string;
  at: string | null;
  atLabel: string;
  subtitle?: string;
  color: string;
  icon: string;
};

const steps = computed<Step[]>(() => {
  const r = props.run;
  if (!r) return [];

  const created: Step = {
    key: "created",
    title: t("runs.timeline.created"),
    at: r.created_at,
    atLabel: formatDateTime(r.created_at, props.locale),
    color: "info",
    icon: "mdi-calendar-plus",
  };

  const started: Step = {
    key: "started",
    title: t("runs.timeline.started"),
    at: r.started_at ?? null,
    atLabel: formatDateTime(r.started_at, props.locale),
    color: "primary",
    icon: "mdi-play",
  };

  const waiting: Step | null = r.wait_state
    ? {
        key: "waiting",
        title: t("runs.timeline.waiting"),
        at: r.wait_since ?? null,
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
    atLabel: formatDateTime(r.finished_at, props.locale),
    subtitle: `${t("runs.timeline.status")}: ${r.status}`,
    color: r.status === "succeeded" ? "success" : r.status === "failed" ? "error" : "secondary",
    icon: r.status === "succeeded" ? "mdi-check" : r.status === "failed" ? "mdi-alert-octagon-outline" : "mdi-flag-outline",
  };

  const out: Step[] = [created];
  if (r.started_at) out.push(started);
  if (waiting) out.push(waiting);
  out.push(finished);
  return out;
});
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
