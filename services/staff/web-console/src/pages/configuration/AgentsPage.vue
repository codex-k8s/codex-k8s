<template>
  <div>
    <PageHeader :title="t('pages.agents.title')" :hint="t('pages.agents.hint')">
      <template #actions>
        <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :loading="loading" @click="loadAgents" />
      </template>
    </PageHeader>

    <VAlert v-if="error" type="error" variant="tonal" class="mt-4">
      {{ t(error.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable :headers="headers" :items="items" :loading="loading" :items-per-page="12" hover>
          <template #item.name="{ item }">
            <div class="d-flex justify-center">
              <RouterLink class="text-primary font-weight-bold text-decoration-none" :to="{ name: 'agent-details', params: { agentId: item.id } }">
                {{ item.name }}
              </RouterLink>
            </div>
          </template>

          <template #item.role_kind="{ item }">
            <div class="d-flex justify-center">
              <VChip size="small" variant="tonal" :color="item.role_kind === 'system' ? 'primary' : 'secondary'">
                {{ item.role_kind }}
              </VChip>
            </div>
          </template>

          <template #item.project_id="{ item }">
            <span class="mono">{{ item.project_id || "-" }}</span>
          </template>

          <template #item.runtime_mode="{ item }">
            <div class="d-flex justify-center">
              <VChip size="small" variant="tonal" :color="runtimeModeColor(item.settings.runtime_mode)">
                {{ item.settings.runtime_mode }}
              </VChip>
            </div>
          </template>

          <template #item.timeout_seconds="{ item }">
            <span class="mono">{{ item.settings.timeout_seconds }}s</span>
          </template>

          <template #item.max_retry_count="{ item }">
            <span class="mono">{{ item.settings.max_retry_count }}</span>
          </template>

          <template #item.prompt_locale="{ item }">
            <span class="mono">{{ item.settings.prompt_locale }}</span>
          </template>

          <template #item.approvals_required="{ item }">
            <div class="d-flex justify-center">
              <VChip size="small" variant="tonal" :color="item.settings.approvals_required ? 'warning' : 'success'">
                {{ item.settings.approvals_required ? t("pages.agents.approvals.required") : t("pages.agents.approvals.notRequired") }}
              </VChip>
            </div>
          </template>

          <template #item.status="{ item }">
            <div class="d-flex justify-center">
              <VChip size="small" variant="tonal" :color="item.is_active ? 'success' : 'default'">
                {{ item.is_active ? t("pages.agents.status.active") : t("pages.agents.status.inactive") }}
              </VChip>
            </div>
          </template>

          <template #item.actions="{ item }">
            <div class="d-flex justify-end">
              <VTooltip :text="t('scaffold.rowActions.view')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    variant="text"
                    icon="mdi-open-in-new"
                    :to="{ name: 'agent-details', params: { agentId: item.id } }"
                  />
                </template>
              </VTooltip>
            </div>
          </template>

          <template #no-data>
            <div class="py-8 text-medium-emphasis">
              {{ t("pages.agents.noData") }}
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

import { listAgents } from "../../features/agents/api";
import type { Agent } from "../../features/agents/types";
import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import PageHeader from "../../shared/ui/PageHeader.vue";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";

const { t } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();

const loading = ref(false);
const error = ref<ApiError | null>(null);
const items = ref<Agent[]>([]);

const headers = computed(() => ([
  { title: t("pages.agents.columns.name"), key: "name", align: "start", width: 220 },
  { title: t("pages.agents.columns.roleKind"), key: "role_kind", align: "center", width: 120, sortable: false },
  { title: t("pages.agents.columns.projectId"), key: "project_id", align: "center", width: 220 },
  { title: t("pages.agents.columns.runtimeMode"), key: "runtime_mode", align: "center", width: 140, sortable: false },
  { title: t("pages.agents.columns.timeoutSeconds"), key: "timeout_seconds", align: "center", width: 140, sortable: false },
  { title: t("pages.agents.columns.maxRetryCount"), key: "max_retry_count", align: "center", width: 120, sortable: false },
  { title: t("pages.agents.columns.promptLocale"), key: "prompt_locale", align: "center", width: 120, sortable: false },
  { title: t("pages.agents.columns.approvals"), key: "approvals_required", align: "center", width: 170, sortable: false },
  { title: t("pages.agents.columns.status"), key: "status", align: "center", width: 120, sortable: false },
  { title: "", key: "actions", align: "end", width: 72, sortable: false },
]) as const);

function runtimeModeColor(mode: string): string {
  return mode === "full-env" ? "primary" : "secondary";
}

async function loadAgents(): Promise<void> {
  loading.value = true;
  error.value = null;
  try {
    items.value = await listAgents(200);
  } catch (e) {
    const normalized = normalizeApiError(e);
    error.value = normalized;
    snackbar.error(t(normalized.messageKey));
  } finally {
    loading.value = false;
  }
}

onMounted(() => void loadAgents());
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>

