<template>
  <div>
    <PageHeader :title="t('pages.runtimeDeployTasks.title')" :hint="t('pages.runtimeDeployTasks.hint')">
      <template #actions>
        <div class="d-flex align-center ga-2">
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
          <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :disabled="loading" @click="loadTasks" />
        </div>
      </template>
    </PageHeader>

    <VAlert v-if="error" type="error" variant="tonal" class="mt-4">
      {{ t(error.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable
          :headers="headers"
          :items="items"
          :loading="loading"
          :items-per-page="15"
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
import { computed, onMounted, ref, watch } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import PageHeader from "../../shared/ui/PageHeader.vue";
import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { formatDateTime } from "../../shared/lib/datetime";
import { colorForRunStatus } from "../../shared/lib/chips";
import { useUiContextStore } from "../../features/ui-context/store";
import { listRuntimeDeployTasks } from "../../features/runtime-deploy/api";
import type { RuntimeDeployTask } from "../../features/runtime-deploy/types";

const { t, locale } = useI18n({ useScope: "global" });
const uiContext = useUiContextStore();

const loading = ref(false);
const error = ref<ApiError | null>(null);
const statusFilter = ref<"" | "pending" | "running" | "succeeded" | "failed" | null>("");
const items = ref<RuntimeDeployTask[]>([]);

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

function matchesUiEnv(item: RuntimeDeployTask, uiEnv: "ai" | "production" | "all"): boolean {
  if (uiEnv === "all") return true;
  const env = normalizeEnv(item.result_target_env || item.target_env);
  return env === uiEnv;
}

async function loadTasks(): Promise<void> {
  loading.value = true;
  error.value = null;
  try {
    // We intentionally load without server-side env filtering to handle legacy values
    // (e.g. target_env="prod" or empty string) and keep UI filter stable.
    const loaded = await listRuntimeDeployTasks({
      status: statusFilter.value || undefined,
      targetEnv: "",
    }, 200);
    items.value = loaded.filter((x) => matchesUiEnv(x, uiContext.env));
  } catch (err) {
    error.value = normalizeApiError(err);
  } finally {
    loading.value = false;
  }
}

onMounted(() => void loadTasks());

watch(
  () => uiContext.env,
  () => void loadTasks(),
);

watch(
  () => statusFilter.value,
  () => void loadTasks(),
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
