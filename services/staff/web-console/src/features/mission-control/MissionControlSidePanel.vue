<template>
  <VCard class="mission-side-panel" variant="outlined">
    <template v-if="loading">
      <VCardText class="py-8">
        <VSkeletonLoader type="article, paragraph, paragraph" />
      </VCardText>
    </template>

    <template v-else-if="error && !details">
      <VCardText>
        <VAlert type="error" variant="tonal">
          {{ t(error.messageKey) }}
        </VAlert>
      </VCardText>
    </template>

    <template v-else-if="details">
      <VCardItem>
        <template #prepend>
          <VAvatar color="primary" variant="tonal" size="42">
            <VIcon :icon="panelIcon" />
          </VAvatar>
        </template>

        <VCardTitle class="mission-side-panel__title">
          {{ details.node.title }}
        </VCardTitle>
        <VCardSubtitle class="mission-side-panel__subtitle">
          {{ t(kindLabelKey) }}
        </VCardSubtitle>

        <template #append>
          <VBtn icon="mdi-close" variant="text" @click="$emit('close')" />
        </template>
      </VCardItem>

      <VCardText class="mission-side-panel__content">
        <div class="d-flex flex-wrap ga-2">
          <VChip size="small" variant="tonal" :color="stateColor">
            {{ t(stateLabelKey) }}
          </VChip>
          <VChip size="small" variant="tonal" :color="continuityColor">
            {{ t(continuityLabelKey) }}
          </VChip>
          <VChip size="small" variant="tonal" :color="visibilityColor">
            {{ t(visibilityLabelKey) }}
          </VChip>
        </div>

        <VCard class="mission-side-panel__section" color="surface" variant="tonal">
          <VCardTitle class="text-subtitle-2">{{ t("pages.missionControl.sidePanel.overview") }}</VCardTitle>
          <VCardText class="mission-side-panel__rows">
            <div v-for="row in infoRows" :key="row.labelKey + ':' + row.value" class="mission-side-panel__row">
              <div class="mission-side-panel__label">{{ t(row.labelKey) }}</div>
              <a
                v-if="row.href"
                class="mission-side-panel__value mission-side-panel__value--link"
                :class="{ mono: row.mono }"
                :href="row.href"
                target="_blank"
                rel="noreferrer"
              >
                {{ row.value }}
              </a>
              <div v-else class="mission-side-panel__value" :class="{ mono: row.mono }">
                {{ formatCellValue(row.value) }}
              </div>
            </div>

            <div v-if="isDiscussionPayload(details.detail_payload) && details.detail_payload.latest_comment_excerpt" class="mission-side-panel__row">
              <div class="mission-side-panel__label">{{ t("pages.missionControl.sidePanel.latestComment") }}</div>
              <div class="mission-side-panel__value">
                {{ details.detail_payload.latest_comment_excerpt }}
              </div>
            </div>

            <div v-if="isWorkItemPayload(details.detail_payload) && details.detail_payload.labels.length" class="mission-side-panel__row">
              <div class="mission-side-panel__label">{{ t("pages.missionControl.sidePanel.labels") }}</div>
              <div class="d-flex flex-wrap ga-2">
                <VChip v-for="label in details.detail_payload.labels" :key="label" size="x-small" variant="outlined">
                  {{ label }}
                </VChip>
              </div>
            </div>

            <div v-if="isWorkItemPayload(details.detail_payload) && details.detail_payload.assignees.length" class="mission-side-panel__row">
              <div class="mission-side-panel__label">{{ t("pages.missionControl.sidePanel.assignees") }}</div>
              <div class="d-flex flex-wrap ga-2">
                <VChip v-for="assignee in details.detail_payload.assignees" :key="assignee" size="x-small" variant="outlined">
                  {{ assignee }}
                </VChip>
              </div>
            </div>
          </VCardText>
        </VCard>

        <VCard class="mission-side-panel__section" color="surface" variant="tonal">
          <VCardTitle class="text-subtitle-2">{{ t("pages.missionControl.sidePanel.continuity") }}</VCardTitle>
          <VCardText>
            <div v-if="details.continuity_gaps.length" class="mission-side-panel__gaps">
              <VSheet
                v-for="gap in details.continuity_gaps"
                :key="gap.gap_id"
                class="mission-side-panel__gap"
                rounded="xl"
                border
              >
                <div class="d-flex align-start ga-3">
                  <VAvatar :color="missionControlGapSeverityColor(gap.severity)" size="34" variant="tonal">
                    <VIcon icon="mdi-alert-circle-outline" size="18" />
                  </VAvatar>
                  <div class="flex-grow-1">
                    <div class="font-weight-medium">{{ t(missionControlGapKindLabelKey(gap.gap_kind)) }}</div>
                    <div class="text-body-2 text-medium-emphasis mt-1">
                      {{ t(missionControlGapSeverityLabelKey(gap.severity)) }} · {{ gap.status }}
                    </div>
                    <div v-if="gap.resolution_hint" class="text-body-2 mt-2">
                      {{ gap.resolution_hint }}
                    </div>
                    <div class="text-caption text-medium-emphasis mt-2 mono">
                      {{ formatCompactDateTime(gap.detected_at, locale) }}
                    </div>
                  </div>
                </div>
              </VSheet>
            </div>
            <div v-else class="text-medium-emphasis">
              {{ t("pages.missionControl.sidePanel.noContinuityGaps") }}
            </div>
          </VCardText>
        </VCard>

        <VCard class="mission-side-panel__section" color="surface" variant="tonal">
          <VCardTitle class="text-subtitle-2">{{ t("pages.missionControl.sidePanel.watermarks") }}</VCardTitle>
          <VCardText>
            <div v-if="details.node_watermarks.length" class="d-flex flex-wrap ga-2">
              <VChip
                v-for="watermark in details.node_watermarks"
                :key="watermark.watermark_kind + ':' + watermark.observed_at"
                size="small"
                variant="tonal"
                :color="missionControlWatermarkColor(watermark.status)"
              >
                {{ t(missionControlWatermarkLabelKey(watermark.watermark_kind)) }} · {{ watermark.summary }}
              </VChip>
            </div>
            <div v-else class="text-medium-emphasis">
              {{ t("pages.missionControl.sidePanel.noWatermarks") }}
            </div>
          </VCardText>
        </VCard>

        <VCard class="mission-side-panel__section" color="surface" variant="tonal">
          <VCardTitle class="text-subtitle-2">{{ t("pages.missionControl.sidePanel.linkedArtifacts") }}</VCardTitle>
          <VCardText class="mission-side-panel__linked">
            <template v-if="relatedSections.length || adjacentRefs.length">
              <div
                v-for="section in relatedSections"
                :key="section.titleKey"
                class="mission-side-panel__linked-section"
              >
                <div class="mission-side-panel__label">{{ t(section.titleKey) }}</div>
                <div class="d-flex flex-wrap ga-2">
                  <VChip
                    v-for="ref in section.refs"
                    :key="ref.node_kind + ':' + ref.node_public_id"
                    size="small"
                    variant="outlined"
                    @click="$emit('selectNode', ref)"
                  >
                    {{ t(missionControlNodeKindLabelKey(ref.node_kind)) }} · {{ ref.node_public_id }}
                  </VChip>
                </div>
              </div>

              <div v-if="adjacentRefs.length" class="mission-side-panel__linked-section">
                <div class="mission-side-panel__label">{{ t("pages.missionControl.sidePanel.adjacentNodes") }}</div>
                <div class="d-flex flex-wrap ga-2">
                  <VChip
                    v-for="ref in adjacentRefs"
                    :key="ref.node_kind + ':' + ref.node_public_id"
                    size="small"
                    variant="outlined"
                    @click="$emit('selectNode', ref)"
                  >
                    {{ t(missionControlNodeKindLabelKey(ref.node_kind)) }} · {{ ref.node_public_id }}
                  </VChip>
                </div>
              </div>
            </template>

            <div v-else class="text-medium-emphasis">
              {{ t("pages.missionControl.sidePanel.noLinkedArtifacts") }}
            </div>
          </VCardText>
        </VCard>

        <VCard class="mission-side-panel__section" color="surface" variant="tonal">
          <VCardTitle class="text-subtitle-2 d-flex align-center justify-space-between ga-2">
            <span>{{ t("pages.missionControl.sidePanel.activity") }}</span>
            <VProgressCircular v-if="activityLoading" indeterminate size="18" width="2" />
          </VCardTitle>
          <VCardText>
            <VAlert v-if="activityError" type="error" variant="tonal" class="mb-4">
              {{ t(activityError.messageKey) }}
            </VAlert>
            <div v-if="activity.length" class="mission-side-panel__activity">
              <div v-for="entry in activity" :key="entry.entry_id" class="mission-side-panel__activity-entry">
                <div class="mission-side-panel__activity-head">
                  <VChip size="x-small" variant="tonal">
                    {{ entry.source_kind }}
                  </VChip>
                  <span class="mono text-medium-emphasis">{{ formatCompactDateTime(entry.occurred_at, locale) }}</span>
                </div>
                <div class="mission-side-panel__activity-summary">{{ entry.summary }}</div>
                <div v-if="entry.body_markdown" class="mission-side-panel__activity-body">
                  {{ entry.body_markdown }}
                </div>
                <a
                  v-if="entry.provider_url"
                  class="mission-side-panel__value--link"
                  :href="entry.provider_url"
                  target="_blank"
                  rel="noreferrer"
                >
                  {{ t("pages.missionControl.sidePanel.openProviderEvent") }}
                </a>
              </div>
            </div>
            <div v-else class="text-medium-emphasis">
              {{ t("pages.missionControl.sidePanel.noActivity") }}
            </div>
            <div v-if="hasMoreActivity" class="mt-4">
              <AdaptiveBtn
                variant="text"
                icon="mdi-chevron-down"
                :label="t('pages.missionControl.sidePanel.loadMoreActivity')"
                :loading="activityLoading"
                @click="$emit('loadMoreActivity')"
              />
            </div>
          </VCardText>
        </VCard>

        <VCard class="mission-side-panel__section" color="surface" variant="tonal">
          <VCardTitle class="text-subtitle-2">{{ t("pages.missionControl.sidePanel.launchSurfaces") }}</VCardTitle>
          <VCardText class="mission-side-panel__actions">
            <div v-if="details.launch_surfaces.length" class="d-flex flex-column ga-3">
              <VSheet
                v-for="surface in details.launch_surfaces"
                :key="surface.action_kind"
                class="mission-side-panel__action"
                rounded="xl"
                border
              >
                <div class="d-flex align-start ga-3">
                  <VAvatar color="primary" variant="tonal" size="36">
                    <VIcon :icon="missionControlLaunchSurfaceIcon(surface.action_kind)" size="18" />
                  </VAvatar>
                  <div class="flex-grow-1">
                    <div class="font-weight-medium">{{ t(missionControlLaunchSurfaceLabelKey(surface.action_kind)) }}</div>
                    <div class="text-body-2 text-medium-emphasis mt-1">
                      {{
                        surface.approval_requirement === "owner_review"
                          ? t("pages.missionControl.sidePanel.ownerReviewRequired")
                          : t("pages.missionControl.sidePanel.noApprovalRequired")
                      }}
                    </div>
                    <div v-if="surface.blocked_reason" class="text-body-2 text-error mt-2">
                      {{ surface.blocked_reason }}
                    </div>
                    <AdaptiveBtn
                      v-if="surface.action_kind === 'preview_next_stage' && surface.command_template"
                      class="mt-3"
                      variant="tonal"
                      icon="mdi-rocket-launch-outline"
                      :label="t('pages.missionControl.preview.open')"
                      @click="$emit('openPreview', surface)"
                    />
                  </div>
                </div>
              </VSheet>
            </div>
            <div v-else class="text-medium-emphasis">
              {{ t("pages.missionControl.sidePanel.noLaunchSurfaces") }}
            </div>
          </VCardText>
        </VCard>

        <VCard class="mission-side-panel__section" color="surface" variant="tonal">
          <VCardTitle class="text-subtitle-2">{{ t("pages.missionControl.sidePanel.providerLinks") }}</VCardTitle>
          <VCardText>
            <div v-if="details.provider_deep_links.length" class="d-flex flex-wrap ga-2">
              <VBtn
                v-for="link in details.provider_deep_links"
                :key="link.action_kind + ':' + link.url"
                color="primary"
                variant="tonal"
                :href="link.url"
                target="_blank"
                rel="noreferrer"
                :prepend-icon="missionControlProviderDeepLinkIcon(link.action_kind)"
              >
                {{ t(missionControlProviderDeepLinkLabelKey(link.action_kind)) }}
              </VBtn>
            </div>
            <div v-else class="text-medium-emphasis">
              {{ t("pages.missionControl.sidePanel.noProviderLinks") }}
            </div>
          </VCardText>
        </VCard>
      </VCardText>
    </template>

    <template v-else>
      <VCardText class="py-10 text-center text-medium-emphasis">
        {{ t("pages.missionControl.sidePanel.placeholder") }}
      </VCardText>
    </template>
  </VCard>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import { formatCompactDateTime } from "../../shared/lib/datetime";
