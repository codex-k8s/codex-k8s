<template>
  <div>
    <PageHeader :title="t('pages.runDetails.title')">
      <template #leading>
        <BackBtn :label="t('common.back')" @click="goBack" />
      </template>
      <template #actions>
        <CopyChip :label="t('pages.runDetails.runId')" :value="runId" icon="mdi-identifier" />
        <CopyChip
          v-if="details.run?.correlation_id"
          :label="t('pages.runDetails.correlation')"
          :value="details.run.correlation_id"
          icon="mdi-link-variant"
        />
        <CopyChip v-if="details.run?.namespace" :label="t('pages.runDetails.namespace')" :value="details.run.namespace" icon="mdi-kubernetes" />

        <AdaptiveBtn
          v-if="canDeleteNamespace"
          color="error"
          variant="tonal"
          icon="mdi-delete-outline"
          :label="t('pages.runDetails.deleteNamespace')"
          :loading="details.deletingNamespace"
          @click="confirmDeleteNamespaceOpen = true"
        />
      </template>
    </PageHeader>

    <VAlert v-if="details.error" type="error" variant="tonal" class="mt-4">
      {{ t(details.error.messageKey) }}
    </VAlert>
    <VAlert v-if="details.deleteNamespaceError" type="error" variant="tonal" class="mt-4">
      {{ t(details.deleteNamespaceError.messageKey) }}
    </VAlert>
    <VAlert v-if="details.namespaceDeleteResult" type="success" variant="tonal" class="mt-4">
      <div class="text-body-2">
        {{ t("pages.runDetails.namespace") }}:
        <span class="mono">{{ details.namespaceDeleteResult.namespace }}</span>
        ·
        {{
          details.namespaceDeleteResult.already_deleted
            ? t("pages.runDetails.namespaceAlreadyDeleted")
            : t("pages.runDetails.namespaceDeleted")
        }}
      </div>
    </VAlert>

    <VRow class="mt-4" density="compact">
      <VCol cols="12" md="7">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-1 d-flex align-center justify-space-between ga-2 flex-wrap">
            <span>{{ t("pages.runDetails.title") }}</span>
            <div class="d-flex align-center ga-2 flex-wrap">
              <VChip size="small" variant="tonal" class="font-weight-bold" :color="colorForRunStatus(details.run?.status)">
                {{ details.run?.status || "-" }}
              </VChip>
              <VChip size="small" variant="tonal" :color="realtimeChipColor">
                {{ t("pages.runDetails.realtime") }}: {{ t(realtimeChipLabelKey) }}
              </VChip>
            </div>
          </VCardTitle>
          <VCardText>
            <div class="d-flex flex-column ga-2">
              <div class="text-body-2">
                <strong>{{ t("pages.runDetails.project") }}:</strong>
                <RouterLink
                  v-if="details.run?.project_id"
                  class="text-primary font-weight-bold text-decoration-none"
                  :to="{ name: 'project-details', params: { projectId: details.run.project_id } }"
                >
                  {{ details.run.project_name || details.run.project_slug || details.run.project_id }}
                </RouterLink>
                <span v-else class="text-medium-emphasis">-</span>
              </div>

              <div class="text-body-2">
                <strong>{{ t("pages.runDetails.issue") }}:</strong>
                <a
                  v-if="details.run?.issue_url && details.run?.issue_number"
                  class="text-primary font-weight-bold text-decoration-none mono"
                  :href="details.run.issue_url"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  #{{ details.run.issue_number }}
                </a>
                <span v-else class="text-medium-emphasis">-</span>
              </div>

              <div class="text-body-2">
                <strong>{{ t("pages.runDetails.pr") }}:</strong>
                <a
                  v-if="details.run?.pr_url && details.run?.pr_number"
                  class="text-primary font-weight-bold text-decoration-none mono"
                  :href="details.run.pr_url"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  #{{ details.run.pr_number }}
                </a>
                <span v-else class="text-medium-emphasis">-</span>
              </div>

              <div class="text-body-2">
                <strong>{{ t("pages.runDetails.triggerKind") }}:</strong>
                <span class="mono">{{ details.run?.trigger_kind || "-" }}</span>
                ·
                <strong>{{ t("pages.runDetails.triggerLabel") }}:</strong>
                <span class="mono">{{ details.run?.trigger_label || "-" }}</span>
              </div>

              <div class="text-body-2">
                <strong>{{ t("pages.runDetails.waitState") }}:</strong>
                <span class="mono">{{ details.run?.wait_state || "-" }}</span>
                ·
                <strong>{{ t("pages.runDetails.waitReason") }}:</strong>
                <span class="mono">{{ details.run?.wait_reason || "-" }}</span>
              </div>

              <div class="text-body-2">
                <strong>{{ t("pages.runDetails.agentKey") }}:</strong>
                <span class="mono">{{ details.run?.agent_key || "-" }}</span>
              </div>
            </div>
          </VCardText>
        </VCard>

        <VCard class="mt-4" variant="outlined">
          <VCardTitle>{{ t("pages.runDetails.flowEvents") }} ({{ details.events.length }})</VCardTitle>
          <VCardText>
            <VAlert v-if="!details.events.length" type="info" variant="tonal">
              {{ t("states.noEvents") }}
            </VAlert>
            <VExpansionPanels v-else density="compact" variant="accordion">
              <VExpansionPanel v-for="e in details.events" :key="e.created_at + ':' + e.event_type">
                <VExpansionPanelTitle>
                  <div class="d-flex align-center justify-space-between ga-2 flex-wrap w-100">
                    <VChip size="x-small" variant="tonal" class="font-weight-bold">{{ e.event_type }}</VChip>
                    <span class="mono text-medium-emphasis">{{ formatDateTime(e.created_at, locale) }}</span>
                  </div>
                </VExpansionPanelTitle>
                <VExpansionPanelText>
                  <pre class="pre">{{ prettyJSON(e.payload_json) }}</pre>
                </VExpansionPanelText>
              </VExpansionPanel>
            </VExpansionPanels>
          </VCardText>
        </VCard>
      </VCol>

      <VCol cols="12" md="5">
        <VCard v-if="waitProjection" variant="outlined">
          <VCardTitle class="text-subtitle-1 d-flex align-center justify-space-between ga-2 flex-wrap">
            <span>{{ t("pages.runDetails.waitProjectionTitle") }}</span>
            <div class="d-flex flex-wrap ga-2">
              <VChip size="small" variant="tonal" color="warning">
                {{ t(waitProjection.waitStateLabelKey) }}
              </VChip>
              <VChip size="small" variant="tonal" color="secondary">
                {{ t(waitProjection.waitReasonLabelKey) }}
              </VChip>
              <VChip size="small" variant="outlined" :color="waitProjection.commentMirror.color">
                {{ t(waitProjection.commentMirror.labelKey) }}
              </VChip>
            </div>
          </VCardTitle>
          <VCardText class="d-flex flex-column ga-4">
            <section class="run-wait-card__section">
              <div class="text-subtitle-2 font-weight-bold">{{ t("pages.runDetails.dominantWaitTitle") }}</div>
              <div class="d-flex flex-wrap ga-1 mt-2">
                <VChip size="x-small" variant="tonal" :color="waitProjection.dominantWait.contourColor">
                  {{ t(waitProjection.dominantWait.contourLabelKey) }}
                </VChip>
                <VChip size="x-small" variant="tonal" :color="waitProjection.dominantWait.limitColor">
                  {{ t(waitProjection.dominantWait.limitLabelKey) }}
                </VChip>
                <VChip size="x-small" variant="outlined" :color="waitProjection.dominantWait.stateColor">
                  {{ t(waitProjection.dominantWait.stateLabelKey) }}
                </VChip>
                <VChip size="x-small" variant="outlined" :color="waitProjection.dominantWait.confidenceColor">
                  {{ t(waitProjection.dominantWait.confidenceLabelKey) }}
                </VChip>
              </div>

              <div class="run-wait-card__grid mt-3">
                <div class="run-wait-card__meta">
                  <div class="run-wait-card__label">{{ t("pages.runDetails.waitId") }}</div>
                  <div class="mono">{{ waitProjection.dominantWait.waitId }}</div>
                </div>
                <div class="run-wait-card__meta">
                  <div class="run-wait-card__label">{{ t("table.fields.operation_class") }}</div>
                  <div>{{ t(waitProjection.dominantWait.operationLabelKey) }}</div>
                </div>
                <div class="run-wait-card__meta">
                  <div class="run-wait-card__label">{{ t("pages.runDetails.waitEnteredAt") }}</div>
                  <div>{{ formatDateTime(waitProjection.dominantWait.enteredAt, locale) }}</div>
                </div>
                <div class="run-wait-card__meta">
                  <div class="run-wait-card__label">{{ t("table.fields.attempts") }}</div>
                  <div>{{ waitProjection.dominantWait.attemptsUsed }}/{{ waitProjection.dominantWait.maxAttempts }}</div>
                </div>
                <div class="run-wait-card__meta">
                  <div class="run-wait-card__label">{{ t("pages.runDetails.recoveryHintSource") }}</div>
                  <div>{{ t(waitProjection.dominantWait.recoveryHint.sourceLabelKey) }}</div>
                </div>
                <div class="run-wait-card__meta">
                  <div class="run-wait-card__label">{{ t(nextStepForWait(waitProjection.dominantWait).scheduledAtLabelKey || "pages.runDetails.resumeNotBefore") }}</div>
                  <div>{{ formatDateTime(nextStepForWait(waitProjection.dominantWait).scheduledAt, locale) }}</div>
                </div>
              </div>

              <div class="d-flex flex-wrap ga-2 mt-3">
                <VChip size="small" variant="tonal" :color="waitProjection.dominantWait.recoveryHint.kindColor">
                  {{ t(waitProjection.dominantWait.recoveryHint.kindLabelKey) }}
                </VChip>
                <VChip
                  v-if="waitProjection.dominantWait.manualAction"
                  size="small"
                  variant="tonal"
                  color="error"
                >
                  {{ t(waitProjection.dominantWait.manualAction.kindLabelKey) }}
                </VChip>
              </div>

              <div class="run-wait-card__details mt-3">
                {{ nextStepForWait(waitProjection.dominantWait).detailsMarkdown }}
              </div>

              <VAlert v-if="waitProjection.dominantWait.manualAction" type="warning" variant="tonal" class="mt-3">
                <div class="font-weight-medium">{{ waitProjection.dominantWait.manualAction.summary }}</div>
                <div class="run-wait-card__details mt-2">
                  {{ waitProjection.dominantWait.manualAction.detailsMarkdown }}
                </div>
              </VAlert>
            </section>

            <section v-if="waitProjection.relatedWaits.length" class="run-wait-card__section">
              <div class="text-subtitle-2 font-weight-bold">{{ t("pages.runDetails.relatedWaitsTitle") }}</div>
              <div class="d-flex flex-column ga-3 mt-3">
                <VSheet
                  v-for="related in waitProjection.relatedWaits"
                  :key="related.waitId"
                  rounded="lg"
                  border
                  class="run-wait-card__related"
                >
                  <div class="d-flex align-center justify-space-between ga-2 flex-wrap">
                    <div class="d-flex flex-wrap ga-1">
                      <VChip size="x-small" variant="tonal" :color="related.contourColor">
                        {{ t(related.contourLabelKey) }}
                      </VChip>
                      <VChip size="x-small" variant="tonal" :color="related.limitColor">
                        {{ t(related.limitLabelKey) }}
                      </VChip>
                      <VChip size="x-small" variant="outlined" :color="related.stateColor">
                        {{ t(related.stateLabelKey) }}
                      </VChip>
                    </div>
                    <span class="text-caption text-medium-emphasis">
                      {{ formatCompactDateTime(related.enteredAt, locale) }}
                    </span>
                  </div>
                  <div class="text-body-2 mt-2">{{ t(related.operationLabelKey) }}</div>
                  <div class="text-caption text-medium-emphasis mt-1">{{ t(related.confidenceLabelKey) }}</div>
                  <div class="run-wait-card__details mt-2">
                    {{ nextStepForWait(related).summary }}
                  </div>
                </VSheet>
              </div>
            </section>
          </VCardText>
        </VCard>

        <VAlert
          v-else-if="details.run?.wait_state === 'waiting_backpressure'"
          type="info"
          variant="tonal"
        >
          {{ t("pages.runDetails.waitProjectionPending") }}
        </VAlert>

        <VCard v-if="showWaitDiagnostics" class="mt-4" variant="outlined">
          <VCardTitle class="text-subtitle-1">{{ t("pages.runDetails.realtimeWaitActivity") }}</VCardTitle>
          <VCardText>
            <div v-if="waitRealtimeEntries.length" class="run-wait-feed">
              <VSheet
                v-for="entry in waitRealtimeEntries"
                :key="entry.id"
                rounded="lg"
                border
                class="run-wait-feed__entry"
              >
                <div class="d-flex align-center justify-space-between ga-3 flex-wrap">
                  <div class="d-flex align-center ga-3">
                    <VAvatar :color="entry.color" variant="tonal" size="32">
                      <VIcon :icon="entry.icon" size="16" />
                    </VAvatar>
                    <div>
                      <div class="font-weight-medium">{{ t(entry.labelKey) }}</div>
                      <div class="text-caption text-medium-emphasis mono">{{ entry.waitId }}</div>
                    </div>
                  </div>
                  <span class="text-caption text-medium-emphasis">
                    {{ formatCompactDateTime(entry.occurredAt, locale) }}
                  </span>
                </div>

                <div class="d-flex flex-wrap ga-1 mt-3">
                  <VChip v-if="entry.contourLabelKey" size="x-small" variant="tonal">
                    {{ t(entry.contourLabelKey) }}
                  </VChip>
                  <VChip v-if="entry.limitLabelKey" size="x-small" variant="tonal">
                    {{ t(entry.limitLabelKey) }}
                  </VChip>
                  <VChip v-if="entry.resolutionLabelKey" size="x-small" variant="tonal" :color="entry.color">
                    {{ t(entry.resolutionLabelKey) }}
                  </VChip>
                  <VChip v-if="entry.manualActionLabelKey" size="x-small" variant="tonal" color="error">
                    {{ t(entry.manualActionLabelKey) }}
                  </VChip>
                </div>

                <div v-if="entry.manualActionSummary" class="text-body-2 mt-2">
                  {{ entry.manualActionSummary }}
                </div>
                <div v-if="entry.detailsMarkdown" class="run-wait-card__details mt-2">
                  {{ entry.detailsMarkdown }}
                </div>
              </VSheet>
            </div>
            <VAlert v-else type="info" variant="tonal">
              {{ t("pages.runDetails.realtimeWaitEmpty") }}
            </VAlert>
          </VCardText>
        </VCard>

        <RunTimeline :run="details.run" :events="details.events" :locale="locale" :class="{ 'mt-4': showWaitDiagnostics }" />
      </VCol>
    </VRow>
  </div>

  <ConfirmDialog
    v-model="confirmDeleteNamespaceOpen"
    :title="t('pages.runDetails.deleteNamespace')"
    :message="t('pages.runDetails.deleteNamespaceConfirm')"
    :confirm-text="t('pages.runDetails.deleteNamespace')"
    :cancel-text="t('common.cancel')"
    danger
    @confirm="doDeleteNamespace"
  />

  <VDialog v-model="codexAuthDialogOpen" max-width="720">
    <VCard>
      <VCardTitle class="text-subtitle-1 d-flex align-center justify-space-between ga-2 flex-wrap">
        <span>{{ t("pages.runDetails.codexAuthRequiredTitle") }}</span>
        <VChip size="small" variant="tonal" color="warning" class="font-weight-bold">
          {{ t("pages.runDetails.codexAuthRequiredBadge") }}
        </VChip>
      </VCardTitle>
      <VCardText>
        <div class="text-body-2">
          {{ t("pages.runDetails.codexAuthRequiredText") }}
        </div>

        <VAlert v-if="codexAuthPayload" type="warning" variant="tonal" class="mt-4">
          <div class="d-flex flex-column ga-2">
            <CopyChip
              :label="t('pages.runDetails.codexAuthUserCode')"
              :value="codexAuthPayload.user_code"
              icon="mdi-key-variant"
            />
            <CopyChip
              :label="t('pages.runDetails.codexAuthVerificationUrl')"
              :value="codexAuthPayload.verification_url"
              icon="mdi-open-in-new"
            />
            <AdaptiveBtn
              variant="tonal"
              icon="mdi-open-in-new"
              :label="t('pages.runDetails.codexAuthOpenPage')"
              @click="openCodexAuthPage"
            />
          </div>
        </VAlert>

        <VAlert type="info" variant="tonal" class="mt-4">
          {{ t("pages.runDetails.codexAuthSecurityHint") }}
        </VAlert>
      </VCardText>
      <VCardActions class="justify-end">
        <AdaptiveBtn variant="text" icon="mdi-close" :label="t('common.close')" @click="codexAuthDialogOpen = false" />
      </VCardActions>
    </VCard>
  </VDialog>
