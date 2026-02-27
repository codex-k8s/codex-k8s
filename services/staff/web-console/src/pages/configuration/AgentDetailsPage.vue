<template>
  <div>
    <PageHeader :title="t('pages.agentDetails.title', { name: displayAgentName })" :hint="t('pages.agentDetails.hint')">
      <template #leading>
        <BackBtn :label="t('common.back')" :to="{ name: 'agents' }" />
      </template>
      <template #actions>
        <CopyChip :label="t('pages.agentDetails.agent')" :value="agentId" icon="mdi-robot-outline" />
        <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :loading="loadingAgent || loadingTemplateData" @click="refreshAll" />
      </template>
    </PageHeader>

    <VAlert v-if="error" type="error" variant="tonal" class="mt-4">
      {{ t(error.messageKey) }}
    </VAlert>

    <VTabs v-model="tab" class="mt-4">
      <VTab value="settings">{{ t("pages.agentDetails.tabs.settings") }}</VTab>
      <VTab value="templates">{{ t("pages.agentDetails.tabs.templates") }}</VTab>
      <VTab value="history">{{ t("pages.agentDetails.tabs.history") }}</VTab>
    </VTabs>

    <VWindow v-model="tab" class="mt-2">
      <VWindowItem value="settings">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-1 d-flex align-center justify-space-between ga-2 flex-wrap">
            <span>{{ t("pages.agentDetails.settings.title") }}</span>
            <VChip size="small" variant="tonal" color="secondary">
              {{ t("pages.agentDetails.settings.settingsVersion") }}: {{ settingsVersion }}
            </VChip>
          </VCardTitle>
          <VCardText>
            <VRow density="compact">
              <VCol cols="12" md="4">
                <VSelect
                  v-model="settingsForm.runtime_mode"
                  :label="t('pages.agentDetails.settings.runtimeMode')"
                  :items="runtimeModeItems"
                  hide-details
                />
              </VCol>
              <VCol cols="12" md="4">
                <VTextField
                  v-model.number="settingsForm.timeout_seconds"
                  type="number"
                  min="1"
                  :label="t('pages.agentDetails.settings.timeoutSeconds')"
                  hide-details
                />
              </VCol>
              <VCol cols="12" md="4">
                <VTextField
                  v-model.number="settingsForm.max_retry_count"
                  type="number"
                  min="0"
                  :label="t('pages.agentDetails.settings.maxRetryCount')"
                  hide-details
                />
              </VCol>
              <VCol cols="12" md="6">
                <VTextField v-model.trim="settingsForm.prompt_locale" :label="t('pages.agentDetails.settings.promptLocale')" hide-details />
              </VCol>
              <VCol cols="12" md="6">
                <VSwitch v-model="settingsForm.approvals_required" :label="t('pages.agentDetails.settings.approvalsRequired')" hide-details />
              </VCol>
            </VRow>
          </VCardText>
          <VCardActions>
            <VSpacer />
            <VBtn variant="text" :disabled="savingSettings" @click="resetSettings">{{ t("common.reset") }}</VBtn>
            <VBtn color="primary" variant="tonal" :loading="savingSettings" :disabled="!canSaveSettings" @click="saveSettings">
              {{ t("common.save") }}
            </VBtn>
          </VCardActions>
        </VCard>
      </VWindowItem>

      <VWindowItem value="templates">
        <VAlert v-if="templateError" type="error" variant="tonal" class="mb-4">
          {{ t(templateError.messageKey) }}
        </VAlert>

        <VRow density="compact">
          <VCol cols="12" md="4">
            <VCard variant="outlined">
              <VCardTitle class="text-subtitle-1">{{ t("pages.agentDetails.templates.controlsTitle") }}</VCardTitle>
              <VCardText>
                <VSelect
                  v-model="selectedTemplateKey"
                  :label="t('pages.agentDetails.templates.templateKey')"
                  :items="templateKeyOptions"
                  item-title="title"
                  item-value="value"
                  :loading="loadingTemplateData"
                  hide-details
                />

                <VAlert v-if="templateKeys.length === 0" type="info" variant="tonal" class="mt-4">
                  {{ t("pages.agentDetails.templates.noTemplates") }}
                </VAlert>

                <template v-else>
                  <VSelect
                    v-model="selectedVersion"
                    class="mt-4"
                    :label="t('pages.agentDetails.templates.selectedVersion')"
                    :items="versionOptions"
                    item-title="title"
                    item-value="value"
                    :loading="loadingTemplateData"
                    hide-details
                  />

                  <VSelect
                    v-model="diffFromVersion"
                    class="mt-4"
                    :label="t('pages.agentDetails.templates.diffFrom')"
                    :items="versionOptions"
                    item-title="title"
                    item-value="value"
                    hide-details
                  />

                  <VSelect
                    v-model="diffToVersion"
                    class="mt-4"
                    :label="t('pages.agentDetails.templates.diffTo')"
                    :items="versionOptions"
                    item-title="title"
                    item-value="value"
                    hide-details
                  />

                  <VTextField v-model.trim="changeReason" class="mt-4" :label="t('pages.agentDetails.templates.changeReason')" hide-details />
                </template>
              </VCardText>
            </VCard>
          </VCol>

          <VCol cols="12" md="8">
            <VCard variant="outlined">
              <VCardTitle class="text-subtitle-1 d-flex align-center justify-space-between ga-2 flex-wrap">
                <span>{{ t("pages.agentDetails.templates.editorTitle") }}</span>
                <div class="d-flex ga-2 flex-wrap">
                  <VChip size="small" variant="tonal" color="secondary">
                    {{ t("pages.agentDetails.templates.expectedVersion") }}: {{ expectedTemplateVersion }}
                  </VChip>
                  <VBtn variant="tonal" prepend-icon="mdi-refresh" :loading="loadingTemplateData" @click="refreshTemplateData">
                    {{ t("common.refresh") }}
                  </VBtn>
                  <VBtn
                    color="primary"
                    variant="tonal"
                    prepend-icon="mdi-content-save-outline"
                    :disabled="!canCreateTemplateVersion"
                    :loading="savingTemplateVersion"
                    @click="createTemplateVersion"
                  >
                    {{ t("pages.agentDetails.templates.createVersion") }}
                  </VBtn>
                  <VBtn
                    variant="tonal"
                    prepend-icon="mdi-check-circle-outline"
                    :disabled="!canActivateTemplateVersion"
                    :loading="activatingTemplateVersion"
                    @click="activateTemplateVersion"
                  >
                    {{ t("pages.agentDetails.templates.activateVersion") }}
                  </VBtn>
                </div>
              </VCardTitle>
              <VCardText>
                <MonacoEditor v-model="editorMarkdown" language="markdown" height="500px" />
              </VCardText>
            </VCard>
          </VCol>
        </VRow>

        <VRow class="mt-4" density="compact">
          <VCol cols="12" md="6">
            <VCard variant="outlined">
              <VCardTitle class="text-subtitle-2">{{ t("pages.agentDetails.templates.diffFromBody") }}</VCardTitle>
              <VCardText>
                <MonacoEditor v-model="diffFromBody" language="markdown" height="320px" read-only />
              </VCardText>
            </VCard>
          </VCol>
          <VCol cols="12" md="6">
            <VCard variant="outlined">
              <VCardTitle class="text-subtitle-2">{{ t("pages.agentDetails.templates.diffToBody") }}</VCardTitle>
              <VCardText>
                <MonacoEditor v-model="diffToBody" language="markdown" height="320px" read-only />
              </VCardText>
            </VCard>
          </VCol>
        </VRow>

        <VCard class="mt-4" variant="outlined">
          <VCardTitle class="text-subtitle-1">{{ t("pages.agentDetails.templates.previewTitle") }}</VCardTitle>
          <VCardText>
            <MonacoEditor v-model="previewBody" language="markdown" height="320px" read-only />
          </VCardText>
        </VCard>

        <VCard class="mt-4" variant="outlined">
          <VCardTitle class="text-subtitle-1">{{ t("pages.agentDetails.templates.versionsTitle") }}</VCardTitle>
          <VCardText>
            <VDataTable :headers="versionsHeaders" :items="versions" :items-per-page="8" density="compact" hover>
              <template #item.status="{ item }">
                <VChip size="x-small" variant="tonal" :color="versionStatusColor(item.status)">
                  {{ item.status }}
                </VChip>
              </template>
              <template #item.updated_at="{ item }">
                <span class="text-medium-emphasis">{{ fmtDateTime(item.updated_at) }}</span>
              </template>
              <template #item.actions="{ item }">
                <VBtn size="x-small" variant="text" icon="mdi-arrow-up-right-bold-box-outline" @click="selectedVersion = item.version" />
              </template>
              <template #no-data>
                <div class="py-4 text-medium-emphasis">
                  {{ t("pages.agentDetails.templates.noVersions") }}
                </div>
              </template>
            </VDataTable>
          </VCardText>
        </VCard>
      </VWindowItem>

      <VWindowItem value="history">
        <VAlert v-if="auditError" type="error" variant="tonal" class="mb-4">
          {{ t(auditError.messageKey) }}
        </VAlert>

        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-1 d-flex align-center justify-space-between ga-2 flex-wrap">
            <span>{{ t("pages.agentDetails.history.title") }}</span>
            <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :loading="loadingAudit" @click="loadAuditEvents" />
          </VCardTitle>
          <VCardText>
            <VDataTable :headers="auditHeaders" :items="auditEvents" :items-per-page="10" density="compact" hover>
              <template #item.created_at="{ item }">
                <span class="text-medium-emphasis">{{ fmtDateTime(item.created_at) }}</span>
              </template>
              <template #item.version="{ item }">
                <span class="mono">{{ item.version ?? "-" }}</span>
              </template>
              <template #item.actor_id="{ item }">
                <span class="mono">{{ item.actor_id || "-" }}</span>
              </template>
              <template #item.correlation_id="{ item }">
                <span class="mono">{{ item.correlation_id }}</span>
              </template>
              <template #no-data>
                <div class="py-8 text-medium-emphasis">
                  {{ t("pages.agentDetails.history.empty") }}
                </div>
              </template>
            </VDataTable>
          </VCardText>
        </VCard>
      </VWindowItem>
    </VWindow>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

