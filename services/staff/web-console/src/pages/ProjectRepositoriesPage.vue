<template>
  <div>
    <PageHeader :title="t('pages.projectRepositories.title')">
      <template #leading>
        <BackBtn :label="t('common.back')" :to="{ name: 'projects' }" />
      </template>
      <template #actions>
        <CopyChip :label="t('pages.projectRepositories.projectId')" :value="projectId" icon="mdi-identifier" />
        <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :loading="repos.loading" @click="load" />
      </template>
    </PageHeader>

    <div class="mt-2 text-body-2 text-medium-emphasis">
      <RouterLink
        v-if="details.item"
        class="text-primary font-weight-bold text-decoration-none"
        :to="{ name: 'project-details', params: { projectId } }"
      >
        {{ details.item.name }}
      </RouterLink>
    </div>

    <VAlert v-if="repos.error" type="error" variant="tonal" class="mt-4">
      {{ t(repos.error.messageKey) }}
    </VAlert>
    <VAlert v-if="repos.attachError" type="error" variant="tonal" class="mt-4">
      {{ t(repos.attachError.messageKey) }}
    </VAlert>
    <VAlert v-if="repos.botUpdateError" type="error" variant="tonal" class="mt-4">
      {{ t(repos.botUpdateError.messageKey) }}
    </VAlert>
    <VAlert v-if="repos.preflightError" type="error" variant="tonal" class="mt-4">
      {{ t(repos.preflightError.messageKey) }}
    </VAlert>
    <VAlert v-if="repos.docsetGroupsError" type="error" variant="tonal" class="mt-4">
      {{ t(repos.docsetGroupsError.messageKey) }}
    </VAlert>
    <VAlert v-if="repos.docsetImportError" type="error" variant="tonal" class="mt-4">
      {{ t(repos.docsetImportError.messageKey) }}
    </VAlert>
    <VAlert v-if="repos.docsetSyncError" type="error" variant="tonal" class="mt-4">
      {{ t(repos.docsetSyncError.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable :headers="headers" :items="repos.items" :loading="repos.loading" :items-per-page="10" hover>
          <template #item.repo="{ item }">
            <span class="mono text-medium-emphasis">{{ item.owner }}/{{ item.name }}</span>
          </template>

          <template #item.services_yaml_path="{ item }">
            <span class="mono text-medium-emphasis">{{ item.services_yaml_path }}</span>
          </template>

          <template #item.bot="{ item }">
            <div class="d-flex flex-column align-center">
              <div class="mono text-body-2">{{ item.bot_username ?? "-" }}</div>
              <div class="mono text-caption text-medium-emphasis">{{ item.bot_email ?? "" }}</div>
            </div>
          </template>

          <template #item.preflight_updated_at="{ item }">
            <span class="text-medium-emphasis">{{ fmtDateTime(item.preflight_updated_at) }}</span>
          </template>

          <template #item.actions="{ item }">
            <div class="d-flex justify-end">
              <VTooltip :text="t('pages.projectRepositories.botParams')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    variant="tonal"
                    icon="mdi-robot-outline"
                    :loading="repos.botUpdating"
                    @click="openBotDialog(item)"
                  />
                </template>
              </VTooltip>
              <VTooltip :text="t('pages.projectRepositories.preflight')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    class="ml-2"
                    variant="tonal"
                    icon="mdi-check-decagram-outline"
                    :loading="repos.preflighting"
                    @click="openPreflightDialog(item)"
                  />
                </template>
              </VTooltip>
              <VTooltip :text="t('pages.projectRepositories.docsetImport')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    class="ml-2"
                    variant="tonal"
                    icon="mdi-book-plus-outline"
                    :loading="repos.docsetImporting"
                    @click="openDocsetImportDialog(item)"
                  />
                </template>
              </VTooltip>
              <VTooltip :text="t('pages.projectRepositories.docsetSync')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    class="ml-2"
                    variant="tonal"
                    icon="mdi-book-sync-outline"
                    :loading="repos.docsetSyncing"
                    @click="openDocsetSyncDialog(item)"
                  />
                </template>
              </VTooltip>
              <VTooltip v-if="!isPlatformRepository(item)" :text="t('common.delete')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    class="ml-2"
                    color="error"
                    variant="tonal"
                    icon="mdi-delete-outline"
                    :loading="repos.removing"
                    @click="askRemove(item.id, `${item.owner}/${item.name}`)"
                  />
                </template>
              </VTooltip>
            </div>
          </template>

          <template #no-data>
            <div class="py-8 text-medium-emphasis">
              {{ t("states.noRepos") }}
            </div>
          </template>
        </VDataTable>
      </VCardText>
    </VCard>

    <VCard class="mt-6" variant="outlined">
      <VCardTitle class="text-subtitle-1">{{ t("pages.projectRepositories.attachTitle") }}</VCardTitle>
      <VCardText>
        <VRow density="compact" class="align-end">
          <VCol cols="12" md="4">
            <VTextField v-model.trim="owner" :label="t('pages.projectRepositories.owner')" :placeholder="t('placeholders.repoOwner')" />
          </VCol>
          <VCol cols="12" md="4">
            <VTextField v-model.trim="name" :label="t('pages.projectRepositories.name')" :placeholder="t('placeholders.repoName')" />
          </VCol>
          <VCol cols="12" md="4">
            <VTextField
              v-model.trim="servicesYamlPath"
              :label="t('pages.projectRepositories.servicesYamlPath')"
              :placeholder="t('placeholders.servicesYamlPath')"
            />
          </VCol>
          <VCol cols="12">
            <VTextField
              v-model="token"
              :label="t('pages.projectRepositories.repoToken')"
              :placeholder="t('placeholders.repoToken')"
              type="password"
            />
          </VCol>
          <VCol cols="12" md="6">
            <VTextField v-model="newBotUsername" :label="t('pages.projectRepositories.botUsername')" hide-details />
          </VCol>
          <VCol cols="12" md="6">
            <VTextField v-model="newBotEmail" :label="t('pages.projectRepositories.botEmail')" hide-details />
          </VCol>
          <VCol cols="12">
            <VTextField v-model="newBotToken" :label="t('pages.projectRepositories.botToken')" type="password" hide-details />
            <div class="text-caption text-medium-emphasis mt-1">{{ t("pages.projectRepositories.botTokenHint") }}</div>
          </VCol>
          <VCol cols="12">
            <div class="d-flex ga-2 flex-wrap">
              <AdaptiveBtn
                color="primary"
                variant="tonal"
                icon="mdi-source-repository"
                :label="t('common.attachEnsureWebhook')"
                :loading="repos.attaching"
                @click="attach"
              />
              <AdaptiveBtn
                variant="tonal"
                icon="mdi-check-decagram-outline"
                :label="t('pages.projectRepositories.runPreflight')"
                :loading="repos.attaching || repos.preflighting"
                @click="attachAndPreflight"
              />
            </div>
          </VCol>
        </VRow>
      </VCardText>
    </VCard>
  </div>

  <ConfirmDialog
    v-model="confirmOpen"
    :title="t('common.delete')"
    :message="confirmName"
    :confirm-text="t('common.delete')"
    :cancel-text="t('common.cancel')"
    danger
    @confirm="doRemove"
  />

  <VDialog v-model="botDialogOpen" max-width="720">
    <VCard>
      <VCardTitle class="text-subtitle-1">{{ t("pages.projectRepositories.botParamsTitle") }}</VCardTitle>
      <VCardText>
        <div class="text-body-2 text-medium-emphasis mb-3">{{ botRepositoryLabel }}</div>
        <VRow density="compact">
          <VCol cols="12" md="6">
            <VTextField v-model="botUsername" :label="t('pages.projectRepositories.botUsername')" hide-details />
          </VCol>
          <VCol cols="12" md="6">
            <VTextField v-model="botEmail" :label="t('pages.projectRepositories.botEmail')" hide-details />
          </VCol>
          <VCol cols="12">
            <VTextField v-model="botToken" :label="t('pages.projectRepositories.botToken')" type="password" hide-details />
            <div class="text-caption text-medium-emphasis mt-1">{{ t("pages.projectRepositories.botTokenHint") }}</div>
          </VCol>
        </VRow>
      </VCardText>
      <VCardActions>
        <VSpacer />
        <VBtn variant="text" @click="botDialogOpen = false">{{ t("common.cancel") }}</VBtn>
        <VBtn color="primary" variant="tonal" :loading="repos.botUpdating" @click="saveBotParams">{{ t("common.save") }}</VBtn>
      </VCardActions>
    </VCard>
  </VDialog>

  <VDialog v-model="preflightDialogOpen" max-width="920">
    <VCard>
      <VCardTitle class="text-subtitle-1">{{ t("pages.projectRepositories.preflightTitle") }}</VCardTitle>
      <VCardText>
        <div class="text-body-2 text-medium-emphasis mb-3">{{ preflightRepositoryLabel }}</div>
        <VBtn color="primary" variant="tonal" :loading="repos.preflighting" @click="runPreflightNow">
          {{ t("pages.projectRepositories.runPreflight") }}
        </VBtn>
        <VAlert v-if="preflightResult" type="info" variant="tonal" class="mt-4">
          <div class="d-flex justify-space-between">
            <div class="text-body-2">
              {{ t("pages.projectRepositories.preflightStatus") }}: <strong>{{ preflightResult.status }}</strong>
            </div>
            <div class="text-body-2 text-medium-emphasis">
              {{ fmtDateTime(preflightResult.finished_at) }}
            </div>
          </div>
        </VAlert>
        <VCard v-if="preflightResult" class="mt-4" variant="outlined">
          <VCardTitle class="text-subtitle-2">{{ t("pages.projectRepositories.preflightChecks") }}</VCardTitle>
          <VCardText class="pt-0">
            <VDataTable
              :headers="[
                { title: t('pages.projectRepositories.check'), key: 'name', align: 'start' },
                { title: t('pages.projectRepositories.status'), key: 'status', width: 140, align: 'center' },
                { title: t('pages.projectRepositories.details'), key: 'details', align: 'start' },
              ]"
              :items="preflightResult.checks"
              :items-per-page="10"
              density="compact"
            />
          </VCardText>
        </VCard>
        <VCard v-if="preflightResult" class="mt-4" variant="outlined">
          <VCardTitle class="text-subtitle-2">{{ t("pages.projectRepositories.preflightRawReport") }}</VCardTitle>
          <VCardText>
            <pre class="mono preflight-json">{{ preflightResult.report_json }}</pre>
          </VCardText>
        </VCard>
      </VCardText>
      <VCardActions>
        <VSpacer />
        <VBtn variant="text" @click="preflightDialogOpen = false">{{ t("common.close") }}</VBtn>
      </VCardActions>
    </VCard>
  </VDialog>

  <VDialog v-model="docsetImportDialogOpen" max-width="920">
    <VCard>
      <VCardTitle class="text-subtitle-1">{{ t("pages.projectRepositories.docsetImportTitle") }}</VCardTitle>
      <VCardText>
        <div class="text-body-2 text-medium-emphasis mb-3">{{ docsetRepositoryLabel }}</div>
        <VRow density="compact">
          <VCol cols="12" md="4">
            <VSelect
              v-model="docsetLocale"
              :items="['ru', 'en']"
              :label="t('pages.projectRepositories.docsetLocale')"
              hide-details
              @update:model-value="reloadDocsetGroups"
            />
          </VCol>
          <VCol cols="12" md="8">
            <VTextField v-model.trim="docsetRef" :label="t('pages.projectRepositories.docsetRef')" hide-details @blur="reloadDocsetGroups" />
          </VCol>
        </VRow>
        <VCard class="mt-4" variant="outlined">
          <VCardTitle class="text-subtitle-2">{{ t("pages.projectRepositories.docsetGroups") }}</VCardTitle>
          <VCardText class="pt-0">
            <VProgressLinear v-if="repos.docsetLoadingGroups" indeterminate class="mb-3" />
            <VCheckbox
              v-for="g in docsetGroups"
              :key="g.id"
              v-model="docsetSelectedGroupIds"
              :label="`${g.title} (${g.id})`"
              :value="g.id"
              density="compact"
            />
          </VCardText>
        </VCard>
      </VCardText>
      <VCardActions>
        <VSpacer />
        <VBtn variant="text" @click="docsetImportDialogOpen = false">{{ t("common.cancel") }}</VBtn>
        <VBtn color="primary" variant="tonal" :loading="repos.docsetImporting" @click="importDocsetNow">{{ t("pages.projectRepositories.docsetImport") }}</VBtn>
      </VCardActions>
    </VCard>
  </VDialog>

  <VDialog v-model="docsetSyncDialogOpen" max-width="720">
    <VCard>
      <VCardTitle class="text-subtitle-1">{{ t("pages.projectRepositories.docsetSyncTitle") }}</VCardTitle>
      <VCardText>
        <div class="text-body-2 text-medium-emphasis mb-3">{{ docsetSyncRepositoryLabel }}</div>
        <VTextField v-model.trim="docsetSyncRef" :label="t('pages.projectRepositories.docsetRef')" hide-details />
      </VCardText>
      <VCardActions>
        <VSpacer />
        <VBtn variant="text" @click="docsetSyncDialogOpen = false">{{ t("common.cancel") }}</VBtn>
        <VBtn color="primary" variant="tonal" :loading="repos.docsetSyncing" @click="syncDocsetNow">{{ t("pages.projectRepositories.docsetSync") }}</VBtn>
      </VCardActions>
    </VCard>
  </VDialog>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import ConfirmDialog from "../shared/ui/ConfirmDialog.vue";
import CopyChip from "../shared/ui/CopyChip.vue";
import PageHeader from "../shared/ui/PageHeader.vue";
import AdaptiveBtn from "../shared/ui/AdaptiveBtn.vue";
import BackBtn from "../shared/ui/BackBtn.vue";
import { useSnackbarStore } from "../shared/ui/feedback/snackbar-store";
import { useProjectRepositoriesStore } from "../features/projects/repositories-store";
import { useProjectDetailsStore } from "../features/projects/details-store";
import type { DocsetGroup, RepositoryBinding, RunRepositoryPreflightResponse } from "../features/projects/types";
import { formatDateTime } from "../shared/lib/datetime";

const props = defineProps<{ projectId: string }>();

const { t, locale } = useI18n({ useScope: "global" });
const repos = useProjectRepositoriesStore();
const details = useProjectDetailsStore();
const snackbar = useSnackbarStore();

const owner = ref("");
const name = ref("");
const servicesYamlPath = ref("services.yaml");
const token = ref("");
const newBotToken = ref("");
const newBotUsername = ref("");
const newBotEmail = ref("");

const confirmOpen = ref(false);
const confirmRepoId = ref("");
const confirmName = ref("");

const headers = [
  { title: t("pages.projectRepositories.provider"), key: "provider", width: 160, align: "start" },
  { title: t("pages.projectRepositories.repo"), key: "repo", sortable: false, width: 260, align: "center" },
  { title: t("pages.projectRepositories.servicesYaml"), key: "services_yaml_path", width: 220, align: "center" },
  { title: t("pages.projectRepositories.bot"), key: "bot", sortable: false, width: 220, align: "center" },
  { title: t("pages.projectRepositories.preflightUpdatedAt"), key: "preflight_updated_at", width: 200, align: "center" },
  { title: "", key: "actions", sortable: false, width: 72, align: "end" },
] as const;

async function load() {
  await details.load(props.projectId);
  await repos.load(props.projectId);
}

async function attach() {
  const binding = await repos.attach({
    owner: owner.value,
    name: name.value,
    token: token.value,
    servicesYamlPath: servicesYamlPath.value,
    botToken: newBotToken.value.trim() === "" ? null : newBotToken.value,
    botUsername: newBotUsername.value.trim() ? newBotUsername.value.trim() : null,
    botEmail: newBotEmail.value.trim() ? newBotEmail.value.trim() : null,
  });
  if (binding && !repos.attachError) {
    owner.value = "";
    name.value = "";
    token.value = "";
    servicesYamlPath.value = "services.yaml";
    newBotToken.value = "";
    newBotUsername.value = "";
    newBotEmail.value = "";
    snackbar.success(t("common.saved"));
  }
}

async function attachAndPreflight() {
  const binding = await repos.attach({
    owner: owner.value,
    name: name.value,
    token: token.value,
    servicesYamlPath: servicesYamlPath.value,
    botToken: newBotToken.value.trim() === "" ? null : newBotToken.value,
    botUsername: newBotUsername.value.trim() ? newBotUsername.value.trim() : null,
    botEmail: newBotEmail.value.trim() ? newBotEmail.value.trim() : null,
  });
  if (!binding || repos.attachError) return;
  openPreflightDialog(binding);
  await runPreflightNow();
}

function isPlatformRepository(item: RepositoryBinding): boolean {
  return `${item.owner}/${item.name}`.toLowerCase() === "codex-k8s/codex-k8s";
}

function askRemove(repositoryId: string, label: string) {
  confirmRepoId.value = repositoryId;
  confirmName.value = t("pages.projectRepositories.deleteConfirmMessage", { repo: label });
  confirmOpen.value = true;
}

async function doRemove() {
  const id = confirmRepoId.value;
  confirmRepoId.value = "";
  if (!id) return;
  await repos.remove(id);
  snackbar.success(t("common.deleted"));
}

function fmtDateTime(value: string | null | undefined): string {
  return formatDateTime(value, locale.value);
}

const botDialogOpen = ref(false);
const botRepositoryId = ref("");
const botRepositoryLabel = ref("");
const botToken = ref("");
const botUsername = ref("");
const botEmail = ref("");

function openBotDialog(item: RepositoryBinding) {
  botRepositoryId.value = item.id;
  botRepositoryLabel.value = `${item.owner}/${item.name}`;
  botToken.value = "";
  botUsername.value = item.bot_username ?? "";
  botEmail.value = item.bot_email ?? "";
  botDialogOpen.value = true;
}

async function saveBotParams() {
  if (!botRepositoryId.value) return;
  await repos.upsertBotParams(botRepositoryId.value, {
    botToken: botToken.value.trim() === "" ? null : botToken.value,
    botUsername: botUsername.value.trim() ? botUsername.value.trim() : null,
    botEmail: botEmail.value.trim() ? botEmail.value.trim() : null,
  });
  if (!repos.botUpdateError) {
    botDialogOpen.value = false;
    snackbar.success(t("common.saved"));
  }
}

const preflightDialogOpen = ref(false);
const preflightRepositoryId = ref("");
const preflightRepositoryLabel = ref("");
const preflightResult = ref<RunRepositoryPreflightResponse | null>(null);

function openPreflightDialog(item: RepositoryBinding) {
  preflightRepositoryId.value = item.id;
  preflightRepositoryLabel.value = `${item.owner}/${item.name}`;
  preflightResult.value = null;
  preflightDialogOpen.value = true;
}

async function runPreflightNow() {
  if (!preflightRepositoryId.value) return;
  const r = await repos.runPreflight(preflightRepositoryId.value);
  if (r) {
    preflightResult.value = r;
  }
}

const docsetImportDialogOpen = ref(false);
const docsetRepositoryId = ref("");
const docsetRepositoryLabel = ref("");
const docsetRef = ref("main");
const docsetLocale = ref<"ru" | "en">("ru");
const docsetGroups = ref<DocsetGroup[]>([]);
const docsetSelectedGroupIds = ref<string[]>([]);

async function openDocsetImportDialog(item: RepositoryBinding) {
  docsetRepositoryId.value = item.id;
  docsetRepositoryLabel.value = `${item.owner}/${item.name}`;
  docsetRef.value = "main";
  docsetLocale.value = "ru";
  docsetGroups.value = await repos.loadDocsetGroups({ docsetRef: docsetRef.value, locale: docsetLocale.value });
  docsetSelectedGroupIds.value = docsetGroups.value.filter((g) => g.default_selected).map((g) => g.id);
  docsetImportDialogOpen.value = true;
}

async function reloadDocsetGroups() {
  docsetGroups.value = await repos.loadDocsetGroups({ docsetRef: docsetRef.value, locale: docsetLocale.value });
  if (docsetSelectedGroupIds.value.length === 0) {
    docsetSelectedGroupIds.value = docsetGroups.value.filter((g) => g.default_selected).map((g) => g.id);
  }
}

async function importDocsetNow() {
  if (!docsetRepositoryId.value) return;
  const resp = await repos.importDocset({
    repositoryId: docsetRepositoryId.value,
    docsetRef: docsetRef.value,
    locale: docsetLocale.value,
    groupIds: docsetSelectedGroupIds.value,
  });
  if (resp) {
    docsetImportDialogOpen.value = false;
    snackbar.success(`${t("pages.projectRepositories.docsetPrCreated")} #${resp.pr_number}`);
  }
}

const docsetSyncDialogOpen = ref(false);
const docsetSyncRepositoryId = ref("");
const docsetSyncRepositoryLabel = ref("");
const docsetSyncRef = ref("main");

function openDocsetSyncDialog(item: RepositoryBinding) {
  docsetSyncRepositoryId.value = item.id;
  docsetSyncRepositoryLabel.value = `${item.owner}/${item.name}`;
  docsetSyncRef.value = "main";
  docsetSyncDialogOpen.value = true;
}

async function syncDocsetNow() {
  if (!docsetSyncRepositoryId.value) return;
  const resp = await repos.syncDocset({ repositoryId: docsetSyncRepositoryId.value, docsetRef: docsetSyncRef.value });
  if (resp) {
    docsetSyncDialogOpen.value = false;
    snackbar.success(`${t("pages.projectRepositories.docsetPrCreated")} #${resp.pr_number}`);
  }
}

onMounted(() => void load());

let attachErrorTimer: ReturnType<typeof setTimeout> | null = null;
watch(
  () => repos.attachError?.messageKey || "",
  (key) => {
    if (attachErrorTimer) {
      clearTimeout(attachErrorTimer);
      attachErrorTimer = null;
    }
    if (key !== "errors.invalidArgument") return;
    attachErrorTimer = setTimeout(() => {
      repos.attachError = null;
    }, 5000);
  },
);
onBeforeUnmount(() => {
  if (!attachErrorTimer) return;
  clearTimeout(attachErrorTimer);
  attachErrorTimer = null;
});
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.preflight-json {
  max-height: 320px;
  overflow: auto;
  padding: 12px;
  border: 1px solid rgba(0, 0, 0, 0.12);
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.02);
}
</style>