</template>

<script setup lang="ts">
// TODO(#19): Доработать Run details: master-detail layout, улучшенный stepper по стадиям/событиям и feedback слой через VSnackbar.
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

import PageHeader from "../shared/ui/PageHeader.vue";
import ConfirmDialog from "../shared/ui/ConfirmDialog.vue";
import CopyChip from "../shared/ui/CopyChip.vue";
import AdaptiveBtn from "../shared/ui/AdaptiveBtn.vue";
import BackBtn from "../shared/ui/BackBtn.vue";
import RunTimeline from "../shared/ui/RunTimeline.vue";
import { formatCompactDateTime, formatDateTime } from "../shared/lib/datetime";
import { colorForRunStatus } from "../shared/lib/chips";
import { subscribeRunRealtime, type RunRealtimeState } from "../features/runs/realtime";
import { useRunDetailsStore } from "../features/runs/store";
import { buildRunWaitNextStepView, buildRunWaitProjectionView, buildRunWaitRealtimeEntryView } from "../features/runs/wait-presenters";
import type { RunWaitItemView, RunWaitNextStepView, RunWaitRealtimeEntryView } from "../features/runs/types";
import { useSnackbarStore } from "../shared/ui/feedback/snackbar-store";

const props = defineProps<{ runId: string }>();

