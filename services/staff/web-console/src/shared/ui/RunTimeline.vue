<template>
  <VCard variant="outlined">
    <VCardTitle class="text-subtitle-1 d-flex align-center ga-2">
      <VIcon icon="mdi-timeline-clock-outline" />
      {{ t("runs.timeline.title") }}
    </VCardTitle>
    <VCardText>
      <VTimeline density="compact" align="start" line-thickness="2">
        <VTimelineItem
          v-for="s in phaseSteps"
          :key="s.key"
          :dot-color="s.color"
          :icon="s.icon"
          size="small"
        >
          <template v-if="s.showSpinner" #icon>
            <VProgressCircular indeterminate size="18" width="2" color="warning" />
          </template>
          <div class="d-flex align-center justify-space-between ga-4 flex-wrap">
            <div class="font-weight-bold">{{ titleForStep(s.key) }}</div>
            <VChip size="x-small" variant="tonal" class="font-weight-bold">
              {{ s.atLabel || t("runs.timeline.inProgress") }}
            </VChip>
          </div>
          <div v-if="subtitleForStep(s)" class="text-body-2 text-medium-emphasis mt-1">
            {{ subtitleForStep(s) }}
          </div>
        </VTimelineItem>
      </VTimeline>

      <template v-if="statusEntries.length">
        <VDivider class="my-4" />
        <div class="text-subtitle-2 font-weight-bold mb-3">{{ t("runs.timeline.statusUpdates") }}</div>
        <div class="d-flex flex-column ga-2">
          <div
            v-for="entry in statusEntries"
            :key="entry.key"
            class="d-flex align-center justify-space-between ga-3 flex-wrap"
          >
            <div class="text-body-2">{{ entry.text }}</div>
            <div class="d-flex align-center ga-2 flex-wrap">
              <VChip v-if="entry.repeatCount > 1" size="x-small" variant="outlined">x{{ entry.repeatCount }}</VChip>
              <VChip size="x-small" variant="tonal" class="font-weight-bold">{{ entry.timeLabel }}</VChip>
            </div>
          </div>
        </div>
      </template>
    </VCardText>
  </VCard>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import type { TimelinePhaseStep, TimelinePhaseStepKey } from "../lib/run-timeline";
import { buildRunTimelinePhases, buildRunTimelineStatuses } from "../lib/run-timeline";
import type { FlowEvent, Run } from "../../features/runs/types";

const props = defineProps<{
  run: Run | null;
  events: FlowEvent[];
  locale: string;
}>();

const { t } = useI18n({ useScope: "global" });

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

function subtitleForStep(step: TimelinePhaseStep): string {
  switch (step.subtitleKind) {
    case "waitState":
      return `${t("runs.timeline.waitState")}: ${step.subtitleValue || "-"}`;
    case "status":
      return `${t("runs.timeline.status")}: ${step.subtitleValue || "-"}`;
    case "buildFailed":
      return t("runs.timeline.buildFailed");
    default:
      return "";
  }
}

const phaseSteps = computed(() => buildRunTimelinePhases(props.run, props.events, props.locale));
const statusEntries = computed(() => buildRunTimelineStatuses(props.events, props.locale));
</script>

<style scoped>
:deep(.v-timeline-item) {
  min-height: 44px;
}
:deep(.v-timeline-item__body) {
  padding-bottom: 6px;
}
</style>
