<template>
  <div>
    <PageHeader :title="t('pages.runtimeDeployTasks.title')" :hint="t('pages.runtimeDeployTasks.hint')">
      <template #actions>
        <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :disabled="loading" @click="loadTasks" />
      </template>
    </PageHeader>

    <VAlert v-if="error" type="error" variant="tonal" class="mt-4">
      {{ t(error.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VRow density="compact">
          <VCol cols="12" md="4">
            <VSelect v-model="statusFilter" :items="statusOptions" :label="t('table.fields.status')" hide-details clearable />
          </VCol>
          <VCol cols="12" md="4">
            <VSelect v-model="targetEnvFilter" :items="targetEnvOptions" :label="t('table.fields.target_env')" hide-details clearable />
          </VCol>
        </VRow>
      </VCardText>
      <VCardActions>
        <VSpacer />
        <AdaptiveBtn variant="tonal" icon="mdi-check" :label="t('pages.runs.applyFilters')" :disabled="loading" @click="loadTasks" />
        <AdaptiveBtn variant="text" icon="mdi-backspace-outline" :label="t('pages.runs.resetFilters')" @click="resetFilters" />
      </VCardActions>
    </VCard>

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
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import PageHeader from "../../shared/ui/PageHeader.vue";
import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { formatDateTime } from "../../shared/lib/datetime";
import { colorForRunStatus } from "../../shared/lib/chips";
import { listRuntimeDeployTasks } from "../../features/runtime-deploy/api";
import type { RuntimeDeployTask } from "../../features/runtime-deploy/types";

const { t, locale } = useI18n({ useScope: "global" });

const loading = ref(false);
const error = ref<ApiError | null>(null);
const statusFilter = ref<"" | "pending" | "running" | "succeeded" | "failed">("");
const targetEnvFilter = ref("");
const items = ref<RuntimeDeployTask[]>([]);

const statusOptions = computed(() => [
  { title: t("context.allObjects"), value: "" },
  { title: "pending", value: "pending" },
  { title: "running", value: "running" },
  { title: "succeeded", value: "succeeded" },
  { title: "failed", value: "failed" },
]);

const targetEnvOptions = computed(() => [
  { title: t("context.allObjects"), value: "" },
  { title: "ai", value: "ai" },
  { title: "production", value: "production" },
  { title: "prod", value: "prod" },
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

async function loadTasks(): Promise<void> {
  loading.value = true;
  error.value = null;
  try {
    items.value = await listRuntimeDeployTasks({
      status: statusFilter.value || undefined,
      targetEnv: targetEnvFilter.value,
    }, 200);
  } catch (err) {
    error.value = normalizeApiError(err);
  } finally {
    loading.value = false;
  }
}

function resetFilters(): void {
  statusFilter.value = "";
  targetEnvFilter.value = "";
}

onMounted(() => void loadTasks());
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