const { t, locale } = useI18n({ useScope: "global" });
const details = useRunDetailsStore();
const router = useRouter();
const snackbar = useSnackbarStore();

const confirmDeleteNamespaceOpen = ref(false);
const canDeleteNamespace = computed(() => Boolean(details.run?.namespace));

type CodexAuthRequiredPayload = { verification_url: string; user_code: string };

const codexAuthDialogOpen = ref(false);
const codexAuthShownKey = ref("");
const realtimeState = ref<RunRealtimeState>("connecting");
const stopRealtimeRef = ref<(() => void) | null>(null);
const fallbackPollTimer = ref<number | null>(null);

const realtimeChipColor = computed(() => {
  if (realtimeState.value === "connected") return "success";
  if (realtimeState.value === "reconnecting") return "warning";
  return "secondary";
});

const realtimeChipLabelKey = computed(() => {
  if (realtimeState.value === "connected") return "pages.runDetails.realtimeConnected";
  if (realtimeState.value === "reconnecting") return "pages.runDetails.realtimeReconnecting";
  return "pages.runDetails.realtimeConnecting";
});
const waitProjection = computed(() => buildRunWaitProjectionView(details.run?.wait_projection));
const waitRealtimeEntries = computed(() =>
  details.waitRealtimeMessages
    .map(buildRunWaitRealtimeEntryView)
    .filter((entry): entry is RunWaitRealtimeEntryView => entry !== null),
);
const showWaitDiagnostics = computed(
  () => Boolean(waitProjection.value || details.run?.wait_state === "waiting_backpressure" || waitRealtimeEntries.value.length),
);

