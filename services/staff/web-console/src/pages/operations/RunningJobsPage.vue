<template>
  <div>
    <PageHeader :title="t('pages.runningJobs.title')" :hint="t('pages.runningJobs.hint')">
      <template #actions>
        <VBtn variant="tonal" prepend-icon="mdi-refresh" :disabled="runs.jobsLoading" @click="runs.loadRunJobs()">
          {{ t("common.refresh") }}
        </VBtn>
      </template>
    </PageHeader>

    <VAlert v-if="runs.error" type="error" variant="tonal" class="mt-4">
      {{ t(runs.error.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VRow density="compact">
          <VCol cols="12" md="4">
            <VTextField v-model.trim="runs.jobsFilters.triggerKind" :label="t('pages.runs.runType')" hide-details clearable />
          </VCol>
          <VCol cols="12" md="4">
            <VTextField v-model.trim="runs.jobsFilters.status" :label="t('pages.runs.status')" hide-details clearable />
          </VCol>
          <VCol cols="12" md="4">
            <VTextField v-model.trim="runs.jobsFilters.agentKey" :label="t('pages.runs.agentKey')" hide-details clearable />
          </VCol>
        </VRow>
        <div class="d-flex ga-2 mt-3 flex-wrap justify-end">
          <VBtn variant="tonal" @click="runs.loadRunJobs()" :disabled="runs.jobsLoading">{{ t("pages.runs.applyFilters") }}</VBtn>
          <VBtn variant="text" @click="reset">{{ t("pages.runs.resetFilters") }}</VBtn>
        </div>
      </VCardText>
    </VCard>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable
          :headers="headers"
          :items="runs.runningJobs"
          :loading="runs.jobsLoading"
          :items-per-page="10"
          density="comfortable"
          hover
        >
          <template #item.project="{ item }">
            <RouterLink
              v-if="item.project_id"
              class="text-primary font-weight-bold text-decoration-none"
              :to="{ name: 'project-details', params: { projectId: item.project_id } }"
            >
              {{ item.project_name || item.project_slug || item.project_id }}
            </RouterLink>
            <span v-else class="text-medium-emphasis">-</span>
          </template>

          <template #item.started_at="{ item }">
            <span class="text-medium-emphasis">{{ formatDateTime(item.started_at, locale) }}</span>
          </template>

          <template #item.actions="{ item }">
            <VBtn size="small" variant="text" :to="{ name: 'run-details', params: { runId: item.id } }">
              {{ t("pages.runs.details") }}
            </VBtn>
          </template>

          <template #no-data>
            <div class="py-8 text-medium-emphasis">
              {{ t("states.noRunningJobs") }}
            </div>
          </template>
        </VDataTable>
      </VCardText>
    </VCard>
  </div>
</template>

<script setup lang="ts">
// TODO(#19): Вынести table settings + row actions menu в общий DataTable wrapper (shared/ui) и подключить master-detail layout.
import { onMounted } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import PageHeader from "../../shared/ui/PageHeader.vue";
import { formatDateTime } from "../../shared/lib/datetime";
import { useRunsStore } from "../../features/runs/store";

const runs = useRunsStore();
const { t, locale } = useI18n({ useScope: "global" });

const headers = [
  { title: "status", key: "status", width: 140 },
  { title: "project", key: "project", sortable: false, width: 220 },
  { title: "run_type", key: "trigger_kind", width: 160 },
  { title: "agent", key: "agent_key", width: 160 },
  { title: "namespace", key: "job_namespace", width: 220 },
  { title: "job", key: "job_name", width: 220 },
  { title: "started", key: "started_at", width: 180 },
  { title: "", key: "actions", sortable: false, width: 120 },
] as const;

function reset(): void {
  runs.jobsFilters.triggerKind = "";
  runs.jobsFilters.status = "";
  runs.jobsFilters.agentKey = "";
}

onMounted(() => void runs.loadRunJobs());
</script>