import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import type { ApiError } from "../../shared/api/errors";
import {
  buildMissionControlInfoRows,
  isDiscussionPayload,
  isWorkItemPayload,
  missionControlAdjacentNodeRefs,
  missionControlContinuityColor,
  missionControlContinuityStatusLabelKey,
  missionControlGapKindLabelKey,
  missionControlGapSeverityColor,
  missionControlGapSeverityLabelKey,
  missionControlLaunchSurfaceIcon,
  missionControlLaunchSurfaceLabelKey,
  missionControlNodeKindLabelKey,
  missionControlProviderDeepLinkIcon,
  missionControlProviderDeepLinkLabelKey,
  missionControlRelatedNodeSections,
  missionControlStateColor,
  missionControlStateLabelKey,
  missionControlVisibilityColor,
  missionControlVisibilityLabelKey,
  missionControlWatermarkColor,
  missionControlWatermarkLabelKey,
} from "./presenters";
import type {
  MissionControlActivityEntry,
  MissionControlLaunchSurface,
  MissionControlNodeDetails,
  MissionControlNodeRef,
} from "./types";

const props = defineProps<{
  details: MissionControlNodeDetails | null;
  loading: boolean;
  error: ApiError | null;
  activity: MissionControlActivityEntry[];
  activityError: ApiError | null;
  activityLoading: boolean;
  hasMoreActivity: boolean;
  locale: string;
}>();