const codexAuthRequiredEvent = computed(() => details.events.find((e) => e.event_type === "run.codex.auth.required") || null);
const codexAuthPayload = computed(() => {
  const raw = codexAuthRequiredEvent.value?.payload_json || "";
  const parsed = parseJSONMaybe(raw);
  if (!parsed || typeof parsed !== "object") return null;
  const candidate = parsed as Partial<CodexAuthRequiredPayload>;
  if (!candidate.verification_url || !candidate.user_code) return null;
  return { verification_url: String(candidate.verification_url), user_code: String(candidate.user_code) };
});

async function loadAll() {
  await details.load(props.runId);
}

function goBack() {
  void router.push({ name: "runs" });
}

function prettyJSON(raw: string): string {
  const value = String(raw || "").trim();
  if (!value) return "";
  try {
    const parsed = JSON.parse(value) as unknown;
    // Some event payloads may be double-encoded as a JSON string.
    if (typeof parsed === "string") {
      const inner = parsed.trim();
      if (!inner) return "";
      try {
        return JSON.stringify(JSON.parse(inner), null, 2);
      } catch {
        return parsed;
      }
    }
    return JSON.stringify(parsed, null, 2);
  } catch {
    return value;
  }
}

function nextStepForWait(item: RunWaitItemView): RunWaitNextStepView {
  return buildRunWaitNextStepView(item);
}

