<template>
  <VDialog :model-value="open" max-width="760" @update:model-value="onDialogModelUpdate">
    <VCard>
      <VCardTitle class="d-flex align-center ga-2">
        <VIcon :icon="dialogIcon" size="22" />
        <span>{{ dialogTitle }}</span>
      </VCardTitle>

      <VCardText v-if="loading" class="py-8 d-flex align-center ga-3">
        <VProgressCircular indeterminate size="22" width="2" color="primary" />
        <span>{{ t("nextStep.loading") }}</span>
      </VCardText>

      <template v-else-if="preview && activeQuery">
        <VCardText class="d-flex flex-column ga-4">
          <VAlert type="info" variant="tonal" density="comfortable">
            <div class="font-weight-medium">{{ dialogTitle }}</div>
            <div class="text-body-2 mt-1">{{ dialogDescription }}</div>
          </VAlert>

          <div class="next-step-grid">
            <div class="next-step-field">
              <div class="next-step-label">{{ t("nextStep.fields.repository") }}</div>
              <div class="next-step-value">{{ activeQuery.repositoryFullName }}</div>
            </div>

            <div class="next-step-field">
              <div class="next-step-label">{{ t("nextStep.fields.thread") }}</div>
              <div class="next-step-value">
                <a v-if="preview.thread_url" :href="preview.thread_url" target="_blank" rel="noreferrer">
                  {{ threadTitle }}
                </a>
                <span v-else>{{ threadTitle }}</span>
              </div>
            </div>

            <div class="next-step-field">
              <div class="next-step-label">{{ t("nextStep.fields.actionKind") }}</div>
              <div class="next-step-value">{{ t(`nextStep.actionKinds.${activeQuery.actionKind}`) }}</div>
            </div>

            <div class="next-step-field">
              <div class="next-step-label">{{ t("nextStep.fields.targetLabel") }}</div>
              <div class="next-step-value"><code>{{ activeQuery.targetLabel }}</code></div>
            </div>
          </div>

          <div class="d-flex flex-column ga-3">
            <div>
              <div class="next-step-label mb-2">{{ t("nextStep.labels.removed") }}</div>
              <div class="d-flex flex-wrap ga-2">
                <VChip v-for="label in preview.removed_labels ?? []" :key="`removed:${label}`" color="error" size="small" variant="tonal">
                  {{ label }}
                </VChip>
                <span v-if="!(preview.removed_labels?.length)" class="text-medium-emphasis">{{ t("nextStep.labels.empty") }}</span>
              </div>
            </div>

            <div>
              <div class="next-step-label mb-2">{{ t("nextStep.labels.added") }}</div>
              <div class="d-flex flex-wrap ga-2">
                <VChip v-for="label in preview.added_labels ?? []" :key="`added:${label}`" color="success" size="small" variant="tonal">
                  {{ label }}
                </VChip>
                <span v-if="!(preview.added_labels?.length)" class="text-medium-emphasis">{{ t("nextStep.labels.empty") }}</span>
              </div>
            </div>

            <div>
              <div class="next-step-label mb-2">{{ t("nextStep.labels.final") }}</div>
              <div class="d-flex flex-wrap ga-2">
                <VChip v-for="label in preview.final_labels ?? []" :key="`final:${label}`" color="primary" size="small" variant="outlined">
                  {{ label }}
                </VChip>
                <span v-if="!(preview.final_labels?.length)" class="text-medium-emphasis">{{ t("nextStep.labels.empty") }}</span>
              </div>
            </div>
          </div>
        </VCardText>

        <VCardActions>
          <VSpacer />
          <VBtn variant="text" :disabled="submitting" @click="closeDialog">
            {{ t("common.cancel") }}
          </VBtn>
          <VBtn color="primary" variant="tonal" :loading="submitting" @click="confirmAction">
            {{ t("nextStep.confirm") }}
          </VBtn>
        </VCardActions>
      </template>
    </VCard>
  </VDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";

import { normalizeApiError } from "../../shared/api/errors";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { useAuthStore } from "../auth/store";
import { executeNextStepAction, previewNextStepAction } from "./api";
import type { NextStepActionDisplayVariant, NextStepActionKind, NextStepActionPreview, NextStepActionQuery } from "./types";

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();
const snackbar = useSnackbarStore();
const { t } = useI18n({ useScope: "global" });

const open = ref(false);
const loading = ref(false);
const submitting = ref(false);
const preview = ref<NextStepActionPreview | null>(null);
const activeQuery = ref<NextStepActionQuery | null>(null);

let syncCounter = 0;

const supportedActionKinds: NextStepActionKind[] = ["issue_stage_transition", "pull_request_label_add"];
const supportedDisplayVariants: NextStepActionDisplayVariant[] = [
  "revise",
  "full_flow",
  "shortened_flow",
  "very_short_flow",
  "full_or_shortened_flow",
  "full_or_very_short_flow",
  "shortened_or_very_short_flow",
  "all_flows",
  "reviewer",
  "rethink",
  "doc_audit",
  "self_improve",
  "prepare_plan",
  "go_to_dev",
  "go_to_qa",
  "restart_full",
  "restart_shortened",
  "restart_very_short",
];

