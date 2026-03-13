<template>
  <VCard variant="outlined">
    <VCardTitle class="text-subtitle-1 d-flex align-center ga-2">
      <VIcon icon="mdi-timeline-clock-outline" />
      {{ t("runs.timeline.title") }}
    </VCardTitle>
    <VCardText>
      <VTimeline density="compact" align="start" line-thickness="2">
        <VTimelineItem
          v-for="item in timelineItems"
          :key="item.key"
          :dot-color="item.color"
          :icon="item.icon"
          size="small"
        >
          <template v-if="item.showSpinner" #icon>
            <VProgressCircular indeterminate size="18" width="2" color="warning" />
          </template>
          <div class="d-flex align-center justify-space-between ga-4 flex-wrap">
            <div class="font-weight-bold">{{ titleForItem(item) }}</div>
            <VChip v-if="item.atLabel" size="x-small" variant="tonal" class="font-weight-bold">
              {{ item.atLabel }}
            </VChip>
          </div>
          <div v-if="subtitleForItem(item)" class="text-body-2 text-medium-emphasis mt-1">
            {{ subtitleForItem(item) }}
          </div>
        </VTimelineItem>
      </VTimeline>
    </VCardText>
  </VCard>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import type { TimelinePhaseStep, TimelinePhaseStepKey, TimelineStatusEntry } from "../lib/run-timeline";
import { buildRunTimelinePhases, buildRunTimelineStatuses } from "../lib/run-timeline";
import type { FlowEvent, Run } from "../../features/runs/types";

const props = defineProps<{
  run: Run | null;
  events: FlowEvent[];
  locale: string;
}>();

const { t } = useI18n({ useScope: "global" });

type TimelineDisplayItem = {
  key: string;
  at: string | null;
  atLabel: string | null;
  color: string;
  icon: string;
  showSpinner?: boolean;
  kind: "phase" | "agentStatus";
  phaseKey?: TimelinePhaseStepKey;
  subtitleKind?: TimelinePhaseStep["subtitleKind"];
  subtitleValue?: string;
  statusText?: string;
  repeatCount?: number;
};

function titleForStep(key: TimelinePhaseStepKey): string {
  switch (key) {
    case "buildDeploy":
      return t("runs.timeline.buildDeploy");
    case "started":
      return t("runs.timeline.started");
    case "authResolved":
      return t("runs.timeline.authResolved");
    case "agentReady":
      return t("runs.timeline.agentReady");
    case "waiting":
      return t("runs.timeline.waiting");
    case "finished":
      return t("runs.timeline.finished");
    case "created":
    default:
      return t("runs.timeline.created");
  }
}

function subtitleForStep(subtitleKind?: TimelinePhaseStep["subtitleKind"], subtitleValue?: string): string {
  switch (subtitleKind) {
    case "waitState":
      return `${t("runs.timeline.waitState")}: ${subtitleValue || "-"}`;
    case "status":
      return `${t("runs.timeline.status")}: ${subtitleValue || "-"}`;
    case "buildFailed":
      return t("runs.timeline.buildFailed");
    default:
      return "";
  }
}

const phaseSteps = computed(() => buildRunTimelinePhases(props.run, props.events, props.locale));
const statusEntries = computed(() =>
  buildRunTimelineStatuses(props.events, props.locale, (key, params) => String(t(key, params ?? {}))),
);

function titleForItem(item: TimelineDisplayItem): string {
  if (item.kind === "agentStatus") {
    return item.statusText || "";
  }
  return titleForStep(item.phaseKey || "created");
}

function subtitleForItem(item: TimelineDisplayItem): string {
  if (item.kind === "agentStatus" && Number(item.repeatCount || 0) > 1) {
    return `x${item.repeatCount}`;
  }
  return subtitleForStep(item.subtitleKind, item.subtitleValue);
}

function toTimelinePhaseItem(step: TimelinePhaseStep): TimelineDisplayItem {
  return {
    key: `phase:${step.key}:${step.at || "pending"}`,
    at: step.at,
    atLabel: step.atLabel,
    color: step.color,
    icon: step.icon,
    showSpinner: step.showSpinner,
    kind: "phase",
    phaseKey: step.key,
    subtitleKind: step.subtitleKind,
    subtitleValue: step.subtitleValue,
  };
}

function toTimelineStatusItem(status: TimelineStatusEntry): TimelineDisplayItem {
  return {
    key: status.key,
    at: status.at,
    atLabel: status.timeLabel,
    color: "primary",
    icon: "mdi-robot-outline",
    kind: "agentStatus",
    statusText: status.text,
    repeatCount: status.repeatCount,
  };
}

function compareTimelineItems(a: TimelineDisplayItem, b: TimelineDisplayItem): number {
  if (a.at && b.at) {
    if (a.at > b.at) return -1;
    if (a.at < b.at) return 1;
    return 0;
  }
  if (a.at && !b.at) return 1;
  if (!a.at && b.at) return -1;
  return 0;
}

const timelineItems = computed(() => {
  const phaseItems = phaseSteps.value.map(toTimelinePhaseItem);
  const statusItems = statusEntries.value.map(toTimelineStatusItem);
  return [...phaseItems, ...statusItems].sort(compareTimelineItems);
});
</script>

<style scoped>
:deep(.v-timeline-item) {
  min-height: 44px;
}
:deep(.v-timeline-item__body) {
  padding-bottom: 6px;
}
</style>
