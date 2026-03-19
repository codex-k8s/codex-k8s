<template>
  <VDialog :model-value="open" max-width="860" @update:model-value="onDialogModelUpdate">
    <VCard>
      <VCardTitle class="d-flex align-center ga-2">
        <VIcon icon="mdi-rocket-launch-outline" size="22" />
        <span>{{ t("pages.missionControl.preview.title") }}</span>
      </VCardTitle>

      <VCardText v-if="loading" class="py-8 d-flex align-center ga-3">
        <VProgressCircular indeterminate size="22" width="2" color="primary" />
        <span>{{ t("pages.missionControl.preview.loading") }}</span>
      </VCardText>

      <template v-else-if="error">
        <VCardText>
          <VAlert type="error" variant="tonal">
            {{ t(error.messageKey) }}
          </VAlert>
        </VCardText>
      </template>

      <template v-else-if="preview && commandTemplate">
        <VCardText class="d-flex flex-column ga-4">
          <VAlert type="info" variant="tonal" density="comfortable">
            <div class="font-weight-medium">{{ t("pages.missionControl.preview.readOnlyTitle") }}</div>
            <div class="text-body-2 mt-1">
              {{ t("pages.missionControl.preview.readOnlyText") }}
            </div>
          </VAlert>

          <div class="mission-preview-grid">
            <div class="mission-preview-field">
              <div class="mission-preview-label">{{ t("pages.missionControl.preview.fields.node") }}</div>
              <div class="mission-preview-value">{{ nodeTitle }}</div>
            </div>
            <div class="mission-preview-field">
              <div class="mission-preview-label">{{ t("pages.missionControl.preview.fields.thread") }}</div>
              <div class="mission-preview-value">
                {{ t(`pages.missionControl.preview.threadKind.${commandTemplate.thread_kind}`) }} #{{ commandTemplate.thread_number }}
              </div>
            </div>
            <div class="mission-preview-field">
              <div class="mission-preview-label">{{ t("pages.missionControl.preview.fields.targetLabel") }}</div>
              <div class="mission-preview-value"><code>{{ commandTemplate.target_label }}</code></div>
            </div>
            <div class="mission-preview-field">
              <div class="mission-preview-label">{{ t("pages.missionControl.preview.fields.approval") }}</div>
              <div class="mission-preview-value">
                {{
                  commandTemplate.approval_requirement === "owner_review"
                    ? t("pages.missionControl.preview.ownerReviewRequired")
                    : t("pages.missionControl.preview.noApprovalRequired")
                }}
              </div>
            </div>
          </div>

          <VAlert v-if="preview.blocking_reason" type="warning" variant="tonal">
            {{ preview.blocking_reason }}
          </VAlert>

          <div class="d-flex flex-column ga-3">
            <div>
              <div class="mission-preview-label mb-2">{{ t("pages.missionControl.preview.labels.removed") }}</div>
              <div class="d-flex flex-wrap ga-2">
                <VChip
                  v-for="label in preview.label_diff.removed_labels"
                  :key="`removed:${label}`"
                  color="error"
                  size="small"
                  variant="tonal"
                >
                  {{ label }}
                </VChip>
                <span v-if="preview.label_diff.removed_labels.length === 0" class="text-medium-emphasis">
                  {{ t("pages.missionControl.preview.empty") }}
                </span>
              </div>
            </div>

            <div>
              <div class="mission-preview-label mb-2">{{ t("pages.missionControl.preview.labels.added") }}</div>
              <div class="d-flex flex-wrap ga-2">
                <VChip
                  v-for="label in preview.label_diff.added_labels"
                  :key="`added:${label}`"
                  color="success"
                  size="small"
                  variant="tonal"
                >
                  {{ label }}
                </VChip>
                <span v-if="preview.label_diff.added_labels.length === 0" class="text-medium-emphasis">
                  {{ t("pages.missionControl.preview.empty") }}
                </span>
              </div>
            </div>

            <div>
              <div class="mission-preview-label mb-2">{{ t("pages.missionControl.preview.labels.final") }}</div>
              <div class="d-flex flex-wrap ga-2">
                <VChip
                  v-for="label in preview.label_diff.final_labels"
                  :key="`final:${label}`"
                  color="primary"
                  size="small"
                  variant="outlined"
                >
                  {{ label }}
                </VChip>
                <span v-if="preview.label_diff.final_labels.length === 0" class="text-medium-emphasis">
                  {{ t("pages.missionControl.preview.empty") }}
                </span>
              </div>
            </div>
          </div>

          <div class="mission-preview-grid">
            <div class="mission-preview-panel">
              <div class="mission-preview-label">{{ t("pages.missionControl.preview.resolvedGaps") }}</div>
              <div class="d-flex flex-wrap ga-2 mt-2">
                <VChip
                  v-for="gapId in preview.continuity_effect.resolved_gap_ids"
                  :key="`resolved:${gapId}`"
                  color="success"
                  size="small"
                  variant="tonal"
                >
                  {{ gapLabel(gapId) }}
                </VChip>
                <span v-if="preview.continuity_effect.resolved_gap_ids.length === 0" class="text-medium-emphasis">
                  {{ t("pages.missionControl.preview.empty") }}
                </span>
              </div>
            </div>

            <div class="mission-preview-panel">
              <div class="mission-preview-label">{{ t("pages.missionControl.preview.remainingGaps") }}</div>
              <div class="d-flex flex-wrap ga-2 mt-2">
                <VChip
                  v-for="gapId in preview.continuity_effect.remaining_gap_ids"
                  :key="`remaining:${gapId}`"
                  color="warning"
                  size="small"
                  variant="tonal"
                >
                  {{ gapLabel(gapId) }}
                </VChip>
                <span v-if="preview.continuity_effect.remaining_gap_ids.length === 0" class="text-medium-emphasis">
                  {{ t("pages.missionControl.preview.empty") }}
                </span>
              </div>
            </div>
          </div>

          <div class="mission-preview-panel">
            <div class="mission-preview-label">{{ t("pages.missionControl.preview.resultingNodes") }}</div>
            <div class="d-flex flex-wrap ga-2 mt-2">
              <VChip
                v-for="ref in preview.continuity_effect.resulting_node_refs"
                :key="ref.node_kind + ':' + ref.node_public_id"
                size="small"
                variant="outlined"
              >
                {{ t(missionControlNodeKindLabelKey(ref.node_kind)) }} · {{ ref.node_public_id }}
              </VChip>
              <span v-if="preview.continuity_effect.resulting_node_refs.length === 0" class="text-medium-emphasis">
                {{ t("pages.missionControl.preview.empty") }}
              </span>
            </div>
          </div>

          <div class="mission-preview-panel">
            <div class="mission-preview-label">{{ t("pages.missionControl.preview.providerRedirects") }}</div>
            <div class="d-flex flex-wrap ga-2 mt-2">
              <VBtn
                v-for="redirect in preview.continuity_effect.provider_redirects"
                :key="redirect"
                variant="tonal"
                color="primary"
                :href="redirect"
                target="_blank"
                rel="noreferrer"
                prepend-icon="mdi-open-in-new"
              >
                {{ t("pages.missionControl.preview.openRedirect") }}
              </VBtn>
              <span v-if="preview.continuity_effect.provider_redirects.length === 0" class="text-medium-emphasis">
                {{ t("pages.missionControl.preview.empty") }}
              </span>
            </div>
          </div>
        </VCardText>

        <VCardActions>
          <VSpacer />
          <VBtn variant="text" @click="$emit('close')">
            {{ t("common.close") }}
          </VBtn>
        </VCardActions>
      </template>
    </VCard>
  </VDialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import type { ApiError } from "../../shared/api/errors";
