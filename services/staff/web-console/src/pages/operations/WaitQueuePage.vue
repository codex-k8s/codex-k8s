<template>
  <div>
    <PageHeader :title="t('pages.waitQueue.title')" :hint="t('pages.waitQueue.hint')">
      <template #actions>
        <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :disabled="runs.waitsLoading" @click="refreshWaits" />
      </template>
    </PageHeader>

    <VAlert v-if="runs.error" type="error" variant="tonal" class="mt-4">
      {{ t(runs.error.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VRow density="compact">
          <VCol cols="12" md="3">
            <VTextField v-model.trim="runs.waitsFilters.triggerKind" :label="t('pages.runs.triggerKind')" hide-details clearable />
          </VCol>
          <VCol cols="12" md="3">
            <VTextField v-model.trim="runs.waitsFilters.status" :label="t('pages.runs.status')" hide-details clearable />
          </VCol>
          <VCol cols="12" md="3">
            <VTextField v-model.trim="runs.waitsFilters.agentKey" :label="t('pages.runs.agentKey')" hide-details clearable />
          </VCol>
          <VCol cols="12" md="3">
            <VTextField v-model.trim="runs.waitsFilters.waitState" :label="t('pages.runs.waitState')" hide-details clearable />
          </VCol>
        </VRow>
        <div class="d-flex ga-2 mt-3 flex-wrap justify-end">
          <AdaptiveBtn
            variant="tonal"
            icon="mdi-check"
            :label="t('pages.runs.applyFilters')"
            @click="refreshWaits"
            :disabled="runs.waitsLoading"
          />
          <AdaptiveBtn variant="text" icon="mdi-backspace-outline" :label="t('pages.runs.resetFilters')" @click="reset" />
        </div>
      </VCardText>
    </VCard>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable
          v-model:page="tablePage"
          :headers="headers"
          :items="waitRows"
          :loading="runs.waitsLoading"
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

          <template #item.project="{ item }">
            <RouterLink
              v-if="item.projectId"
              class="text-primary font-weight-bold text-decoration-none"
              :to="{ name: 'project-details', params: { projectId: item.projectId } }"
            >
              {{ item.projectLabel }}
            </RouterLink>
            <span v-else class="font-weight-bold">{{ item.projectLabel }}</span>
            <div class="text-caption text-medium-emphasis mono mt-1">{{ item.runId }}</div>
          </template>

          <template #item.run="{ item }">
            <div class="d-flex flex-column ga-1">
              <span class="font-weight-medium mono">{{ item.triggerKind || "-" }}</span>
              <span class="text-body-2 text-medium-emphasis mono">{{ item.agentKey || "-" }}</span>
              <span class="text-caption text-medium-emphasis mono">{{ item.waitState || "-" }}</span>
            </div>
          </template>

          <template #item.dominant_wait="{ item }">
            <div v-if="item.projection" class="d-flex flex-column ga-2">
              <div class="d-flex flex-wrap ga-1">
                <VChip size="x-small" variant="tonal" :color="item.projection.dominantWait.contourColor">
                  {{ t(item.projection.dominantWait.contourLabelKey) }}
                </VChip>
                <VChip size="x-small" variant="tonal" :color="item.projection.dominantWait.limitColor">
                  {{ t(item.projection.dominantWait.limitLabelKey) }}
                </VChip>
                <VChip size="x-small" variant="outlined" :color="item.projection.dominantWait.stateColor">
                  {{ t(item.projection.dominantWait.stateLabelKey) }}
                </VChip>
                <VChip size="x-small" variant="outlined" :color="item.projection.commentMirror.color">
                  {{ t(item.projection.commentMirror.labelKey) }}
                </VChip>
              </div>
              <div class="text-body-2">{{ t(item.projection.dominantWait.operationLabelKey) }}</div>
              <div class="text-caption text-medium-emphasis">
                {{ t(item.projection.dominantWait.confidenceLabelKey) }}
              </div>
              <div v-if="item.projection.relatedWaits.length" class="d-flex flex-wrap ga-1">
                <VChip
                  v-for="related in item.projection.relatedWaits"
                  :key="related.waitId"
                  size="x-small"
                  variant="outlined"
                  :color="related.contourColor"
                >
                  {{ t(related.contourLabelKey) }}
                </VChip>
              </div>
            </div>
            <div v-else class="d-flex flex-column ga-1">
              <span class="text-body-2 mono">{{ item.waitState || "-" }}</span>
              <span class="text-caption text-medium-emphasis mono">{{ item.waitReason || "-" }}</span>
            </div>
          </template>

          <template #item.next_step="{ item }">
            <div v-if="item.projection" class="d-flex flex-column ga-2">
              <VChip size="small" variant="tonal" :color="nextStepForRow(item)?.color">
                {{ t(nextStepForRow(item)?.labelKey || "runs.waits.hints.manualOnly") }}
              </VChip>
              <div class="text-body-2 wait-queue__details">
                {{ nextStepForRow(item)?.summary }}
              </div>
              <div v-if="nextStepForRow(item)?.scheduledAt" class="text-caption text-medium-emphasis">
                {{ t(nextStepForRow(item)?.scheduledAtLabelKey || "pages.runDetails.resumeNotBefore") }}:
                {{ formatDateTime(nextStepForRow(item)?.scheduledAt, locale) }}
              </div>
            </div>
            <div v-else class="text-body-2 text-medium-emphasis">
              {{ t("pages.waitQueue.genericWait") }}
            </div>
          </template>

          <template #item.wait_since="{ item }">
            <div class="d-flex flex-column ga-1">
              <span class="font-weight-medium">{{ formatDateTime(item.waitSince, locale) }}</span>
              <span class="text-caption text-medium-emphasis">{{ formatDurationSince(item.waitSince, locale) }}</span>
            </div>
          </template>

          <template #item.actions="{ item }">
            <VTooltip :text="t('pages.runs.details')">
              <template #activator="{ props: tipProps }">
                <VBtn
                  v-bind="tipProps"
                  size="small"
                  variant="text"
                  icon="mdi-open-in-new"
                  :to="{ name: 'run-details', params: { runId: item.runId } }"
                />
              </template>
            </VTooltip>
          </template>

          <template #no-data>
            <div class="py-8 text-medium-emphasis">
              {{ t("states.noWaitQueue") }}
            </div>
          </template>
        </VDataTable>
      </VCardText>
    </VCard>
  </div>
</template>

<script setup lang="ts">
// TODO(#19): Добавить SLA/heartbeat индикацию и перейти на общий DataTable wrapper (table settings + row actions menu).
import { computed, onMounted, watch } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import PageHeader from "../../shared/ui/PageHeader.vue";
import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import { formatDateTime, formatDurationSince } from "../../shared/lib/datetime";
import { colorForRunStatus } from "../../shared/lib/chips";
import { createProgressiveTableState } from "../../shared/lib/progressive-table";
import { useRunsStore } from "../../features/runs/store";
import { buildRunWaitNextStepView, buildRunWaitQueueRow } from "../../features/runs/wait-presenters";
import type { RunWaitNextStepView, RunWaitQueueRowView } from "../../features/runs/types";

const runs = useRunsStore();
const { t, locale } = useI18n({ useScope: "global" });
const itemsPerPage = 10;
const paging = createProgressiveTableState({ itemsPerPage });
const tablePage = paging.page;
const waitRows = computed(() => runs.waitQueue.map(buildRunWaitQueueRow));

const headers = [
  { title: t("table.fields.status"), key: "status", width: 140, align: "center" },
  { title: t("table.fields.project"), key: "project", sortable: false, width: 220, align: "center" },
  { title: t("table.fields.run"), key: "run", sortable: false, width: 170, align: "center" },
  { title: t("table.fields.dominant_wait"), key: "dominant_wait", sortable: false, width: 270, align: "center" },
  { title: t("table.fields.next_step"), key: "next_step", sortable: false, width: 260, align: "center" },
  { title: t("table.fields.wait_since"), key: "wait_since", width: 180, align: "center" },
  { title: "", key: "actions", sortable: false, width: 72, align: "end" },
] as const;

function nextStepForRow(row: RunWaitQueueRowView): RunWaitNextStepView | null {
  if (!row.projection) {
    return null;
  }

  return buildRunWaitNextStepView(row.projection.dominantWait);
}

function reset(): void {
  runs.waitsFilters.triggerKind = "";
  runs.waitsFilters.status = "";
  runs.waitsFilters.agentKey = "";
  runs.waitsFilters.waitState = "";
}

async function loadWaits(): Promise<void> {
  await runs.loadRunWaits(paging.limit.value);
  paging.markLoaded(runs.waitQueue.length);
}

async function refreshWaits(): Promise<void> {
  paging.reset();
  await loadWaits();
}

async function loadMoreWaitsIfNeeded(nextPage: number, prevPage: number): Promise<void> {
  if (runs.waitsLoading) {
    return;
  }
  if (!paging.shouldGrowForPage(runs.waitQueue.length, nextPage, prevPage)) {
    return;
  }
  await loadWaits();
}

watch(
  tablePage,
  (nextPage, prevPage) => void loadMoreWaitsIfNeeded(nextPage, prevPage),
);

onMounted(() => void refreshWaits());
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.wait-queue__details {
  white-space: pre-line;
}
</style>
