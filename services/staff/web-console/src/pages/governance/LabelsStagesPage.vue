<template>
  <div>
    <TableScaffoldPage
      titleKey="pages.labelsStages.title"
      hintKey="pages.labelsStages.hint"
      :headers="headers"
      :items="rows"
    >
      <template #below-header>
        <DismissibleWarningAlert
          alert-id="labels_stages"
          :title="t('warnings.labelsStages.title')"
          :text="t('warnings.labelsStages.text')"
        />

        <VAlert
          v-if="intentErrorKey"
          class="mt-4"
          type="warning"
          variant="tonal"
        >
          {{ t(intentErrorKey) }}
        </VAlert>

        <VCard
          v-else-if="transitionIntent"
          class="mt-4"
          variant="outlined"
        >
          <VCardTitle class="text-subtitle-1">
            {{ t("pages.labelsStages.transitionCardTitle") }}
          </VCardTitle>
          <VCardText>
            <div class="text-body-2 text-medium-emphasis">
              {{ t("pages.labelsStages.transitionCardHint") }}
            </div>
            <div class="mt-4 d-flex flex-column ga-2">
              <div class="text-body-2">
                <strong>{{ t("pages.labelsStages.repository") }}:</strong>
                <span class="mono">{{ transitionIntent.repositoryFullName }}</span>
              </div>
              <div class="text-body-2">
                <strong>{{ t("pages.labelsStages.issue") }}:</strong>
                <a
                  class="text-primary font-weight-bold text-decoration-none mono"
                  :href="issueURL"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  #{{ transitionIntent.issueNumber }}
                </a>
              </div>
              <div class="text-body-2">
                <strong>{{ t("pages.labelsStages.targetLabel") }}:</strong>
                <span class="mono">{{ transitionIntent.targetLabel }}</span>
              </div>
            </div>

            <VAlert v-if="applyErrorKey" class="mt-4" type="error" variant="tonal">
              {{ t(applyErrorKey) }}
            </VAlert>

            <VAlert
              v-if="transitionResult"
              class="mt-4"
              type="success"
              variant="tonal"
            >
              <div class="text-body-2">{{ t("pages.labelsStages.transitionApplied") }}</div>
              <div class="mt-2 text-body-2">
                <strong>{{ t("pages.labelsStages.removedLabels") }}:</strong>
                <span class="mono">{{ formatLabels(transitionResult.removed_labels) }}</span>
              </div>
              <div class="mt-1 text-body-2">
                <strong>{{ t("pages.labelsStages.addedLabels") }}:</strong>
                <span class="mono">{{ formatLabels(transitionResult.added_labels) }}</span>
              </div>
              <div class="mt-1 text-body-2">
                <strong>{{ t("pages.labelsStages.finalLabels") }}:</strong>
                <span class="mono">{{ formatLabels(transitionResult.final_labels) }}</span>
              </div>
            </VAlert>
          </VCardText>
          <VCardActions class="justify-end">
            <AdaptiveBtn
              variant="tonal"
              icon="mdi-open-in-new"
              :label="t('pages.labelsStages.openIssue')"
              :href="issueURL"
              target="_blank"
              rel="noopener noreferrer"
            />
            <AdaptiveBtn
              color="primary"
              variant="tonal"
              icon="mdi-check-bold"
              :label="t('pages.labelsStages.applyTransition')"
              :loading="applyLoading"
              @click="confirmOpen = true"
            />
          </VCardActions>
        </VCard>
      </template>
    </TableScaffoldPage>

    <ConfirmDialog
      v-model="confirmOpen"
      :title="t('pages.labelsStages.confirmTitle')"
      :message="confirmMessage"
      :confirm-text="t('pages.labelsStages.applyTransition')"
      :cancel-text="t('common.cancel')"
      @confirm="applyTransition"
    />
  </div>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальный stage/label policy (OpenAPI контракт + store) и режимы редактирования с аудитом.
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import DismissibleWarningAlert from "../../shared/ui/DismissibleWarningAlert.vue";
import { useI18n } from "vue-i18n";
import type { LocationQuery, LocationQueryValue } from "vue-router";
import ConfirmDialog from "../../shared/ui/ConfirmDialog.vue";
import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import { normalizeApiError } from "../../shared/api/errors";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { transitionIssueStageLabel } from "../../features/governance/stage-transition/api";
import type { TransitionIssueStageLabelResponse } from "../../shared/api/generated";

type PolicyRow = {
  kind: "label" | "stage";
  key: string;
  description: string;
  status: "active" | "planned";
};