const dialogVariant = computed<NextStepActionDisplayVariant>(() => activeQuery.value?.displayVariant ?? "revise");
const dialogTitle = computed(() => t(`nextStep.variants.${dialogVariant.value}.title`));
const dialogDescription = computed(() => t(`nextStep.variants.${dialogVariant.value}.description`));
const dialogIcon = computed(() => {
  switch (dialogVariant.value) {
    case "reviewer":
      return "mdi-magnify";
    case "rethink":
      return "mdi-refresh-circle";
    case "doc_audit":
      return "mdi-book-search-outline";
    case "self_improve":
      return "mdi-brain";
    case "prepare_plan":
      return "mdi-file-document-edit-outline";
    case "go_to_dev":
      return "mdi-hammer-wrench";
    case "go_to_qa":
      return "mdi-test-tube";
    case "revise":
      return "mdi-file-edit-outline";
    default:
      return "mdi-rocket-launch-outline";
  }
});

const threadTitle = computed(() => {
  if (!preview.value) return "";
  const threadKey = preview.value.thread_kind === "pull_request" ? "nextStep.thread.pullRequest" : "nextStep.thread.issue";
  return `${t(threadKey)} #${preview.value.thread_number}`;
});

watch(
  [() => route.query, () => auth.status],
  () => {
    void syncFromRoute();
  },
  { deep: true, immediate: true },
);

function normalizeQueryString(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function parsePositiveInt(value: unknown): number | undefined {
  if (typeof value !== "string") return undefined;
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : undefined;
}

function parseRouteQuery(): NextStepActionQuery | null {
  if (normalizeQueryString(route.query.modal) !== "next-step") {
    return null;
  }

  const repositoryFullName = normalizeQueryString(route.query.repository_full_name);
  const actionKind = normalizeQueryString(route.query.action_kind) as NextStepActionKind;
  const targetLabel = normalizeQueryString(route.query.target_label);
  const displayVariant = normalizeQueryString(route.query.display_variant) as NextStepActionDisplayVariant;
  const issueNumber = parsePositiveInt(route.query.issue_number);
  const pullRequestNumber = parsePositiveInt(route.query.pull_request_number);

  if (repositoryFullName === "" || targetLabel === "") {
    return null;
  }
  if (!supportedActionKinds.includes(actionKind)) {
    return null;
  }
  if (!supportedDisplayVariants.includes(displayVariant)) {
    return null;
  }
  if (actionKind === "issue_stage_transition" && typeof issueNumber !== "number") {
    return null;
  }
  if (actionKind === "pull_request_label_add" && typeof pullRequestNumber !== "number") {
    return null;
  }

  return {
    repositoryFullName,
    issueNumber,
    pullRequestNumber,
    actionKind,
    targetLabel,
    displayVariant,
  };
}

async function clearModalQuery(): Promise<void> {
  const nextQuery = { ...route.query };
  delete nextQuery.modal;
  delete nextQuery.repository_full_name;
  delete nextQuery.issue_number;
  delete nextQuery.pull_request_number;
  delete nextQuery.action_kind;
  delete nextQuery.target_label;
  delete nextQuery.display_variant;
  await router.replace({ query: nextQuery });
}

function resetState(): void {
  open.value = false;
  loading.value = false;
  submitting.value = false;
  preview.value = null;
  activeQuery.value = null;
}

async function syncFromRoute(): Promise<void> {
  const parsed = parseRouteQuery();
  if (!parsed) {
    resetState();
    return;
  }
  if (auth.status !== "authed") {
    return;
  }

  const queryKey = JSON.stringify(parsed);
  if (JSON.stringify(activeQuery.value) === queryKey && preview.value) {
    open.value = true;
    return;
  }

  const requestId = ++syncCounter;
  open.value = true;
  loading.value = true;
  preview.value = null;
  activeQuery.value = parsed;
  try {
    const resp = await previewNextStepAction(parsed);
    if (requestId !== syncCounter) return;
    preview.value = resp;
  } catch (error) {
    if (requestId !== syncCounter) return;
    snackbar.error(t(normalizeApiError(error).messageKey));
    await clearModalQuery();
  } finally {
    if (requestId === syncCounter) {
      loading.value = false;
    }
  }
}

async function closeDialog(): Promise<void> {
  open.value = false;
  await clearModalQuery();
}

async function confirmAction(): Promise<void> {
  if (!activeQuery.value || submitting.value) return;

  submitting.value = true;
  try {
    preview.value = await executeNextStepAction(activeQuery.value);
    snackbar.success(t("nextStep.executeSuccess"));
    await closeDialog();
  } catch (error) {
    snackbar.error(t(normalizeApiError(error).messageKey));
  } finally {
    submitting.value = false;
  }
}

function onDialogModelUpdate(nextValue: boolean): void {
  if (nextValue) return;
  void closeDialog();
}
</script>

<style scoped>
.next-step-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.next-step-field {
  min-width: 0;
}

.next-step-label {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.02em;
  opacity: 0.7;
  text-transform: uppercase;
}

.next-step-value {
  margin-top: 4px;
  line-height: 1.4;
  word-break: break-word;
}

@media (max-width: 760px) {
  .next-step-grid {
    grid-template-columns: 1fr;
  }
}
</style>
