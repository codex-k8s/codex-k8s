<template>
  <div>
    <PageHeader :title="t('pages.runtimeDeployTaskDetails.title')">
      <template #leading>
        <BackBtn :label="t('common.back')" @click="goBack" />
      </template>
      <template #actions>
        <div class="d-flex align-center ga-2">
          <AdaptiveBtn
            v-if="canCancel"
            color="warning"
            variant="tonal"
            icon="mdi-cancel"
            :label="t('pages.runtimeDeployTaskDetails.cancelAction')"
            :loading="actionSubmitting && requestedAction === 'cancel'"
            @click="openActionConfirm('cancel')"
          />
          <AdaptiveBtn
            v-if="canStop"
            color="error"
            variant="tonal"
            icon="mdi-stop-circle-outline"
            :label="t('pages.runtimeDeployTaskDetails.stopAction')"
            :loading="actionSubmitting && requestedAction === 'stop'"
            @click="openActionConfirm('stop')"
          />
        </div>
      </template>
    </PageHeader>

    <VAlert v-if="error" type="error" variant="tonal" class="mt-4">
      {{ t(error.messageKey) }}
    </VAlert>
    <VAlert v-if="actionError" type="error" variant="tonal" class="mt-4">
      {{ t(actionError.messageKey) }}
    </VAlert>
    <VAlert v-if="actionResult" type="success" variant="tonal" class="mt-4">
      {{
        actionResult.already_terminal
          ? t("pages.runtimeDeployTaskDetails.actionAlreadyTerminalResult", { status: actionResult.current_status })
          : t("pages.runtimeDeployTaskDetails.actionRequestedResult", {
            action: t(actionTitleKey(actionResult.action)),
            status: actionResult.current_status,
          })
      }}
    </VAlert>

    <template v-if="task">
      <VCard variant="tonal" color="info" class="mt-4">
        <VCardText class="text-body-2">
          <div class="text-subtitle-2 mb-2">{{ t("pages.runtimeDeployTaskDetails.actionsGuideTitle") }}</div>
          <div class="text-medium-emphasis mb-3">{{ t("pages.runtimeDeployTaskDetails.actionsGuideIntro") }}</div>
          <div class="action-guide-grid">
            <div class="action-guide-item">
              <strong>{{ t("pages.runtimeDeployTaskDetails.cancelAction") }}</strong>
              <span>{{ t("pages.runtimeDeployTaskDetails.cancelHelp") }}</span>
            </div>
            <div class="action-guide-item">
              <strong>{{ t("pages.runtimeDeployTaskDetails.stopAction") }}</strong>
              <span>{{ t("pages.runtimeDeployTaskDetails.stopHelp") }}</span>
            </div>
          </div>
          <div class="text-medium-emphasis mt-3">{{ stopAvailabilityText }}</div>
        </VCardText>
      </VCard>

      <VRow class="mt-4" density="compact">
        <VCol cols="12">
          <VCard variant="outlined">
            <VCardTitle class="d-flex align-center justify-space-between ga-2 flex-wrap">
              <span>{{ t("pages.runtimeDeployTaskDetails.summary") }}</span>
              <div class="d-flex align-center ga-2">
                <VChip size="small" variant="tonal" class="font-weight-bold" :color="colorForRunStatus(task.status)">
                  {{ task.status }}
                </VChip>
                <VChip size="small" variant="tonal" :color="realtimeChipColor">
                  {{ t("pages.runtimeDeployTaskDetails.realtime") }}: {{ t(realtimeChipLabelKey) }}
                </VChip>
              </div>
            </VCardTitle>
            <VCardText>
              <div class="summary-grid text-body-2">
                <div><strong>{{ t("table.fields.run") }}:</strong> <span class="mono">{{ task.run_id }}</span></div>
                <div><strong>{{ t("table.fields.repository_full_name") }}:</strong> <span class="mono">{{ task.repository_full_name }}</span></div>
                <div><strong>{{ t("table.fields.runtime_mode") }}:</strong> <span class="mono">{{ task.runtime_mode }}</span></div>
                <div><strong>{{ t("table.fields.target_env") }}:</strong> <span class="mono">{{ task.target_env }}</span></div>
                <div><strong>{{ t("table.fields.namespace") }}:</strong> <span class="mono">{{ task.namespace }}</span></div>
                <div><strong>{{ t("table.fields.slot_no") }}:</strong> <span class="mono">{{ task.slot_no }}</span></div>
                <div><strong>{{ t("table.fields.services_yaml_path") }}:</strong> <span class="mono">{{ task.services_yaml_path }}</span></div>
                <div><strong>{{ t("table.fields.build_ref") }}:</strong> <span class="mono">{{ task.build_ref }}</span></div>
                <div><strong>{{ t("table.fields.attempts") }}:</strong> <span class="mono">{{ task.attempts }}</span></div>
                <div><strong>{{ t("table.fields.created_at") }}:</strong> {{ formatDateTime(task.created_at, locale) }}</div>
                <div><strong>{{ t("table.fields.started_at") }}:</strong> {{ formatDateTime(task.started_at, locale) }}</div>
                <div><strong>{{ t("table.fields.finished_at") }}:</strong> {{ formatDateTime(task.finished_at, locale) }}</div>
                <div v-if="task.terminal_status_source || task.terminal_event_seq" class="summary-wide">
                  <strong>{{ t("pages.runtimeDeployTaskDetails.terminalAudit") }}:</strong>
                  <span class="mono">{{ task.terminal_status_source || "-" }}</span>
                  ·
                  <span class="mono">seq={{ task.terminal_event_seq || 0 }}</span>
                </div>
                <div v-if="task.cancel_requested_at || task.cancel_requested_by || task.cancel_reason" class="summary-wide">
                  <strong>{{ t("pages.runtimeDeployTaskDetails.cancelAudit") }}:</strong>
                  {{ actionAuditText("cancel") }}
                </div>
                <div v-if="task.stop_requested_at || task.stop_requested_by || task.stop_reason" class="summary-wide">
                  <strong>{{ t("pages.runtimeDeployTaskDetails.stopAudit") }}:</strong>
                  {{ actionAuditText("stop") }}
                </div>
                <div v-if="task.last_error" class="summary-wide">
                  <strong>{{ t("table.fields.last_error") }}:</strong> {{ task.last_error }}
                </div>
              </div>
            </VCardText>
          </VCard>
        </VCol>
        <VCol cols="12">
          <VCard variant="outlined">
            <VCardTitle>{{ t("pages.runtimeDeployTaskDetails.logs") }}</VCardTitle>
            <VCardText>
              <VDataTable
                :headers="logHeaders"
                :items="sortedLogs"
                :items-per-page="25"
                density="compact"
              >
                <template #item.level="{ item }">
                  <div class="d-flex justify-center">
                    <VChip size="x-small" variant="tonal" class="font-weight-bold" :color="colorForLevel(item.level)">
                      {{ item.level }}
                    </VChip>
                  </div>
                </template>
                <template #item.created_at="{ item }">
                  <span class="text-medium-emphasis">{{ formatDateTime(item.created_at, locale) }}</span>
                </template>
                <template #item.message="{ item }">
                  <div class="log-message mono">{{ stripAnsi(item.message) }}</div>
                </template>
                <template #no-data>
                  <div class="py-6 text-medium-emphasis">{{ t("states.noRunLogs") }}</div>
                </template>
              </VDataTable>
            </VCardText>
          </VCard>
        </VCol>
      </VRow>
    </template>
  </div>

  <ConfirmDialog
    v-model="actionConfirmOpen"
    :title="t(requestedAction === 'stop' ? 'pages.runtimeDeployTaskDetails.stopAction' : 'pages.runtimeDeployTaskDetails.cancelAction')"
    :message="t(requestedAction === 'stop' ? 'pages.runtimeDeployTaskDetails.stopConfirm' : 'pages.runtimeDeployTaskDetails.cancelConfirm')"
    :confirm-text="t(requestedAction === 'stop' ? 'pages.runtimeDeployTaskDetails.stopAction' : 'pages.runtimeDeployTaskDetails.cancelAction')"
    :cancel-text="t('common.cancel')"
    :danger="requestedAction === 'stop'"
    @confirm="confirmRequestedAction"
  >
    <VTextarea
      v-model.trim="actionReason"
      :label="t('pages.runtimeDeployTaskDetails.actionReasonLabel')"
      :placeholder="t('pages.runtimeDeployTaskDetails.actionReasonPlaceholder')"
      :hint="t('pages.runtimeDeployTaskDetails.actionReasonHint')"
      auto-grow
      rows="3"
      persistent-hint
    />
  </ConfirmDialog>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import BackBtn from "../../shared/ui/BackBtn.vue";
