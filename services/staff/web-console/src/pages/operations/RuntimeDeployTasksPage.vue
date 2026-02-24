<template>
  <div>
    <PageHeader :title="t('pages.runtimeDeployTasks.title')" :hint="t('pages.runtimeDeployTasks.hint')">
      <template #actions>
        <div class="d-flex align-center ga-2">
          <VChip size="small" variant="tonal" :color="realtimeChipColor">
            {{ t("pages.runtimeDeployTasks.realtime") }}: {{ t(realtimeChipLabelKey) }}
          </VChip>
          <VSelect
            v-model="statusFilter"
            class="status-select"
            density="compact"
            variant="outlined"
            :items="statusOptions"
            :label="t('table.fields.status')"
            hide-details
            clearable
          />
        </div>
      </template>
    </PageHeader>

    <VAlert v-if="error" type="error" variant="tonal" class="mt-4">
      {{ t(error.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable
          v-model:page="tablePage"
          :headers="headers"
          :items="items"
          :loading="loading"
          :items-per-page="itemsPerPage"
          density="comfortable"
          hover
        >
          <template #item.status="{ item }">
            <div class="d-flex justify-center">
              <VChip size="small" variant="tonal" class="font-weight-bold" :color="colorForRunStatus(item.status)">
                {{ item.status }}
              </VChip>
            </div>
          </template>
          <template #item.repository_full_name="{ item }">
            <RouterLink
              class="text-primary font-weight-bold text-decoration-none mono"
              :to="{ name: 'runtime-deploy-task-details', params: { runId: item.run_id } }"
            >
              {{ item.repository_full_name || "-" }}
            </RouterLink>
          </template>
          <template #item.target_env="{ item }">
            <span class="mono text-medium-emphasis">{{ envLabel(item.result_target_env || item.target_env) }}</span>
          </template>
          <template #item.namespace="{ item }">
            <span class="mono text-medium-emphasis">{{ item.result_namespace || item.namespace || "-" }}</span>
          </template>
          <template #item.updated_at="{ item }">
            <span class="text-medium-emphasis">{{ formatDateTime(item.updated_at || item.created_at, locale) }}</span>
          </template>
          <template #item.actions="{ item }">
            <VTooltip :text="t('scaffold.rowActions.view')">
              <template #activator="{ props: tipProps }">
                <VBtn
                  v-bind="tipProps"
                  size="small"
                  variant="text"
                  icon="mdi-open-in-new"
                  :to="{ name: 'runtime-deploy-task-details', params: { runId: item.run_id } }"
                />
              </template>
            </VTooltip>
          </template>
          <template #no-data>
            <div class="py-8 text-medium-emphasis">
              {{ t("states.noRuntimeDeployTasks") }}
            </div>
          </template>
        </VDataTable>
      </VCardText>
    </VCard>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import PageHeader from "../../shared/ui/PageHeader.vue";
import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { formatDateTime } from "../../shared/lib/datetime";
import { colorForRunStatus } from "../../shared/lib/chips";
import { createProgressiveTableState } from "../../shared/lib/progressive-table";
import { bindRealtimePageLifecycle } from "../../shared/ws/lifecycle";
import { useUiContextStore } from "../../features/ui-context/store";
import { listRuntimeDeployTasks } from "../../features/runtime-deploy/api";
import { subscribeRuntimeDeployRealtime, type RuntimeDeployRealtimeState } from "../../features/runtime-deploy/realtime";
import type { RuntimeDeployTaskListItem } from "../../features/runtime-deploy/types";

const { t, locale } = useI18n({ useScope: "global" });
const uiContext = useUiContextStore();

const loading = ref(false);
const error = ref<ApiError | null>(null);
const statusFilter = ref<"" | "pending" | "running" | "succeeded" | "failed" | null>("");
const items = ref<RuntimeDeployTaskListItem[]>([]);
const itemsPerPage = 15;
const paging = createProgressiveTableState({ itemsPerPage });
const tablePage = paging.page;
const trackedRunRealtimeStates = ref<Record<string, RuntimeDeployRealtimeState>>({});
const reloadPending = ref(false);
const realtimeReloadTimer = ref<number | null>(null);
const fallbackPollTimer = ref<number | null>(null);
const realtimeSubscriptions = new Map<string, () => void>();
const stopLifecycleBindingRef = ref<(() => void) | null>(null);

const activeRunIDs = computed(() => {
  const out: string[] = [];
  for (const item of items.value) {
    const status = String(item.status || "").trim().toLowerCase();
    if (status !== "pending" && status !== "running") {
      continue;
    }
    const runID = String(item.run_id || "").trim();
    if (!runID) continue;
    out.push(runID);
  }
  return out;
});

const realtimeState = computed<RuntimeDeployRealtimeState>(() => {
  if (!activeRunIDs.value.length) return "connected";
  const states = activeRunIDs.value.map((runID) => trackedRunRealtimeStates.value[runID] ?? "connecting");
  if (states.some((state) => state === "connected")) return "connected";
  if (states.some((state) => state === "reconnecting")) return "reconnecting";
  return "connecting";
});

const realtimeChipColor = computed(() => {
  if (realtimeState.value === "connected") return "success";
  if (realtimeState.value === "reconnecting") return "warning";
  return "secondary";
});

const realtimeChipLabelKey = computed(() => {
  if (realtimeState.value === "connected") return "pages.runtimeDeployTasks.realtimeConnected";
  if (realtimeState.value === "reconnecting") return "pages.runtimeDeployTasks.realtimeReconnecting";
  return "pages.runtimeDeployTasks.realtimeConnecting";
});

const statusOptions = computed(() => [
  { title: t("context.allObjects"), value: "" },
  { title: "pending", value: "pending" },
  { title: "running", value: "running" },
  { title: "succeeded", value: "succeeded" },
  { title: "failed", value: "failed" },
]);

const headers = computed(() => ([
  { title: t("table.fields.status"), key: "status", align: "center", width: 140 },
  { title: t("table.fields.repository_full_name"), key: "repository_full_name", align: "center", width: 360 },
  { title: t("table.fields.target_env"), key: "target_env", align: "center", width: 140 },
  { title: t("table.fields.namespace"), key: "namespace", align: "center", width: 220 },
  { title: t("table.fields.runtime_mode"), key: "runtime_mode", align: "center", width: 140 },
  { title: t("table.fields.build_ref"), key: "build_ref", align: "center", width: 160 },
  { title: t("table.fields.updated_at"), key: "updated_at", align: "center", width: 180 },
  { title: "", key: "actions", sortable: false, align: "end", width: 72 },
]) as const);

function normalizeEnv(value: string | null | undefined): "ai" | "production" | string {
  const v = String(value || "").trim().toLowerCase();
  if (v === "" || v === "prod" || v === "production") return "production";
  if (v === "ai") return "ai";
  return v;
}

function envLabel(value: string | null | undefined): string {
  const v = normalizeEnv(value);
  if (v === "production") return "production";
  if (v === "ai") return "ai";
  return v || "-";
}

function matchesUiEnv(item: RuntimeDeployTaskListItem, uiEnv: "ai" | "production" | "all"): boolean {
  if (uiEnv === "all") return true;
  const env = normalizeEnv(item.result_target_env || item.target_env);
  return env === uiEnv;
}

async function loadTasks(): Promise<void> {
  if (loading.value) {
    reloadPending.value = true;
    return;
  }

  loading.value = true;
  error.value = null;
  try {
    // We intentionally load without server-side env filtering to handle legacy values
    // (e.g. target_env="prod" or empty string) and keep UI filter stable.
    const loaded = await listRuntimeDeployTasks({
      status: statusFilter.value || undefined,
      targetEnv: "",
    }, paging.limit.value);
    paging.markLoaded(loaded.length);
    items.value = loaded.filter((x) => matchesUiEnv(x, uiContext.env));
  } catch (err) {
    error.value = normalizeApiError(err);
  } finally {
    loading.value = false;
    if (reloadPending.value) {
      reloadPending.value = false;
      void loadTasks();
    }
  }
}

async function reloadTasks(): Promise<void> {
  paging.reset();
  await loadTasks();
}

async function loadMoreTasksIfNeeded(nextPage: number, prevPage: number): Promise<void> {
  if (loading.value) {
    return;
  }
  if (!paging.shouldGrowForPage(items.value.length, nextPage, prevPage)) {
    return;
  }
  await loadTasks();
}

function upsertRealtimeState(runID: string, state: RuntimeDeployRealtimeState): void {
  trackedRunRealtimeStates.value = {
    ...trackedRunRealtimeStates.value,
    [runID]: state,
  };
}

function removeRealtimeState(runID: string): void {
  if (!(runID in trackedRunRealtimeStates.value)) return;
  const next = { ...trackedRunRealtimeStates.value };
  delete next[runID];
  trackedRunRealtimeStates.value = next;
}

function stopRealtimeSubscriptions(): void {
  for (const stop of realtimeSubscriptions.values()) {
    stop();
  }
  realtimeSubscriptions.clear();
  trackedRunRealtimeStates.value = {};
}

function syncRealtimeSubscriptions(): void {
  const nextRunIDs = new Set(activeRunIDs.value);

  for (const [runID, stop] of realtimeSubscriptions.entries()) {
    if (nextRunIDs.has(runID)) continue;
    stop();
    realtimeSubscriptions.delete(runID);
    removeRealtimeState(runID);
  }

  for (const runID of nextRunIDs.values()) {
    if (realtimeSubscriptions.has(runID)) continue;
    upsertRealtimeState(runID, "connecting");
    const stop = subscribeRuntimeDeployRealtime({
      runId: runID,
      onMessage: () => {
        scheduleRealtimeReload();
      },
      onStateChange: (state) => {
        upsertRealtimeState(runID, state);
      },
    });
    realtimeSubscriptions.set(runID, stop);
  }
}

function scheduleRealtimeReload(): void {
  if (realtimeReloadTimer.value !== null) return;
  realtimeReloadTimer.value = window.setTimeout(() => {
    realtimeReloadTimer.value = null;
    void loadTasks();
  }, 700);
}

function clearRealtimeReloadTimer(): void {
  if (realtimeReloadTimer.value === null) return;
  window.clearTimeout(realtimeReloadTimer.value);
  realtimeReloadTimer.value = null;
}

function startFallbackPolling(): void {
  if (fallbackPollTimer.value !== null) return;
  fallbackPollTimer.value = window.setInterval(() => {
    void loadTasks();
  }, 15000);
}

function stopFallbackPolling(): void {
  if (fallbackPollTimer.value === null) return;
  window.clearInterval(fallbackPollTimer.value);
  fallbackPollTimer.value = null;
}

function stopLifecycleBinding(): void {
  stopLifecycleBindingRef.value?.();
  stopLifecycleBindingRef.value = null;
}

function handlePageSuspend(): void {
  clearRealtimeReloadTimer();
  stopRealtimeSubscriptions();
}

function handlePageResume(): void {
  syncRealtimeSubscriptions();
  void loadTasks();
}

onMounted(() => {
  void reloadTasks();
  startFallbackPolling();
  stopLifecycleBindingRef.value = bindRealtimePageLifecycle({
    onResume: handlePageResume,
    onSuspend: handlePageSuspend,
  });
});

onBeforeUnmount(() => {
  stopLifecycleBinding();
  stopFallbackPolling();
  clearRealtimeReloadTimer();
  stopRealtimeSubscriptions();
});

watch(
  () => uiContext.env,
  () => void reloadTasks(),
);

watch(
  () => statusFilter.value,
  () => void reloadTasks(),
);

watch(
  activeRunIDs,
  () => {
    syncRealtimeSubscriptions();
  },
  { immediate: true },
);

watch(
  tablePage,
  (nextPage, prevPage) => void loadMoreTasksIfNeeded(nextPage, prevPage),
);
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
.status-select {
  min-width: 220px;
}
</style>
