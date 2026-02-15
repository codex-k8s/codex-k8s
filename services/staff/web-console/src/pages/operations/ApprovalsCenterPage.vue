<template>
  <div>
    <PageHeader :title="t('pages.approvals.title')" :hint="t('pages.approvals.hint')">
      <template #actions>
        <VBtn
          variant="tonal"
          icon="mdi-refresh"
          :title="t('common.refresh')"
          :disabled="runs.approvalsLoading"
          @click="runs.loadPendingApprovals()"
        />
      </template>
    </PageHeader>

    <VAlert v-if="runs.approvalsError" type="error" variant="tonal" class="mt-4">
      {{ t(runs.approvalsError.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VRow density="compact">
          <VCol cols="12" md="4">
            <VTextField
              v-model.trim="toolFilter"
              :label="t('pages.approvals.filters.tool')"
              prepend-inner-icon="mdi-wrench-outline"
              hide-details
              clearable
            />
          </VCol>
          <VCol cols="12" md="4">
            <VTextField
              v-model.trim="actionFilter"
              :label="t('pages.approvals.filters.action')"
              prepend-inner-icon="mdi-flash-outline"
              hide-details
              clearable
            />
          </VCol>
          <VCol cols="12" md="4">
            <VTextField
              v-model.trim="requesterFilter"
              :label="t('pages.approvals.filters.requestedBy')"
              prepend-inner-icon="mdi-account-outline"
              hide-details
              clearable
            />
          </VCol>
        </VRow>
      </VCardText>
    </VCard>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable
          :headers="headers"
          :items="filtered"
          :loading="runs.approvalsLoading"
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

          <template #item.run="{ item }">
            <RouterLink
              v-if="item.run_id"
              class="text-primary font-weight-bold text-decoration-none"
              :to="{ name: 'run-details', params: { runId: item.run_id } }"
            >
              {{ item.run_id }}
            </RouterLink>
            <span v-else class="text-medium-emphasis">-</span>
          </template>

          <template #item.created_at="{ item }">
            <span class="text-medium-emphasis">{{ formatDateTime(item.created_at, locale) }}</span>
          </template>

          <template #item.actions="{ item }">
            <div class="d-flex ga-2 justify-end flex-wrap">
              <VTooltip :text="t('pages.approvals.approve')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    variant="tonal"
                    color="success"
                    icon="mdi-check"
                    :disabled="runs.resolvingApprovalID === item.id"
                    @click="openDecision(item.id, 'approved')"
                  />
                </template>
              </VTooltip>
              <VTooltip :text="t('pages.approvals.deny')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    variant="tonal"
                    color="error"
                    icon="mdi-close"
                    :disabled="runs.resolvingApprovalID === item.id"
                    @click="openDecision(item.id, 'denied')"
                  />
                </template>
              </VTooltip>
            </div>
          </template>

          <template #no-data>
            <div class="py-8 text-medium-emphasis">
              {{ t("states.noPendingApprovals") }}
            </div>
          </template>
        </VDataTable>
      </VCardText>
    </VCard>

    <VDialog v-model="decisionDialogOpen" max-width="560">
      <VCard>
        <VCardTitle class="text-subtitle-1">{{ decisionTitle }}</VCardTitle>
        <VCardText>
          <VTextarea v-model="decisionReason" :label="t('pages.approvals.reason')" auto-grow rows="4" />
        </VCardText>
        <VCardActions>
          <VSpacer />
          <VBtn variant="text" @click="decisionDialogOpen = false">{{ t("common.cancel") }}</VBtn>
          <VBtn variant="tonal" :color="decisionColor" @click="submitDecision" :disabled="decisionId === null">
            {{ t("common.save") }}
          </VBtn>
        </VCardActions>
      </VCard>
    </VDialog>
  </div>
</template>

<script setup lang="ts">
// TODO(#19): Доработать approvals center: фильтры по tool/action/state, история решений, и единый feedback слой (VSnackbar).
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import PageHeader from "../../shared/ui/PageHeader.vue";
import { formatDateTime } from "../../shared/lib/datetime";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { useRunsStore } from "../../features/runs/store";

const runs = useRunsStore();
const snackbar = useSnackbarStore();
const { t, locale } = useI18n({ useScope: "global" });

const toolFilter = ref("");
const actionFilter = ref("");
const requesterFilter = ref("");

const filtered = computed(() => {
  const tool = toolFilter.value.trim().toLowerCase();
  const action = actionFilter.value.trim().toLowerCase();
  const requester = requesterFilter.value.trim().toLowerCase();

  return runs.pendingApprovals.filter((a) => {
    if (tool && !a.tool_name.toLowerCase().includes(tool)) return false;
    if (action && !a.action.toLowerCase().includes(action)) return false;
    if (requester && !a.requested_by.toLowerCase().includes(requester)) return false;
    return true;
  });
});

const headers = [
  { title: t("table.fields.project"), key: "project", sortable: false, width: 220, align: "start" },
  { title: t("table.fields.run"), key: "run", sortable: false, width: 220, align: "center" },
  { title: t("table.fields.tool_name"), key: "tool_name", width: 180, align: "center" },
  { title: t("table.fields.action"), key: "action", align: "center" },
  { title: t("table.fields.requested_by"), key: "requested_by", width: 180, align: "center" },
  { title: t("table.fields.created_at"), key: "created_at", width: 180, align: "center" },
  { title: "", key: "actions", sortable: false, width: 120, align: "end" },
] as const;

const decisionDialogOpen = ref(false);
const decisionId = ref<number | null>(null);
const decision = ref<"approved" | "denied">("approved");
const decisionReason = ref("");

const decisionTitle = computed(() => (decision.value === "approved" ? t("pages.approvals.approveTitle") : t("pages.approvals.denyTitle")));
const decisionColor = computed(() => (decision.value === "approved" ? "success" : "error"));

function openDecision(id: number, next: "approved" | "denied") {
  decisionId.value = id;
  decision.value = next;
  decisionReason.value = "";
  decisionDialogOpen.value = true;
}

async function submitDecision() {
  if (decisionId.value === null) return;
  const id = decisionId.value;
  decisionDialogOpen.value = false;
  decisionId.value = null;
  const resp = await runs.resolvePendingApproval(id, decision.value, decisionReason.value);
  if (resp) {
    snackbar.success(t("common.saved"));
  }
}

onMounted(() => void runs.loadPendingApprovals());
</script>
