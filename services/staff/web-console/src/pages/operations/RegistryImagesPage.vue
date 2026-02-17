<template>
  <div>
    <PageHeader :title="t('pages.registryImages.title')" :hint="t('pages.registryImages.hint')">
      <template #actions>
        <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :disabled="loading" @click="loadImages" />
      </template>
    </PageHeader>

    <VAlert v-if="error" type="error" variant="tonal" class="mt-4">
      {{ t(error.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VRow density="compact">
          <VCol cols="12" md="6">
            <VTextField v-model.trim="repositoryFilter" :label="t('pages.registryImages.repositoryFilter')" hide-details clearable />
          </VCol>
          <VCol cols="12" md="3">
            <VTextField
              v-model.number="limitRepositories"
              type="number"
              min="1"
              max="1000"
              :label="t('pages.registryImages.limitRepositories')"
              hide-details
            />
          </VCol>
          <VCol cols="12" md="3">
            <VTextField
              v-model.number="limitTags"
              type="number"
              min="1"
              max="1000"
              :label="t('pages.registryImages.limitTags')"
              hide-details
            />
          </VCol>
        </VRow>
      </VCardText>
      <VCardActions>
        <VSpacer />
        <AdaptiveBtn variant="tonal" icon="mdi-check" :label="t('pages.runs.applyFilters')" :disabled="loading" @click="loadImages" />
      </VCardActions>
    </VCard>

    <VCard class="mt-4" variant="outlined">
      <VCardTitle>{{ t("pages.registryImages.cleanupTitle") }}</VCardTitle>
      <VCardText>
        <VRow density="compact">
          <VCol cols="12" md="5">
            <VTextField v-model.trim="cleanupPrefix" :label="t('pages.registryImages.cleanupPrefix')" hide-details clearable />
          </VCol>
          <VCol cols="12" md="2">
            <VTextField
              v-model.number="cleanupKeepTags"
              type="number"
              min="1"
              max="1000"
              :label="t('pages.registryImages.keepTags')"
              hide-details
            />
          </VCol>
          <VCol cols="12" md="2">
            <VTextField
              v-model.number="cleanupLimitRepositories"
              type="number"
              min="1"
              max="1000"
              :label="t('pages.registryImages.limitRepositories')"
              hide-details
            />
          </VCol>
          <VCol cols="12" md="3" class="d-flex align-center">
            <VCheckbox v-model="cleanupDryRun" :label="t('pages.registryImages.dryRun')" hide-details density="comfortable" />
          </VCol>
        </VRow>
        <div v-if="cleanupResult" class="text-body-2 text-medium-emphasis mt-2">
          {{
            t("pages.registryImages.cleanupResult", {
              repositories: cleanupResult.repositories_scanned,
              deleted: cleanupResult.tags_deleted,
              skipped: cleanupResult.tags_skipped,
            })
          }}
        </div>
      </VCardText>
      <VCardActions>
        <VSpacer />
        <AdaptiveBtn
          variant="tonal"
          icon="mdi-broom"
          :label="cleanupDryRun ? t('pages.registryImages.cleanupDryRun') : t('pages.registryImages.cleanupApply')"
          :disabled="cleanupLoading"
          @click="runCleanup"
        />
      </VCardActions>
    </VCard>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable
          :headers="headers"
          :items="rows"
          :loading="loading"
          :items-per-page="20"
          density="comfortable"
          hover
        >
          <template #item.repository="{ item }">
            <span class="mono">{{ item.repository }}</span>
          </template>
          <template #item.tag="{ item }">
            <div class="d-flex align-center justify-center ga-2">
              <span class="mono">{{ shortTag(item.tag) }}</span>
              <VTooltip :text="t('common.copy')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="x-small"
                    variant="text"
                    icon="mdi-content-save-outline"
                    :disabled="loading"
                    @click="copyToClipboard(item.tag)"
                  />
                </template>
              </VTooltip>
            </div>
          </template>
          <template #item.created_at="{ item }">
            <span class="text-medium-emphasis">{{ formatDateTime(item.created_at, locale) }}</span>
          </template>
          <template #item.config_size_bytes="{ item }">
            <span class="text-medium-emphasis">{{ formatBytes(item.config_size_bytes) }}</span>
          </template>
          <template #item.actions="{ item }">
            <VTooltip :text="t('common.delete')">
              <template #activator="{ props: tipProps }">
                <VBtn
                  v-bind="tipProps"
                  size="small"
                  variant="text"
                  color="error"
                  icon="mdi-delete-outline"
                  :disabled="deleteLoading"
                  @click="confirmDelete(item.repository, item.tag)"
                />
              </template>
            </VTooltip>
          </template>
          <template #no-data>
            <div class="py-8 text-medium-emphasis">
              {{ t("states.noRegistryImages") }}
            </div>
          </template>
        </VDataTable>
      </VCardText>
    </VCard>
  </div>

  <ConfirmDialog
    v-model="deleteConfirmOpen"
    :title="t('common.delete')"
    :message="deleteConfirmMessage"
    :confirm-text="t('common.delete')"
    :cancel-text="t('common.cancel')"
    danger
    @confirm="deleteTag"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import ConfirmDialog from "../../shared/ui/ConfirmDialog.vue";
import PageHeader from "../../shared/ui/PageHeader.vue";
import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { formatDateTime } from "../../shared/lib/datetime";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import {
  cleanupRegistryImages,
  deleteRegistryImageTag,
  listRegistryImages,
} from "../../features/runtime-deploy/api";
import type {
  CleanupRegistryImagesResponse,
  RegistryImageRepository,
} from "../../features/runtime-deploy/types";

type RegistryImageRow = {
  repository: string;
  tag: string;
  created_at?: string | null;
  config_size_bytes: number;
};

const { t, locale } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();

const loading = ref(false);
const cleanupLoading = ref(false);
const deleteLoading = ref(false);
const error = ref<ApiError | null>(null);
const repositories = ref<RegistryImageRepository[]>([]);

const repositoryFilter = ref("");
const limitRepositories = ref(100);
const limitTags = ref(50);

const cleanupPrefix = ref("");
const cleanupKeepTags = ref(5);
const cleanupLimitRepositories = ref(100);
const cleanupDryRun = ref(true);
const cleanupResult = ref<CleanupRegistryImagesResponse | null>(null);

const deleteConfirmOpen = ref(false);
const deleteRepository = ref("");
const deleteTagName = ref("");

const deleteConfirmMessage = computed(() => `${deleteRepository.value}:${deleteTagName.value}`);

const headers = computed(() => ([
  { title: t("table.fields.repository"), key: "repository", align: "center", width: 420 },
  { title: t("table.fields.tag"), key: "tag", align: "center", width: 140 },
  { title: t("table.fields.created_at"), key: "created_at", align: "center", width: 180 },
  { title: t("table.fields.config_size_bytes"), key: "config_size_bytes", align: "center", width: 160 },
  { title: "", key: "actions", sortable: false, align: "end", width: 72 },
]) as const);

const rows = computed<RegistryImageRow[]>(() => {
  const out: RegistryImageRow[] = [];
  for (const repositoryItem of repositories.value) {
    for (const tagItem of repositoryItem.tags || []) {
      out.push({
        repository: repositoryItem.repository,
        tag: tagItem.tag,
        created_at: tagItem.created_at,
        config_size_bytes: Number(tagItem.config_size_bytes || 0),
      });
    }
  }
  return out;
});

function shortTag(value: string): string {
  const tag = String(value || "").trim();
  if (tag.length <= 8) return tag;

  const lastDash = tag.lastIndexOf("-");
  if (lastDash >= 0 && lastDash < tag.length - 1) {
    const prefix = tag.slice(0, lastDash + 1);
    const suffix = tag.slice(lastDash + 1);
    if (/^[0-9a-f]+$/i.test(suffix) && suffix.length >= 8) {
      return `${prefix}${suffix.slice(0, 8)}`;
    }
  }

  if (/^[0-9a-f]+$/i.test(tag)) {
    return tag.slice(0, 8);
  }

  return `${tag.slice(0, 8)}...`;
}

function formatBytes(value: number): string {
  if (!Number.isFinite(value) || value <= 0) return "-";
  const units = ["B", "KB", "MB", "GB"];
  let size = value;
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit++;
  }
  return `${size.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`;
}

async function copyToClipboard(value: string): Promise<void> {
  const text = String(value || "");
  try {
    await navigator.clipboard.writeText(text);
    snackbar.success(t("common.copied"));
  } catch {
    snackbar.error(t("errors.copyFailed"));
  }
}

async function loadImages(): Promise<void> {
  loading.value = true;
  error.value = null;
  try {
    repositories.value = await listRegistryImages({
      repository: repositoryFilter.value,
      limitRepositories: limitRepositories.value,
      limitTags: limitTags.value,
    });
  } catch (err) {
    error.value = normalizeApiError(err);
  } finally {
    loading.value = false;
  }
}

function confirmDelete(repository: string, tag: string): void {
  deleteRepository.value = repository;
  deleteTagName.value = tag;
  deleteConfirmOpen.value = true;
}

async function deleteTag(): Promise<void> {
  if (!deleteRepository.value || !deleteTagName.value) return;
  deleteLoading.value = true;
  error.value = null;
  try {
    const result = await deleteRegistryImageTag(deleteRepository.value, deleteTagName.value);
    if (result.deleted) {
      snackbar.success(t("common.deleted"));
    } else {
      snackbar.info(t("pages.registryImages.notDeleted"));
    }
    await loadImages();
  } catch (err) {
    error.value = normalizeApiError(err);
  } finally {
    deleteLoading.value = false;
    deleteRepository.value = "";
    deleteTagName.value = "";
  }
}

async function runCleanup(): Promise<void> {
  cleanupLoading.value = true;
  error.value = null;
  try {
    cleanupResult.value = await cleanupRegistryImages({
      repositoryPrefix: cleanupPrefix.value,
      keepTags: cleanupKeepTags.value,
      limitRepositories: cleanupLimitRepositories.value,
      dryRun: cleanupDryRun.value,
    });
    snackbar.success(t("common.saved"));
    await loadImages();
  } catch (err) {
    error.value = normalizeApiError(err);
  } finally {
    cleanupLoading.value = false;
  }
}

onMounted(() => void loadImages());
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