import ConfirmDialog from "../../shared/ui/ConfirmDialog.vue";
import PageHeader from "../../shared/ui/PageHeader.vue";
import { ApiError, normalizeApiError } from "../../shared/api/errors";
import { formatDateTime } from "../../shared/lib/datetime";
import { colorForRunStatus } from "../../shared/lib/chips";
import {
  hasActiveRuntimeDeployLease,
  localizeRuntimeDeployActionError,
  resolveRuntimeDeployLeaseStatus,
} from "../../shared/lib/runtime-deploy-task-actions";
import { bindRealtimePageLifecycle } from "../../shared/ws/lifecycle";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { cancelRuntimeDeployTask, getRuntimeDeployTask, stopRuntimeDeployTask } from "../../features/runtime-deploy/api";
import { subscribeRuntimeDeployRealtime, type RuntimeDeployRealtimeState } from "../../features/runtime-deploy/realtime";
import type { RuntimeDeployTask, RuntimeDeployTaskActionResponse } from "../../features/runtime-deploy/types";

const props = defineProps<{ runId: string }>();

const { t, locale } = useI18n({ useScope: "global" });
const router = useRouter();
const snackbar = useSnackbarStore();

const task = ref<RuntimeDeployTask | null>(null);
const loading = ref(false);
const error = ref<ApiError | null>(null);
const actionError = ref<ApiError | null>(null);
const actionResult = ref<RuntimeDeployTaskActionResponse | null>(null);
const actionSubmitting = ref(false);
const actionConfirmOpen = ref(false);
const requestedAction = ref<"cancel" | "stop" | null>(null);
const actionReason = ref("");
const realtimeState = ref<RuntimeDeployRealtimeState>("connecting");
const stopRealtimeRef = ref<(() => void) | null>(null);
const reloadPending = ref(false);
const stopLifecycleBindingRef = ref<(() => void) | null>(null);