function parseJSONMaybe(raw: string): unknown {
  const value = String(raw || "").trim();
  if (!value) return null;
  try {
    const parsed = JSON.parse(value) as unknown;
    if (typeof parsed === "string") {
      const inner = parsed.trim();
      if (!inner) return null;
      try {
        return JSON.parse(inner) as unknown;
      } catch {
        return parsed;
      }
    }
    return parsed;
  } catch {
    return null;
  }
}

function openCodexAuthPage(): void {
  const url = codexAuthPayload.value?.verification_url;
  if (!url) return;
  window.open(url, "_blank", "noopener,noreferrer");
}

async function doDeleteNamespace() {
  await details.deleteNamespace(props.runId);
  if (!details.deleteNamespaceError) {
    snackbar.success(t("common.saved"));
  }
}

function clearFallbackPolling(): void {
  if (fallbackPollTimer.value !== null) {
    window.clearInterval(fallbackPollTimer.value);
    fallbackPollTimer.value = null;
  }
}

function ensureFallbackPolling(): void {
  if (fallbackPollTimer.value !== null) return;
  fallbackPollTimer.value = window.setInterval(() => {
    void details.load(props.runId);
  }, 10000);
}

function stopRealtime(): void {
  stopRealtimeRef.value?.();
  stopRealtimeRef.value = null;
}