import {
  activatePromptTemplateVersion as activatePromptTemplateVersionRequest,
  createPromptTemplateVersion,
  diffPromptTemplateVersions,
  getAgent as getAgentRequest,
  listPromptTemplateAuditEvents,
  listPromptTemplateKeys,
  listPromptTemplateVersions,
  previewPromptTemplate,
  updateAgentSettings as updateAgentSettingsRequest,
} from "../../features/agents/api";
import type {
  Agent,
  AgentSettings,
  PromptTemplateAuditEvent,
  PromptTemplateDiffResponse,
  PromptTemplateKey,
  PromptTemplateSource,
  PromptTemplateVersion,
  PreviewPromptTemplateResponse,
} from "../../features/agents/types";
import { useAuthStore } from "../../features/auth/store";
import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { formatDateTime } from "../../shared/lib/datetime";
import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import BackBtn from "../../shared/ui/BackBtn.vue";
import CopyChip from "../../shared/ui/CopyChip.vue";
import PageHeader from "../../shared/ui/PageHeader.vue";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import MonacoEditor from "../../shared/ui/monaco/MonacoEditor.vue";

type AgentTab = "settings" | "templates" | "history";

const props = defineProps<{ agentId: string }>();

const { t, locale } = useI18n({ useScope: "global" });
const auth = useAuthStore();
const snackbar = useSnackbarStore();