const realtimeChipColor = computed(() => {
  if (realtimeState.value === "connected") return "success";
  if (realtimeState.value === "reconnecting") return "warning";
  return "secondary";
});

const realtimeChipLabelKey = computed(() => {
  if (realtimeState.value === "connected") return "pages.runtimeDeployTaskDetails.realtimeConnected";
  if (realtimeState.value === "reconnecting") return "pages.runtimeDeployTaskDetails.realtimeReconnecting";
  return "pages.runtimeDeployTaskDetails.realtimeConnecting";
});

const sortedLogs = computed(() => {
  const logs = task.value?.logs ? [...task.value.logs] : [];
  logs.sort((a, b) => String(b.created_at || "").localeCompare(String(a.created_at || "")));
  return logs;
});

const normalizedTaskStatus = computed(() => String(task.value?.status || "").trim().toLowerCase());
const canCancel = computed(() => normalizedTaskStatus.value === "pending" || normalizedTaskStatus.value === "running");
const leaseStatus = computed(() => resolveRuntimeDeployLeaseStatus(task.value));
const canStop = computed(() => normalizedTaskStatus.value === "running" && hasActiveRuntimeDeployLease(task.value));
const stopAvailabilityText = computed(() => {
  if (leaseStatus.value === "active") {
    return t("pages.runtimeDeployTaskDetails.stopAvailabilityActive", {
      owner: task.value?.lease_owner || "-",
      until: formatDateTime(task.value?.lease_until, locale.value),
    });
  }
  if (leaseStatus.value === "expired") {
    return t("pages.runtimeDeployTaskDetails.stopAvailabilityExpired", {
      until: formatDateTime(task.value?.lease_until, locale.value),
    });
  }
  return t("pages.runtimeDeployTaskDetails.stopAvailabilityMissing");
});

const logHeaders = computed(() => ([
  { title: t("table.fields.created_at"), key: "created_at", align: "center", width: 180 },
  { title: t("table.fields.stage"), key: "stage", align: "center", width: 140 },
  { title: t("table.fields.level"), key: "level", align: "center", width: 100 },
  { title: t("table.fields.message"), key: "message", align: "start" },
]) as const);

function colorForLevel(value: string): string {
  switch (String(value || "").toLowerCase()) {
    case "error":
      return "error";
    case "warn":
      return "warning";
    default:
      return "info";
  }
}