import { missionControlGapKindLabelKey, missionControlNodeKindLabelKey } from "./presenters";
import type {
  MissionControlContinuityGap,
  MissionControlLaunchPreview,
  MissionControlStageNextStepTemplate,
} from "./types";

const props = defineProps<{
  open: boolean;
  loading: boolean;
  error: ApiError | null;
  preview: MissionControlLaunchPreview | null;
  commandTemplate: MissionControlStageNextStepTemplate | null;
  nodeTitle: string;
  knownGaps: MissionControlContinuityGap[];
}>();

const emit = defineEmits<{
  close: [];
}>();

const { t } = useI18n({ useScope: "global" });

const gapById = computed(() => {
  const index = new Map<number, MissionControlContinuityGap>();
  for (const gap of props.knownGaps) {
    index.set(gap.gap_id, gap);
  }
  return index;
});

function gapLabel(gapId: number): string {
  const knownGap = gapById.value.get(gapId);
  if (!knownGap) {
    return `#${gapId}`;
  }
  return `${t(missionControlGapKindLabelKey(knownGap.gap_kind))} (#${gapId})`;
}

function onDialogModelUpdate(nextValue: boolean): void {
  if (nextValue) return;
  emit("close");
}
</script>

<style scoped>
.mission-preview-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.mission-preview-field,
.mission-preview-panel {
  border-radius: 22px;
  padding: 16px;
  background: rgba(248, 250, 252, 0.92);
  border: 1px solid rgba(15, 23, 42, 0.08);
}

.mission-preview-label {
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: rgba(15, 23, 42, 0.56);
}

.mission-preview-value {
  margin-top: 8px;
  color: rgb(15, 23, 42);
  word-break: break-word;
}

@media (max-width: 760px) {
  .mission-preview-grid {
    grid-template-columns: 1fr;
  }
}
</style>
