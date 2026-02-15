<template>
  <div>
    <PageHeader :title="t('pages.waitQueue.title')" :hint="t('pages.waitQueue.hint')">
      <template #actions>
        <VBtn variant="tonal" icon="mdi-refresh" :title="t('common.refresh')" :disabled="runs.waitsLoading" @click="runs.loadRunWaits()" />
      </template>
    </PageHeader>

    <VAlert v-if="runs.error" type="error" variant="tonal" class="mt-4">
      {{ t(runs.error.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VRow density="compact">
          <VCol cols="12" md="3">
            <VTextField v-model.trim="runs.waitsFilters.triggerKind" :label="t('pages.runs.runType')" hide-details clearable />
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
          <VBtn
            variant="tonal"
            icon="mdi-check"
            :title="t('pages.runs.applyFilters')"
            @click="runs.loadRunWaits()"
            :disabled="runs.waitsLoading"
          />
          <VBtn variant="text" icon="mdi-backspace-outline" :title="t('pages.runs.resetFilters')" @click="reset" />
        </div>
      </VCardText>
    </VCard>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable
          :headers="headers"
          :items="runs.waitQueue"
          :loading="runs.waitsLoading"
          :items-per-page="10"
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
              v-if="item.project_id"
              class="text-primary font-weight-bold text-decoration-none"
              :to="{ name: 'project-details', params: { projectId: item.project_id } }"
            >
              {{ item.project_name || item.project_slug || item.project_id }}
            </RouterLink>
            <span v-else class="text-medium-emphasis">-</span>
          </template>

          <template #item.wait_since="{ item }">
            <span class="text-medium-emphasis">{{ formatDateTime(item.wait_since, locale) }}</span>
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
import { onMounted } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import PageHeader from "../../shared/ui/PageHeader.vue";
import { formatDateTime } from "../../shared/lib/datetime";
import { colorForRunStatus } from "../../shared/lib/chips";
import { useRunsStore } from "../../features/runs/store";

const runs = useRunsStore();
const { t, locale } = useI18n({ useScope: "global" });

const headers = [
  { title: t("table.fields.status"), key: "status", width: 140, align: "center" },
  { title: t("table.fields.project"), key: "project", sortable: false, width: 220, align: "center" },
  { title: t("table.fields.trigger_kind"), key: "trigger_kind", width: 160, align: "center" },
  { title: t("table.fields.agent_key"), key: "agent_key", width: 160, align: "center" },
  { title: t("table.fields.wait_state"), key: "wait_state", width: 160, align: "center" },
  { title: t("table.fields.wait_since"), key: "wait_since", width: 180, align: "center" },
  { title: "", key: "actions", sortable: false, width: 72, align: "end" },
] as const;

function reset(): void {
  runs.waitsFilters.triggerKind = "";
  runs.waitsFilters.status = "";
  runs.waitsFilters.agentKey = "";
  runs.waitsFilters.waitState = "";
}

onMounted(() => void runs.loadRunWaits());
</script>
