<template>
  <div>
    <PageHeader :title="t('pages.runs.title')">
      <template #actions>
        <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :loading="runs.loading" @click="refreshAll" />
      </template>
    </PageHeader>

    <VAlert v-if="runs.error" type="error" variant="tonal" class="mt-4">
      {{ t(runs.error.messageKey) }}
    </VAlert>
    <VAlert v-if="runs.approvalsError" type="error" variant="tonal" class="mt-4">
      {{ t(runs.approvalsError.messageKey) }}
    </VAlert>

    <VRow class="mt-4" density="compact">
      <VCol cols="12" md="3">
        <VCard variant="tonal">
          <VCardText>
            <div class="text-caption text-medium-emphasis">{{ t("pages.runs.title") }}</div>
            <div class="text-h6 font-weight-bold">{{ runs.items.length }}</div>
          </VCardText>
        </VCard>
      </VCol>
      <VCol cols="12" md="3">
        <VCard variant="tonal">
          <VCardText class="d-flex align-center justify-space-between ga-2 flex-wrap">
            <div>
              <div class="text-caption text-medium-emphasis">{{ t("pages.runs.runningJobs") }}</div>
              <div class="text-h6 font-weight-bold">{{ runs.runningJobs.length }}</div>
            </div>
            <VBtn size="small" variant="text" icon="mdi-open-in-new" :title="t('pages.runs.details')" :to="{ name: 'running-jobs' }" />
          </VCardText>
        </VCard>
      </VCol>
      <VCol cols="12" md="3">
        <VCard variant="tonal">
          <VCardText class="d-flex align-center justify-space-between ga-2 flex-wrap">
            <div>
              <div class="text-caption text-medium-emphasis">{{ t("pages.runs.waitQueue") }}</div>
              <div class="text-h6 font-weight-bold">{{ runs.waitQueue.length }}</div>
            </div>
            <VBtn size="small" variant="text" icon="mdi-open-in-new" :title="t('pages.runs.details')" :to="{ name: 'wait-queue' }" />
          </VCardText>
        </VCard>
      </VCol>
      <VCol cols="12" md="3">
        <VCard variant="tonal">
          <VCardText class="d-flex align-center justify-space-between ga-2 flex-wrap">
            <div>
              <div class="text-caption text-medium-emphasis">{{ t("pages.runs.pendingApprovals") }}</div>
              <div class="text-h6 font-weight-bold">{{ runs.pendingApprovals.length }}</div>
            </div>
            <VBtn size="small" variant="text" icon="mdi-open-in-new" :title="t('pages.runs.details')" :to="{ name: 'approvals' }" />
          </VCardText>
        </VCard>
      </VCol>
    </VRow>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable v-model:page="runsTablePage" :headers="headers" :items="runs.items" :loading="runs.loading" :items-per-page="runsItemsPerPage" hover>
          <template #item.status="{ item }">
            <div class="d-flex justify-center">
              <VChip size="small" variant="tonal" class="font-weight-bold" :color="colorForRunStatus(item.status)">
                {{ item.status }}
              </VChip>
            </div>
          </template>

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

          <template #item.issue="{ item }">
            <a
              v-if="item.issue_url && item.issue_number"
              class="text-primary font-weight-bold text-decoration-none mono"
              :href="item.issue_url"
              target="_blank"
              rel="noopener noreferrer"
            >
              #{{ item.issue_number }}
            </a>
            <span v-else class="text-medium-emphasis">-</span>
          </template>

          <template #item.pr="{ item }">
            <a
              v-if="item.pr_url && item.pr_number"
              class="text-primary font-weight-bold text-decoration-none mono"
              :href="item.pr_url"
              target="_blank"
              rel="noopener noreferrer"
            >
              #{{ item.pr_number }}
            </a>
            <span v-else class="text-medium-emphasis">-</span>
          </template>

          <template #item.started_at="{ item }">
            <span class="mono text-medium-emphasis">{{ formatDateTime(item.started_at, locale) }}</span>
          </template>
          <template #item.finished_at="{ item }">
            <span class="mono text-medium-emphasis">{{ formatDateTime(item.finished_at, locale) }}</span>
          </template>

          <template #item.actions="{ item }">
            <VTooltip :text="t('pages.runs.details')">
              <template #activator="{ props: tipProps }">
                <VBtn
                  v-bind="tipProps"
                  size="small"
                  variant="text"
                  icon="mdi-open-in-new"
                  :to="{ name: 'run-details', params: { runId: item.id } }"
                />
              </template>
            </VTooltip>
          </template>

          <template #no-data>
            <div class="py-8 text-medium-emphasis">
              {{ t("states.noRuns") }}
            </div>
          </template>
        </VDataTable>
      </VCardText>
    </VCard>
  </div>
</template>

<script setup lang="ts">
// TODO(#19): Добавить table settings + row actions menu через общий DataTable wrapper и master-detail layout для Runs/Approvals.
import { onMounted, watch } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import PageHeader from "../shared/ui/PageHeader.vue";
import AdaptiveBtn from "../shared/ui/AdaptiveBtn.vue";
import { formatDateTime } from "../shared/lib/datetime";
import { colorForRunStatus } from "../shared/lib/chips";
import { createProgressiveTableState } from "../shared/lib/progressive-table";
import { useRunsStore } from "../features/runs/store";

const { t, locale } = useI18n({ useScope: "global" });
const runs = useRunsStore();
const runsItemsPerPage = 20;
const runsPaging = createProgressiveTableState({ itemsPerPage: runsItemsPerPage });
const runsTablePage = runsPaging.page;

const headers = [
  { title: t("pages.runs.status"), key: "status", width: 140, align: "center" },
  { title: t("pages.runs.project"), key: "project", sortable: false, width: 220, align: "center" },
  { title: t("pages.runs.issue"), key: "issue", sortable: false, width: 120, align: "center" },
  { title: t("pages.runs.pr"), key: "pr", sortable: false, width: 120, align: "center" },
  { title: t("pages.runs.runType"), key: "trigger_kind", width: 160, align: "center" },
  { title: t("pages.runs.triggerLabel"), key: "trigger_label", width: 200, align: "center" },
  { title: t("pages.runs.started"), key: "started_at", width: 180, align: "center" },
  { title: t("pages.runs.finished"), key: "finished_at", width: 180, align: "center" },
  { title: "", key: "actions", sortable: false, width: 72, align: "end" },
] as const;

async function loadAll() {
  await Promise.all([
    loadRuns(),
    runs.loadRuntimeViews({ jobsLimit: 20, waitsLimit: 20 }),
    runs.loadPendingApprovals(20),
  ]);
}

async function loadRuns(): Promise<void> {
  await runs.load(runsPaging.limit.value);
  runsPaging.markLoaded(runs.items.length);
}

async function refreshAll(): Promise<void> {
  runsPaging.reset();
  await loadAll();
}

async function loadMoreRunsIfNeeded(nextPage: number, prevPage: number): Promise<void> {
  if (runs.loading) {
    return;
  }
  if (!runsPaging.shouldGrowForPage(runs.items.length, nextPage, prevPage)) {
    return;
  }
  await loadRuns();
}

watch(
  runsTablePage,
  (nextPage, prevPage) => void loadMoreRunsIfNeeded(nextPage, prevPage),
);

onMounted(() => void refreshAll());
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