const { t } = useI18n({ useScope: "global" });
const route = useRoute();
const snackbar = useSnackbarStore();

type TransitionIntent = {
  repositoryFullName: string;
  issueNumber: number;
  targetLabel: string;
  issueURL: string;
};

const confirmOpen = ref(false);
const applyLoading = ref(false);
const applyErrorKey = ref("");
const transitionResult = ref<TransitionIssueStageLabelResponse | null>(null);

const headers = [
  { key: "kind", width: 120 },
  { key: "key", width: 220 },
  { key: "description" },
  { key: "status", width: 140 },
  { key: "actions", sortable: false, width: 48 },
] as const;

const rows: PolicyRow[] = [
  { kind: "stage", key: "intake", description: "Issue intake and validation", status: "active" },
  { kind: "stage", key: "plan", description: "Work planning and decomposition", status: "active" },
  { kind: "stage", key: "impl", description: "Implementation (agent-run)", status: "active" },
  { kind: "stage", key: "review", description: "PR review / owner review", status: "active" },
  { kind: "stage", key: "ops", description: "Apply to cluster / smoke checks", status: "planned" },
  { kind: "label", key: "run:dev", description: "Run dev flow", status: "active" },
  { kind: "label", key: "need:owner_review", description: "Waiting for owner review", status: "active" },
  { kind: "label", key: "need:mcp_approval", description: "Waiting for MCP approval", status: "active" },
  { kind: "label", key: "state:blocked", description: "Execution blocked", status: "planned" },
];

const transitionIntentState = computed(() => parseTransitionIntent(route.query));
const transitionIntent = computed(() => transitionIntentState.value.intent);
const intentErrorKey = computed(() => transitionIntentState.value.errorKey);

const issueURL = computed(() => {
  const intent = transitionIntent.value;
  if (!intent) return "";
  return intent.issueURL || `https://github.com/${intent.repositoryFullName}/issues/${intent.issueNumber}`;
});

const confirmMessage = computed(() => {
  const intent = transitionIntent.value;
  if (!intent) {
    return "";
  }
  return t("pages.labelsStages.confirmMessage", {
    issue: `#${intent.issueNumber}`,
    label: intent.targetLabel,
  });
});

async function applyTransition(): Promise<void> {
  const intent = transitionIntent.value;
  if (!intent) return;

  applyLoading.value = true;
  applyErrorKey.value = "";
  try {
    transitionResult.value = await transitionIssueStageLabel({
      repositoryFullName: intent.repositoryFullName,
      issueNumber: intent.issueNumber,
      targetLabel: intent.targetLabel,
    });
    snackbar.success(t("pages.labelsStages.transitionApplied"));
  } catch (error) {
    const normalized = normalizeApiError(error);
    applyErrorKey.value = normalized.messageKey;
    snackbar.error(t(normalized.messageKey));
  } finally {
    applyLoading.value = false;
  }
}

function parseTransitionIntent(query: LocationQuery): { intent: TransitionIntent | null; errorKey: string } {
  const repositoryFullName = readQueryString(query.repo);
  const issueRaw = readQueryString(query.issue);
  const targetLabel = readQueryString(query.target);
  const issueURL = readQueryString(query.issue_url);

  if (!repositoryFullName && !issueRaw && !targetLabel) {
    return { intent: null, errorKey: "" };
  }
  if (!repositoryFullName || !issueRaw || !targetLabel) {
    return { intent: null, errorKey: "pages.labelsStages.transitionInvalidQuery" };
  }
  if (!repositoryFullName.includes("/") || repositoryFullName.startsWith("/") || repositoryFullName.endsWith("/")) {
    return { intent: null, errorKey: "pages.labelsStages.transitionInvalidQuery" };
  }

  const issueNumber = Number.parseInt(issueRaw, 10);
  if (!Number.isFinite(issueNumber) || issueNumber <= 0) {
    return { intent: null, errorKey: "pages.labelsStages.transitionInvalidQuery" };
  }
  if (!targetLabel.startsWith("run:")) {
    return { intent: null, errorKey: "pages.labelsStages.transitionInvalidQuery" };
  }

  return {
    intent: {
      repositoryFullName,
      issueNumber,
      targetLabel,
      issueURL,
    },
    errorKey: "",
  };
}

function readQueryString(value: LocationQueryValue | LocationQueryValue[] | undefined): string {
  if (Array.isArray(value)) {
    return (value[0] || "").trim();
  }
  return (value || "").trim();
}

function formatLabels(labels: string[]): string {
  if (!labels.length) return "-";
  return labels.join(", ");
}
</script>