defineEmits<{
  close: [];
  selectNode: [ref: MissionControlNodeRef];
  loadMoreActivity: [];
  openPreview: [surface: MissionControlLaunchSurface];
}>();

const { t } = useI18n({ useScope: "global" });

const panelIcon = computed(() => {
  switch (props.details?.node.node_kind) {
    case "discussion":
      return "mdi-message-text-outline";
    case "work_item":
      return "mdi-briefcase-outline";
    case "run":
      return "mdi-robot-outline";
    case "pull_request":
      return "mdi-source-pull";
    default:
      return "mdi-radar";
  }
});

const kindLabelKey = computed(() => (props.details ? missionControlNodeKindLabelKey(props.details.node.node_kind) : ""));
const stateLabelKey = computed(() => (props.details ? missionControlStateLabelKey(props.details.node.active_state) : ""));
const continuityLabelKey = computed(() =>
  props.details ? missionControlContinuityStatusLabelKey(props.details.node.continuity_status) : "",
);
const visibilityLabelKey = computed(() =>
  props.details ? missionControlVisibilityLabelKey(props.details.node.visibility_tier) : "",
);
const stateColor = computed(() => (props.details ? missionControlStateColor(props.details.node.active_state) : "secondary"));
const continuityColor = computed(() =>
  props.details ? missionControlContinuityColor(props.details.node.continuity_status) : "secondary",
);
const visibilityColor = computed(() =>
  props.details ? missionControlVisibilityColor(props.details.node.visibility_tier) : "secondary",
);
const infoRows = computed(() => (props.details ? buildMissionControlInfoRows(props.details) : []));
const relatedSections = computed(() => (props.details ? missionControlRelatedNodeSections(props.details) : []));
const adjacentRefs = computed(() => (props.details ? missionControlAdjacentNodeRefs(props.details) : []));

