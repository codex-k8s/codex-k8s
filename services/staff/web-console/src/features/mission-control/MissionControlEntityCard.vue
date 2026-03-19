<template>
  <button
    class="mission-node-card"
    :class="{
      'mission-node-card--selected': selected,
      'mission-node-card--dimmed': node.visibility_tier === 'secondary_dimmed',
      'mission-node-card--blocking': node.has_blocking_gap,
    }"
    type="button"
    @click="$emit('select')"
  >
    <div class="mission-node-card__topline">
      <VChip size="x-small" variant="tonal" :color="kindColor">
        {{ t(kindLabelKey) }}
      </VChip>
      <VChip size="x-small" variant="tonal" :color="stateColor">
        {{ t(stateLabelKey) }}
      </VChip>
    </div>

    <div class="mission-node-card__title">
      {{ node.title }}
    </div>

    <div class="mission-node-card__meta">
      <div class="mission-node-card__meta-row">
        <VIcon icon="mdi-graph-outline" size="16" />
        <span>{{ t(continuityLabelKey) }}</span>
      </div>
      <div class="mission-node-card__meta-row">
        <VIcon icon="mdi-eye-outline" size="16" />
        <span>{{ t(visibilityLabelKey) }}</span>
      </div>
      <div class="mission-node-card__meta-row">
        <VIcon icon="mdi-link-variant" size="16" />
        <span class="mono">{{ node.provider_reference?.external_id || node.node_public_id }}</span>
      </div>
      <div class="mission-node-card__meta-row">
        <VIcon icon="mdi-clock-outline" size="16" />
        <span class="mono">{{ formatCompactDateTime(node.last_activity_at, locale) }}</span>
      </div>
    </div>

    <div class="mission-node-card__footer">
      <VChip size="x-small" variant="outlined" :color="visibilityColor">
        {{ t(coverageLabelKey) }}
      </VChip>
      <VChip v-if="node.has_blocking_gap" size="x-small" variant="tonal" color="error">
        {{ t("pages.missionControl.blockingGap") }}
      </VChip>
    </div>

    <div v-if="node.badges.length" class="mission-node-card__badges">
      <VChip
        v-for="badge in node.badges"
        :key="badge"
        size="x-small"
        variant="outlined"
        class="mission-node-card__badge"
      >
        {{ t(missionControlBadgeLabelKey(badge)) }}
      </VChip>
    </div>
  </button>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import { formatCompactDateTime } from "../../shared/lib/datetime";
import {
  missionControlBadgeLabelKey,
  missionControlContinuityStatusLabelKey,
  missionControlCoverageClassLabelKey,
  missionControlNodeKindLabelKey,
  missionControlStateColor,
  missionControlStateLabelKey,
  missionControlVisibilityColor,
  missionControlVisibilityLabelKey,
} from "./presenters";
import type { MissionControlNode } from "./types";

const props = defineProps<{
  node: MissionControlNode;
  selected: boolean;
  locale: string;
}>();

defineEmits<{
  select: [];
}>();

const { t } = useI18n({ useScope: "global" });

const kindLabelKey = computed(() => missionControlNodeKindLabelKey(props.node.node_kind));
const stateLabelKey = computed(() => missionControlStateLabelKey(props.node.active_state));
const continuityLabelKey = computed(() => missionControlContinuityStatusLabelKey(props.node.continuity_status));
const visibilityLabelKey = computed(() => missionControlVisibilityLabelKey(props.node.visibility_tier));
const coverageLabelKey = computed(() => missionControlCoverageClassLabelKey(props.node.coverage_class));
const stateColor = computed(() => missionControlStateColor(props.node.active_state));
const visibilityColor = computed(() => missionControlVisibilityColor(props.node.visibility_tier));
const kindColor = computed(() => {
  switch (props.node.node_kind) {
    case "discussion":
      return "info";
    case "work_item":
      return "primary";
    case "run":
      return "warning";
    case "pull_request":
      return "success";
  }
});
</script>

<style scoped>
.mission-node-card {
  width: 100%;
  border: 1px solid rgba(15, 23, 42, 0.1);
  border-radius: 22px;
  padding: 16px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.96)),
    radial-gradient(240px 120px at 100% 0%, rgba(255, 219, 176, 0.32), transparent 65%);
  text-align: left;
  cursor: pointer;
  transition: transform 0.18s ease, box-shadow 0.18s ease, border-color 0.18s ease, opacity 0.18s ease;
}

.mission-node-card:hover {
  transform: translateY(-2px);
  border-color: rgba(15, 23, 42, 0.16);
  box-shadow: 0 18px 32px rgba(15, 23, 42, 0.1);
}

.mission-node-card--selected {
  border-color: rgba(14, 116, 144, 0.42);
  box-shadow: 0 0 0 2px rgba(14, 116, 144, 0.08), 0 18px 32px rgba(15, 23, 42, 0.1);
}

.mission-node-card--dimmed {
  opacity: 0.72;
}

.mission-node-card--blocking {
  border-color: rgba(220, 38, 38, 0.34);
}

.mission-node-card__topline {
  display: flex;
  gap: 8px;
  justify-content: space-between;
  flex-wrap: wrap;
}

.mission-node-card__title {
  margin-top: 12px;
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.35;
  color: rgb(15, 23, 42);
}

.mission-node-card__meta {
  display: grid;
  gap: 8px;
  margin-top: 14px;
  color: rgba(15, 23, 42, 0.72);
}

.mission-node-card__meta-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 0.92rem;
}

.mission-node-card__footer {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 14px;
}

.mission-node-card__badges {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 14px;
}

.mission-node-card__badge {
  border-color: rgba(15, 23, 42, 0.12);
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