function stripAnsi(value: string): string {
  return String(value || "").replace(/\u001b\[[0-9;]*m/g, "");
}

async function loadTask(): Promise<void> {
  if (loading.value) {
    reloadPending.value = true;
    return;
  }

  loading.value = true;
  error.value = null;
  try {
    task.value = await getRuntimeDeployTask(props.runId);
  } catch (err) {
    error.value = normalizeApiError(err);
    task.value = null;
  } finally {
    loading.value = false;
    if (reloadPending.value) {
      reloadPending.value = false;
      void loadTask();
    }
  }
}

function actionTitleKey(action: string | null | undefined): string {
  return String(action || "").trim().toLowerCase() === "stop"
    ? "pages.runtimeDeployTaskDetails.stopAction"
    : "pages.runtimeDeployTaskDetails.cancelAction";
}

function actionAuditText(kind: "cancel" | "stop"): string {
  const currentTask = task.value;
  if (!currentTask) return "-";
  const requestedAt = kind === "cancel" ? currentTask.cancel_requested_at : currentTask.stop_requested_at;
  const requestedBy = kind === "cancel" ? currentTask.cancel_requested_by : currentTask.stop_requested_by;
  const reason = kind === "cancel" ? currentTask.cancel_reason : currentTask.stop_reason;

  const parts: string[] = [];
  if (requestedAt) parts.push(formatDateTime(requestedAt, locale.value));
  if (requestedBy) parts.push(String(requestedBy));
  if (reason) parts.push(String(reason));
  return parts.length ? parts.join(" · ") : "-";
}

function openActionConfirm(action: "cancel" | "stop"): void {
  if ((action === "cancel" && !canCancel.value) || (action === "stop" && !canStop.value)) {
    actionError.value = localizeRuntimeDeployActionError(
      new ApiError({
        kind: "http",
        status: 409,
        code: "failed_precondition",
        messageKey: "errors.failedPrecondition",
      }),
      action,
    );
    return;
  }

  actionError.value = null;
  actionReason.value = "";
  requestedAction.value = action;
  actionConfirmOpen.value = true;
}

async function confirmRequestedAction(): Promise<void> {
  if (!requestedAction.value) return;

  const action = requestedAction.value;
  if ((action === "cancel" && !canCancel.value) || (action === "stop" && !canStop.value)) {
    actionError.value = localizeRuntimeDeployActionError(
      new ApiError({
        kind: "http",
        status: 409,
        code: "failed_precondition",
        messageKey: "errors.failedPrecondition",
      }),
      action,
    );
    requestedAction.value = null;
    return;
  }

  actionSubmitting.value = true;
  actionError.value = null;

  try {
    actionResult.value = action === "stop"
      ? await stopRuntimeDeployTask(props.runId, actionReason.value)
      : await cancelRuntimeDeployTask(props.runId, actionReason.value);
    await loadTask();
    snackbar.success(
      actionResult.value.already_terminal
        ? t("pages.runtimeDeployTaskDetails.actionAlreadyTerminalResult", { status: actionResult.value.current_status })
        : t("pages.runtimeDeployTaskDetails.actionRequestedResult", {
          action: t(actionTitleKey(action)),
          status: actionResult.value.current_status,
        }),
    );
  } catch (err) {
    actionError.value = localizeRuntimeDeployActionError(normalizeApiError(err), action);
  } finally {
    actionSubmitting.value = false;
    actionReason.value = "";
    requestedAction.value = null;
  }
}

function goBack(): void {
  void router.push({ name: "runtime-deploy-tasks" });
}

function stopRealtime(): void {
  stopRealtimeRef.value?.();
  stopRealtimeRef.value = null;
}

function stopLifecycleBinding(): void {
  stopLifecycleBindingRef.value?.();
  stopLifecycleBindingRef.value = null;
}

function startRealtime(): void {
  stopRealtime();
  realtimeState.value = "connecting";
  stopRealtimeRef.value = subscribeRuntimeDeployRealtime({
    runId: props.runId,
    onMessage: () => {
      void loadTask();
    },
    onStateChange: (state) => {
      realtimeState.value = state;
    },
  });
}

function handlePageSuspend(): void {
  stopRealtime();
}

function handlePageResume(): void {
  void loadTask();
  startRealtime();
}

onMounted(async () => {
  await loadTask();
  startRealtime();
  stopLifecycleBindingRef.value = bindRealtimePageLifecycle({
    onResume: handlePageResume,
    onSuspend: handlePageSuspend,
  });
});

onBeforeUnmount(() => {
  stopLifecycleBinding();
  stopRealtime();
});

watch(
  () => props.runId,
  async (nextRunID, prevRunID) => {
    if (nextRunID === prevRunID) return;
    await loadTask();
    startRealtime();
  },
);
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.summary-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px 16px;
  align-items: start;
}

@media (min-width: 960px) {
  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.summary-wide {
  grid-column: 1 / -1;
}

.action-guide-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

.action-guide-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

@media (min-width: 960px) {
  .action-guide-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.log-message {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  word-break: break-word;
}
</style>