function formatCellValue(value: string): string {
  if (/^\d{4}-\d{2}-\d{2}T/.test(value)) {
    return formatCompactDateTime(value, props.locale);
  }
  return value;
}
</script>

<style scoped>
.mission-side-panel {
  border-radius: 30px;
  overflow: hidden;
}

.mission-side-panel__title {
  white-space: normal;
  line-height: 1.35;
}

.mission-side-panel__subtitle {
  margin-top: 2px;
}

.mission-side-panel__content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.mission-side-panel__section {
  border-radius: 24px;
}

.mission-side-panel__rows {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.mission-side-panel__row,
.mission-side-panel__linked-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.mission-side-panel__label {
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: rgba(15, 23, 42, 0.56);
}

.mission-side-panel__value {
  color: rgb(15, 23, 42);
  word-break: break-word;
}

.mission-side-panel__value--link {
  color: rgb(8, 145, 178);
  text-decoration: none;
}

.mission-side-panel__linked,
.mission-side-panel__actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.mission-side-panel__gaps,
.mission-side-panel__activity {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.mission-side-panel__gap,
.mission-side-panel__action,
.mission-side-panel__activity-entry {
  padding: 14px;
  background: rgba(255, 255, 255, 0.88);
}

.mission-side-panel__activity-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}

.mission-side-panel__activity-summary {
  margin-top: 10px;
  font-weight: 600;
  color: rgb(15, 23, 42);
}

.mission-side-panel__activity-body {
  margin-top: 8px;
  color: rgba(15, 23, 42, 0.74);
  white-space: pre-wrap;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