function startRealtime(): void {
  stopRealtime();
  stopRealtimeRef.value = subscribeRunRealtime({
    runId: props.runId,
    includeLogs: false,
    onMessage: (message) => {
      details.applyRealtimeMessage(message);
    },
    onStateChange: (state) => {
      realtimeState.value = state;
      if (state === "connected") {
        clearFallbackPolling();
        return;
      }
      ensureFallbackPolling();
    },
  });
}

onMounted(async () => {
  await loadAll();
  startRealtime();
});

onBeforeUnmount(() => {
  stopRealtime();
  clearFallbackPolling();
});

watch(
  () => props.runId,
  async (nextRunID, prevRunID) => {
    if (nextRunID === prevRunID) return;
    clearFallbackPolling();
    await details.load(nextRunID);
    startRealtime();
  },
);

watch(
  () => [codexAuthRequiredEvent.value?.created_at, codexAuthRequiredEvent.value?.event_type],
  (keyParts) => {
    const createdAt = String(keyParts?.[0] || "").trim();
    const eventType = String(keyParts?.[1] || "").trim();
    if (!createdAt || !eventType || !codexAuthPayload.value) return;

    const key = `${eventType}:${createdAt}`;
    if (codexAuthShownKey.value === key) return;

    codexAuthShownKey.value = key;
    codexAuthDialogOpen.value = true;
  },
  { immediate: true },
);
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
.pre {
  margin: 0;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  overflow: auto;
  max-height: 520px;
  font-size: 12px;
  opacity: 0.95;
}

.run-wait-card__section + .run-wait-card__section {
  border-top: 1px solid rgba(0, 0, 0, 0.08);
  padding-top: 16px;
}

.run-wait-card__grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
}

.run-wait-card__meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.run-wait-card__label {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.62);
}

.run-wait-card__details {
  white-space: pre-line;
  overflow-wrap: anywhere;
}

.run-wait-card__related {
  padding: 12px;
}

.run-wait-feed {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.run-wait-feed__entry {
  padding: 12px;
}
</style>