const tab = ref<AgentTab>("settings");

const loadingAgent = ref(false);
const loadingTemplateData = ref(false);
const loadingAudit = ref(false);
const savingSettings = ref(false);
const savingTemplateVersion = ref(false);
const activatingTemplateVersion = ref(false);

const error = ref<ApiError | null>(null);
const templateError = ref<ApiError | null>(null);
const auditError = ref<ApiError | null>(null);

const agent = ref<Agent | null>(null);
const settingsVersion = ref(1);
const settingsForm = ref<AgentSettings>({
  runtime_mode: "code-only",
  timeout_seconds: 3600,
  max_retry_count: 1,
  prompt_locale: "en",
  approvals_required: false,
});

const templateKeys = ref<PromptTemplateKey[]>([]);
const selectedTemplateKey = ref("");
const versions = ref<PromptTemplateVersion[]>([]);
const selectedVersion = ref<number | null>(null);
const editorMarkdown = ref("");
const changeReason = ref("");

const preview = ref<PreviewPromptTemplateResponse | null>(null);
const diff = ref<PromptTemplateDiffResponse | null>(null);
const diffFromVersion = ref<number | null>(null);
const diffToVersion = ref<number | null>(null);

const auditEvents = ref<PromptTemplateAuditEvent[]>([]);

const runtimeModeItems = computed(() => ([
  { title: "full-env", value: "full-env" },
  { title: "code-only", value: "code-only" },
] as const));

const displayAgentName = computed(() => agent.value?.name || props.agentId);
const agentId = computed(() => props.agentId);

const templateKeyOptions = computed(() =>
  templateKeys.value.map((item) => ({
    title: `${item.template_key} · v${item.active_version}`,
    value: item.template_key,
  })),
);

