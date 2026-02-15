<template>
  <div>
    <PageHeader :title="t('pages.runDetails.title')">
      <template #leading>
        <VBtn variant="text" icon="mdi-arrow-left" :title="t('common.back')" @click="goBack" />
      </template>
      <template #actions>
        <CopyChip :label="t('pages.runDetails.runId')" :value="runId" icon="mdi-identifier" />
        <CopyChip
          v-if="details.run?.correlation_id"
          :label="t('pages.runDetails.correlation')"
          :value="details.run.correlation_id"
          icon="mdi-link-variant"
        />
        <CopyChip v-if="details.run?.namespace" :label="t('pages.runDetails.namespace')" :value="details.run.namespace" icon="mdi-kubernetes" />

        <VBtn variant="tonal" icon="mdi-refresh" :title="t('common.refresh')" :loading="details.loading" @click="loadAll" />
        <VBtn
          v-if="canDeleteNamespace"
          color="error"
          variant="tonal"
          icon="mdi-delete-outline"
          :title="t('pages.runDetails.deleteNamespace')"
          :loading="details.deletingNamespace"
          @click="confirmDeleteNamespaceOpen = true"
        />
      </template>
    </PageHeader>

    <VAlert v-if="details.error" type="error" variant="tonal" class="mt-4">
      {{ t(details.error.messageKey) }}
    </VAlert>
    <VAlert v-if="details.deleteNamespaceError" type="error" variant="tonal" class="mt-4">
      {{ t(details.deleteNamespaceError.messageKey) }}
    </VAlert>
    <VAlert v-if="details.namespaceDeleteResult" type="success" variant="tonal" class="mt-4">
      <div class="text-body-2">
        {{ t("pages.runDetails.namespace") }}:
        <span class="mono">{{ details.namespaceDeleteResult.namespace }}</span>
        ·
        {{
          details.namespaceDeleteResult.already_deleted
            ? t("pages.runDetails.namespaceAlreadyDeleted")
            : t("pages.runDetails.namespaceDeleted")
        }}
      </div>
    </VAlert>

    <VRow class="mt-4" density="compact">
      <VCol cols="12" md="5">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-1 d-flex align-center justify-space-between ga-2 flex-wrap">
            <span>{{ t("pages.runDetails.title") }}</span>
            <VChip size="small" variant="tonal" class="font-weight-bold" :color="colorForRunStatus(details.run?.status)">
              {{ details.run?.status || "-" }}
            </VChip>
          </VCardTitle>
          <VCardText>
            <div class="d-flex flex-column ga-2">
              <div class="text-body-2">
                <strong>{{ t("pages.runDetails.project") }}:</strong>
                <RouterLink
                  v-if="details.run?.project_id"
                  class="text-primary font-weight-bold text-decoration-none"
                  :to="{ name: 'project-details', params: { projectId: details.run.project_id } }"
                >
                  {{ details.run.project_name || details.run.project_slug || details.run.project_id }}
                </RouterLink>
                <span v-else class="text-medium-emphasis">-</span>
              </div>

              <div class="text-body-2">
                <strong>{{ t("pages.runDetails.issue") }}:</strong>
                <a
                  v-if="details.run?.issue_url && details.run?.issue_number"
                  class="text-primary font-weight-bold text-decoration-none mono"
                  :href="details.run.issue_url"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  #{{ details.run.issue_number }}
                </a>
                <span v-else class="text-medium-emphasis">-</span>
              </div>

              <div class="text-body-2">
                <strong>{{ t("pages.runDetails.pr") }}:</strong>
                <a
                  v-if="details.run?.pr_url && details.run?.pr_number"
                  class="text-primary font-weight-bold text-decoration-none mono"
                  :href="details.run.pr_url"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  #{{ details.run.pr_number }}
                </a>
                <span v-else class="text-medium-emphasis">-</span>
              </div>

              <div class="text-body-2">
                <strong>{{ t("pages.runDetails.triggerKind") }}:</strong>
                <span class="mono">{{ details.run?.trigger_kind || "-" }}</span>
                ·
                <strong>{{ t("pages.runDetails.triggerLabel") }}:</strong>
                <span class="mono">{{ details.run?.trigger_label || "-" }}</span>
              </div>

              <div class="text-body-2">
                <strong>{{ t("pages.runDetails.waitState") }}:</strong>
                <span class="mono">{{ details.run?.wait_state || "-" }}</span>
                ·
                <strong>{{ t("pages.runDetails.waitReason") }}:</strong>
                <span class="mono">{{ details.run?.wait_reason || "-" }}</span>
              </div>

              <div class="text-body-2">
                <strong>{{ t("pages.runDetails.agentKey") }}:</strong>
                <span class="mono">{{ details.run?.agent_key || "-" }}</span>
              </div>
            </div>
          </VCardText>
        </VCard>

        <RunTimeline class="mt-4" :run="details.run" :locale="locale" />
      </VCol>

      <VCol cols="12" md="7">
        <LogsViewer
          :lines="details.logs?.tail_lines || []"
          :status="details.logs?.status || ''"
          :updated-at-label="formatDateTime(details.logs?.updated_at, locale)"
          :loading="details.loading"
          :file-name="`run-${runId}.log`"
          @refresh="(n) => details.refreshLogs(runId, n)"
        />

        <VExpansionPanels class="mt-4" variant="accordion">
          <VExpansionPanel>
            <VExpansionPanelTitle>
              {{ t("pages.runDetails.flowEvents") }} ({{ details.events.length }})
            </VExpansionPanelTitle>
            <VExpansionPanelText>
              <VAlert v-if="!details.events.length" type="info" variant="tonal">
                {{ t("states.noEvents") }}
              </VAlert>
              <VList v-else density="compact">
                <VListItem v-for="e in details.events" :key="e.created_at + ':' + e.event_type">
                  <template #title>
                    <div class="d-flex align-center justify-space-between ga-2 flex-wrap">
                      <VChip size="x-small" variant="tonal" class="font-weight-bold">{{ e.event_type }}</VChip>
                      <span class="mono text-medium-emphasis">{{ formatDateTime(e.created_at, locale) }}</span>
                    </div>
                  </template>
                  <template #subtitle>
                    <pre class="pre mt-2">{{ e.payload_json }}</pre>
                  </template>
                </VListItem>
              </VList>
            </VExpansionPanelText>
          </VExpansionPanel>

          <VExpansionPanel>
            <VExpansionPanelTitle>
              {{ t("pages.runDetails.rawLogsSnapshot") }}
            </VExpansionPanelTitle>
            <VExpansionPanelText>
              <pre class="pre">{{ details.logs?.snapshot_json || "{}" }}</pre>
            </VExpansionPanelText>
          </VExpansionPanel>
        </VExpansionPanels>
      </VCol>
    </VRow>
  </div>

  <ConfirmDialog
    v-model="confirmDeleteNamespaceOpen"
    :title="t('pages.runDetails.deleteNamespace')"
    :message="t('pages.runDetails.deleteNamespaceConfirm')"
    :confirm-text="t('pages.runDetails.deleteNamespace')"
    :cancel-text="t('common.cancel')"
    danger
    @confirm="doDeleteNamespace"
  />
