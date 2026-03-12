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
        <VDataTableServer
          v-model:page="tablePage"
          v-model:items-per-page="itemsPerPage"
          :headers="headers"
          :items="items"
          :items-length="totalCount"
          :loading="loading"
          :items-per-page-options="itemsPerPageOptions"
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
        </VDataTableServer>
      </VCardText>
    </VCard>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import PageHeader from "../../shared/ui/PageHeader.vue";
import { ApiError } from "../../shared/api/errors";
import { formatDateTime } from "../../shared/lib/datetime";
import { colorForRunStatus } from "../../shared/lib/chips";
import { bindRealtimePageLifecycle } from "../../shared/ws/lifecycle";
import { subscribeRuntimeDeployTasksRealtime, type RuntimeDeployTasksListRealtimeState } from "../../features/runtime-deploy/list-realtime";
import type { RuntimeDeployTaskListItem, RuntimeDeployTasksRealtimeMessage } from "../../features/runtime-deploy/types";

const { t, locale } = useI18n({ useScope: "global" });

const loading = ref(true);
const error = ref<ApiError | null>(null);
const statusFilter = ref<"" | "pending" | "running" | "succeeded" | "failed" | "canceled" | null>("");
const items = ref<RuntimeDeployTaskListItem[]>([]);
const totalCount = ref(0);
const tablePage = ref(1);
const itemsPerPage = ref(15);
const itemsPerPageOptions = [10, 15, 30, 50, 100];
const stopRealtimeRef = ref<(() => void) | null>(null);
const stopLifecycleBindingRef = ref<(() => void) | null>(null);

const realtimeState = ref<RuntimeDeployTasksListRealtimeState>("connecting");

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
  { title: "canceled", value: "canceled" },
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

async function loadTasks(): Promise<void> {
  stopRealtime();
  loading.value = true;
  error.value = null;
  realtimeState.value = "connecting";
  stopRealtimeRef.value = subscribeRuntimeDeployTasksRealtime({
    page: tablePage.value,
    pageSize: itemsPerPage.value,
    status: statusFilter.value || undefined,
    onMessage: applyRealtimeMessage,
    onInitialMessageTimeout: handleInitialRealtimeTimeout,
    onStateChange: (state) => {
      realtimeState.value = state;
    },
  });
}

function applyRealtimeMessage(message: RuntimeDeployTasksRealtimeMessage): void {
  if (message.type === "error") {
    error.value = new ApiError({ kind: "unknown", messageKey: "errors.unknown" });
    loading.value = false;
    return;
  }
  if (!message.pagination) return;
  const maxPage = Math.max(1, Math.ceil(message.pagination.total_count / message.pagination.page_size));
  if (tablePage.value > maxPage) {
    tablePage.value = maxPage;
    return;
  }
  items.value = message.items ?? [];
  totalCount.value = message.pagination.total_count;
  error.value = null;
  loading.value = false;
}

function handleInitialRealtimeTimeout(): void {
  error.value = new ApiError({ kind: "network", messageKey: "errors.realtimeUnavailable" });
  loading.value = false;
}

function stopRealtime(): void {
  stopRealtimeRef.value?.();
  stopRealtimeRef.value = null;
  realtimeState.value = "connecting";
}

function stopLifecycleBinding(): void {
  stopLifecycleBindingRef.value?.();
  stopLifecycleBindingRef.value = null;
}

function handlePageSuspend(): void {
  stopRealtime();
}

function handlePageResume(): void {
  void loadTasks();
}

onMounted(() => {
  void loadTasks();
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
  () => statusFilter.value,
  () => {
    if (tablePage.value !== 1) {
      tablePage.value = 1;
      return;
    }
    void loadTasks();
  },
);

watch(
  itemsPerPage,
  (nextValue, prevValue) => {
    if (nextValue === prevValue) return;
    if (tablePage.value !== 1) {
      tablePage.value = 1;
      return;
    }
    void loadTasks();
  },
);

watch(
  tablePage,
  (nextValue, prevValue) => {
    if (nextValue === prevValue) return;
    void loadTasks();
  },
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