const versionOptions = computed(() =>
  versions.value.map((item) => ({
    title: `v${item.version} · ${item.status}`,
    value: item.version,
  })),
);

const expectedTemplateVersion = computed(() => versions.value[0]?.version ?? 0);

const canSaveSettings = computed(() => {
  const currentAgent = agent.value;
  if (!currentAgent) {
    return false;
  }
  if (!settingsForm.value.prompt_locale.trim()) {
    return false;
  }
  if (settingsForm.value.timeout_seconds <= 0) {
    return false;
  }
  if (settingsForm.value.max_retry_count < 0) {
    return false;
  }
  return JSON.stringify(currentAgent.settings) !== JSON.stringify(settingsForm.value);
});

const canCreateTemplateVersion = computed(() => {
  if (!selectedTemplateKey.value.trim()) {
    return false;
  }
  if (!editorMarkdown.value.trim()) {
    return false;
  }
  return expectedTemplateVersion.value > 0;
});

const canActivateTemplateVersion = computed(() => {
  if (!selectedTemplateKey.value.trim() || !selectedVersion.value) {
    return false;
  }
  if (!changeReason.value.trim()) {
    return false;
  }
  return expectedTemplateVersion.value > 0;
});

const previewBody = computed(() => preview.value?.body_markdown || "");
const diffFromBody = computed(() => diff.value?.from_body_markdown || "");
const diffToBody = computed(() => diff.value?.to_body_markdown || "");

const versionsHeaders = computed(() => ([
  { title: t("pages.agentDetails.templates.columns.version"), key: "version", align: "start", width: 90 },
  { title: t("pages.agentDetails.templates.columns.status"), key: "status", align: "center", width: 120 },
  { title: t("pages.agentDetails.templates.columns.source"), key: "source", align: "center", width: 150 },
  { title: t("pages.agentDetails.templates.columns.updatedBy"), key: "updated_by", align: "center", width: 160 },
  { title: t("pages.agentDetails.templates.columns.updatedAt"), key: "updated_at", align: "center", width: 180 },
  { title: "", key: "actions", align: "end", width: 70, sortable: false },
]) as const);

const auditHeaders = computed(() => ([
  { title: t("pages.agentDetails.history.columns.createdAt"), key: "created_at", align: "start", width: 180 },
  { title: t("pages.agentDetails.history.columns.eventType"), key: "event_type", align: "center", width: 220 },
  { title: t("pages.agentDetails.history.columns.version"), key: "version", align: "center", width: 90 },
  { title: t("pages.agentDetails.history.columns.actor"), key: "actor_id", align: "center", width: 180 },
  { title: t("pages.agentDetails.history.columns.correlation"), key: "correlation_id", align: "center", width: 220 },
]) as const);

function cloneSettings(settings: AgentSettings): AgentSettings {
  return {
    runtime_mode: settings.runtime_mode,
    timeout_seconds: settings.timeout_seconds,
    max_retry_count: settings.max_retry_count,
    prompt_locale: settings.prompt_locale,
    approvals_required: settings.approvals_required,
  };
}

function templateSourceByKey(templateKey: string): PromptTemplateSource {
  return templateKey.startsWith("project/") ? "project_override" : "global_override";
}

function parseProjectIDFromTemplateKey(templateKey: string): string {
  const value = templateKey.trim();
  if (!value.startsWith("project/")) {
    return "";
  }
  const parts = value.split("/");
  if (parts.length < 5) {
    return "";
  }
  return parts[1]?.trim() || "";
}

function normalizeProjectID(projectID: string | null | undefined): string {
  return projectID?.trim() || "";
}

function currentAgentProjectID(): string {
  return normalizeProjectID(agent.value?.project_id);
}

function selectDefaultTemplateKey(items: PromptTemplateKey[], projectID: string): string {
  if (items.length === 0) {
    return "";
  }
  if (!projectID) {
    return items[0]?.template_key || "";
  }
  const projectScoped = items.find((item) => normalizeProjectID(item.project_id) === projectID);
  if (projectScoped) {
    return projectScoped.template_key;
  }
  return items[0]?.template_key || "";
}

function resolveTemplateProjectIDContext(templateKey: string): string {
  const projectIDFromKey = parseProjectIDFromTemplateKey(templateKey);
  if (projectIDFromKey) {
    return projectIDFromKey;
  }
  return currentAgentProjectID();
}

function versionStatusColor(status: string): string {
  if (status === "active") {
    return "success";
  }
  if (status === "draft") {
    return "warning";
  }
  return "default";
}

