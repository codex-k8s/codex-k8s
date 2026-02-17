<template>
  <div>
    <PageHeader :title="t('pages.runtimeDeployTaskDetails.title')">
      <template #leading>
        <BackBtn :label="t('common.back')" @click="goBack" />
      </template>
      <template #actions>
        <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :disabled="loading" @click="loadTask" />
      </template>
    </PageHeader>

    <VAlert v-if="error" type="error" variant="tonal" class="mt-4">
      {{ t(error.messageKey) }}
    </VAlert>

    <template v-if="task">
      <VRow class="mt-4" density="compact">
        <VCol cols="12">
          <VCard variant="outlined">
            <VCardTitle class="d-flex align-center justify-space-between ga-2 flex-wrap">
              <span>{{ t("pages.runtimeDeployTaskDetails.summary") }}</span>
              <VChip size="small" variant="tonal" class="font-weight-bold" :color="colorForRunStatus(task.status)">
                {{ task.status }}
              </VChip>
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
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import BackBtn from "../../shared/ui/BackBtn.vue";
import PageHeader from "../../shared/ui/PageHeader.vue";
import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { formatDateTime } from "../../shared/lib/datetime";
import { colorForRunStatus } from "../../shared/lib/chips";
import { getRuntimeDeployTask } from "../../features/runtime-deploy/api";
import type { RuntimeDeployTask } from "../../features/runtime-deploy/types";

const props = defineProps<{ runId: string }>();

const { t, locale } = useI18n({ useScope: "global" });
const router = useRouter();

const task = ref<RuntimeDeployTask | null>(null);
const loading = ref(false);
const error = ref<ApiError | null>(null);

const sortedLogs = computed(() => {
  const logs = task.value?.logs ? [...task.value.logs] : [];
  logs.sort((a, b) => String(b.created_at || "").localeCompare(String(a.created_at || "")));
  return logs;
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

async function loadTask(): Promise<void> {
  loading.value = true;
  error.value = null;
  try {
    task.value = await getRuntimeDeployTask(props.runId);
  } catch (err) {
    error.value = normalizeApiError(err);
    task.value = null;
  } finally {
    loading.value = false;
  }
}

function goBack(): void {
  void router.push({ name: "runtime-deploy-tasks" });
}

onMounted(() => void loadTask());
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
</style>