</template>

<script setup lang="ts">
// TODO(#19): Доработать Run details: master-detail layout, улучшенный stepper по стадиям/событиям и feedback слой через VSnackbar.
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

import PageHeader from "../shared/ui/PageHeader.vue";
import ConfirmDialog from "../shared/ui/ConfirmDialog.vue";
import CopyChip from "../shared/ui/CopyChip.vue";
import LogsViewer from "../shared/ui/LogsViewer.vue";
import RunTimeline from "../shared/ui/RunTimeline.vue";
import { formatDateTime } from "../shared/lib/datetime";
import { colorForRunStatus } from "../shared/lib/chips";
import { useRunDetailsStore } from "../features/runs/store";
import { useSnackbarStore } from "../shared/ui/feedback/snackbar-store";

const props = defineProps<{ runId: string }>();

const { t, locale } = useI18n({ useScope: "global" });
const details = useRunDetailsStore();
const router = useRouter();
const snackbar = useSnackbarStore();

const confirmDeleteNamespaceOpen = ref(false);
const canDeleteNamespace = computed(() => Boolean(details.run?.job_exists && details.run?.namespace));

async function loadAll() {
  await details.load(props.runId);
}

function goBack() {
  void router.push({ name: "runs" });
}

async function doDeleteNamespace() {
  await details.deleteNamespace(props.runId);
  if (!details.deleteNamespaceError) {
    snackbar.success(t("common.saved"));
  }
}

onMounted(() => void loadAll());
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
.pre {
  margin: 0;
  white-space: pre-wrap;
  font-size: 12px;
  opacity: 0.95;
}
</style>