function fmtDateTime(value: string | null | undefined): string {
  return formatDateTime(value, locale.value);
}

function resetSettings(): void {
  if (!agent.value) {
    return;
  }
  settingsForm.value = cloneSettings(agent.value.settings);
}

async function loadAgent(): Promise<void> {
  loadingAgent.value = true;
  error.value = null;
  try {
    const item = await getAgentRequest(props.agentId);
    agent.value = item;
    settingsForm.value = cloneSettings(item.settings);
    settingsVersion.value = item.settings_version;
  } catch (e) {
    const normalized = normalizeApiError(e);
    error.value = normalized;
    snackbar.error(t(normalized.messageKey));
  } finally {
    loadingAgent.value = false;
  }
}

async function loadTemplateKeys(): Promise<void> {
  if (!agent.value) {
    templateKeys.value = [];
    selectedTemplateKey.value = "";
    return;
  }

  loadingTemplateData.value = true;
  templateError.value = null;
  try {
    const projectID = currentAgentProjectID();
    const items = await listPromptTemplateKeys({
      role: agent.value.agent_key,
      scope: projectID ? undefined : "global",
      projectId: projectID || undefined,
      limit: 500,
    });
    templateKeys.value = [...items].sort((left, right) => left.template_key.localeCompare(right.template_key));
    if (templateKeys.value.length === 0) {
      selectedTemplateKey.value = "";
      versions.value = [];
      selectedVersion.value = null;
      diff.value = null;
      preview.value = null;
      return;
    }

    const hasSelectedKey = templateKeys.value.some((item) => item.template_key === selectedTemplateKey.value);
    if (!hasSelectedKey) {
      selectedTemplateKey.value = selectDefaultTemplateKey(templateKeys.value, projectID);
    } else {
      await loadTemplateData();
    }
  } catch (e) {
    const normalized = normalizeApiError(e);
    templateError.value = normalized;
    snackbar.error(t(normalized.messageKey));
  } finally {
    loadingTemplateData.value = false;
  }
}

async function loadPreview(): Promise<void> {
  if (!selectedTemplateKey.value.trim()) {
    preview.value = null;
    return;
  }
  const projectID = resolveTemplateProjectIDContext(selectedTemplateKey.value);
  preview.value = await previewPromptTemplate({
    templateKey: selectedTemplateKey.value,
    projectId: projectID || undefined,
  });
}

async function loadDiff(): Promise<void> {
  if (!selectedTemplateKey.value.trim() || !diffFromVersion.value || !diffToVersion.value) {
    diff.value = null;
    return;
  }
  if (diffFromVersion.value === diffToVersion.value) {
    diff.value = null;
    return;
  }
  diff.value = await diffPromptTemplateVersions({
    templateKey: selectedTemplateKey.value,
    fromVersion: diffFromVersion.value,
    toVersion: diffToVersion.value,
  });
}

async function loadTemplateData(): Promise<void> {
  if (!selectedTemplateKey.value.trim()) {
    versions.value = [];
    selectedVersion.value = null;
    diff.value = null;
    preview.value = null;
    return;
  }

  loadingTemplateData.value = true;
  templateError.value = null;
  try {
    const loadedVersions = await listPromptTemplateVersions(selectedTemplateKey.value, 200);
    versions.value = loadedVersions;

    const hasSelectedVersion = loadedVersions.some((item) => item.version === selectedVersion.value);
    if (!hasSelectedVersion) {
      selectedVersion.value = loadedVersions[0]?.version ?? null;
    }
    if (selectedVersion.value) {
      const selected = loadedVersions.find((item) => item.version === selectedVersion.value);
      editorMarkdown.value = selected?.body_markdown || "";
    } else {
      editorMarkdown.value = "";
    }

    if (loadedVersions.length >= 2) {
      diffToVersion.value = loadedVersions[0]?.version ?? null;
      diffFromVersion.value = loadedVersions[1]?.version ?? null;
    } else {
      diffToVersion.value = null;
      diffFromVersion.value = null;
    }

    await loadPreview();
    await loadDiff();
  } catch (e) {
    const normalized = normalizeApiError(e);
    templateError.value = normalized;
    snackbar.error(t(normalized.messageKey));
  } finally {
    loadingTemplateData.value = false;
  }
}

async function saveSettings(): Promise<void> {
  if (!agent.value) {
    return;
  }

  savingSettings.value = true;
  error.value = null;
  try {
    const updated = await updateAgentSettingsRequest({
      agentId: agent.value.id,
      expectedVersion: settingsVersion.value,
      settings: {
        runtime_mode: settingsForm.value.runtime_mode,
        timeout_seconds: Number(settingsForm.value.timeout_seconds),
        max_retry_count: Number(settingsForm.value.max_retry_count),
        prompt_locale: settingsForm.value.prompt_locale.trim(),
        approvals_required: settingsForm.value.approvals_required,
      },
    });
    agent.value = updated;
    settingsForm.value = cloneSettings(updated.settings);
    settingsVersion.value = updated.settings_version;
    snackbar.success(t("common.saved"));
  } catch (e) {
    const normalized = normalizeApiError(e);
    error.value = normalized;
    snackbar.error(t(normalized.messageKey));
  } finally {
    savingSettings.value = false;
  }
}

async function createTemplateVersion(): Promise<void> {
  if (!canCreateTemplateVersion.value) {
    return;
  }

  savingTemplateVersion.value = true;
  templateError.value = null;
  try {
    const created = await createPromptTemplateVersion({
      templateKey: selectedTemplateKey.value,
      bodyMarkdown: editorMarkdown.value,
      expectedVersion: expectedTemplateVersion.value,
      source: templateSourceByKey(selectedTemplateKey.value),
      changeReason: changeReason.value,
    });

    selectedVersion.value = created.version;
    await loadTemplateData();
    await loadAuditEvents();
    snackbar.success(t("common.saved"));
  } catch (e) {
    const normalized = normalizeApiError(e);
    templateError.value = normalized;
    snackbar.error(t(normalized.messageKey));
  } finally {
    savingTemplateVersion.value = false;
  }
}

async function activateTemplateVersion(): Promise<void> {
  if (!canActivateTemplateVersion.value || !selectedVersion.value) {
    return;
  }

  activatingTemplateVersion.value = true;
  templateError.value = null;
  try {
    await activatePromptTemplateVersionRequest({
      templateKey: selectedTemplateKey.value,
      version: selectedVersion.value,
      expectedVersion: expectedTemplateVersion.value,
      changeReason: changeReason.value,
    });

    await loadTemplateData();
    await loadAuditEvents();
    changeReason.value = "";
    snackbar.success(t("common.saved"));
  } catch (e) {
    const normalized = normalizeApiError(e);
    templateError.value = normalized;
    snackbar.error(t(normalized.messageKey));
  } finally {
    activatingTemplateVersion.value = false;
  }
}

async function loadAuditEvents(): Promise<void> {
  loadingAudit.value = true;
  auditError.value = null;
  try {
    const projectID = resolveTemplateProjectIDContext(selectedTemplateKey.value);
    if (!auth.isPlatformAdmin && !projectID) {
      auditEvents.value = [];
      return;
    }
    auditEvents.value = await listPromptTemplateAuditEvents({
      templateKey: selectedTemplateKey.value || undefined,
      projectId: projectID || undefined,
      limit: 200,
    });
  } catch (e) {
    const normalized = normalizeApiError(e);
    auditError.value = normalized;
    snackbar.error(t(normalized.messageKey));
  } finally {
    loadingAudit.value = false;
  }
}

async function refreshTemplateData(): Promise<void> {
  await loadTemplateData();
  await loadAuditEvents();
}

async function refreshAll(): Promise<void> {
  await loadAgent();
  await loadTemplateKeys();
  await loadAuditEvents();
}

watch(
  () => selectedTemplateKey.value,
  async () => {
    if (!selectedTemplateKey.value.trim()) {
      versions.value = [];
      selectedVersion.value = null;
      preview.value = null;
      diff.value = null;
      await loadAuditEvents();
      return;
    }
    await loadTemplateData();
    await loadAuditEvents();
  },
);

watch(
  () => selectedVersion.value,
  (nextVersion) => {
    if (!nextVersion) {
      return;
    }
    const selected = versions.value.find((item) => item.version === nextVersion);
    if (selected) {
      editorMarkdown.value = selected.body_markdown;
    }
  },
);

watch(
  () => [diffFromVersion.value, diffToVersion.value, selectedTemplateKey.value],
  async () => {
    try {
      await loadDiff();
    } catch (e) {
      const normalized = normalizeApiError(e);
      templateError.value = normalized;
      snackbar.error(t(normalized.messageKey));
    }
  },
);

onMounted(async () => {
  await loadAgent();
  await loadTemplateKeys();
  await loadAuditEvents();
});
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
