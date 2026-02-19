<template>
  <div>
    <PageHeader :title="t('pages.systemSettings.title')" :hint="t('pages.systemSettings.hint')">
      <template #actions>
        <AdaptiveBtn
          variant="tonal"
          icon="mdi-refresh"
          :label="t('common.refresh')"
          :disabled="platformTokensLoading"
          class="mr-2"
          @click="loadPlatformTokens"
        />
        <AdaptiveBtn
          variant="tonal"
          icon="mdi-plus"
          :label="t('pages.systemSettings.addLocale')"
          @click="addLocaleDialogOpen = true"
        />
      </template>
    </PageHeader>

    <VAlert v-if="platformTokensError" type="error" variant="tonal" class="mt-4">
      {{ t(platformTokensError.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardTitle class="text-subtitle-2">{{ t("pages.systemSettings.platformTokensTitle") }}</VCardTitle>
      <VCardText>
        <VAlert type="info" variant="tonal">
          <div class="text-body-2">
            {{ t("pages.systemSettings.platformTokensHint") }}
          </div>
        </VAlert>

        <VRow density="compact" class="mt-3">
          <VCol v-for="f in platformTokenFields" :key="f.key" cols="12" md="6">
            <VCard variant="tonal">
              <VCardTitle class="text-subtitle-2 d-flex align-center justify-space-between ga-2 flex-wrap">
                <span>{{ t(f.titleKey) }}</span>
                <VChip
                  v-if="f.kind === 'secret'"
                  size="x-small"
                  variant="tonal"
                  class="font-weight-bold"
                  :color="f.configured ? 'success' : 'secondary'"
                >
                  {{ f.configured ? t("pages.systemSettings.configured") : t("pages.systemSettings.notConfigured") }}
                </VChip>
              </VCardTitle>
              <VCardText>
                <VTextField
                  v-model="f.value"
                  :type="f.kind === 'secret' ? 'password' : 'text'"
                  :label="t('pages.systemSettings.newValue')"
                  hide-details
                />
                <div v-if="f.kind === 'variable'" class="text-caption text-medium-emphasis mt-1">
                  {{ t("pages.systemSettings.currentValue") }}: <span class="mono">{{ f.currentValue || "-" }}</span>
                </div>
                <div v-if="f.updatedAt" class="text-caption text-medium-emphasis mt-1">
                  {{ t("pages.systemSettings.updatedAt") }}: {{ f.updatedAt }}
                </div>
                <div v-if="f.syncTargets.length === 0" class="text-caption text-warning mt-1">
                  {{ t("pages.systemSettings.noSyncTargets") }}
                </div>
              </VCardText>
              <VCardActions>
                <VSpacer />
                <VBtn
                  color="primary"
                  variant="tonal"
                  :disabled="!canSaveTokenField(f)"
                  :loading="platformTokensSaving"
                  @click="saveTokenField(f)"
                >
                  {{ t("common.save") }}
                </VBtn>
              </VCardActions>
            </VCard>
          </VCol>
        </VRow>
      </VCardText>
    </VCard>

    <VRow class="mt-4" density="compact">
      <VCol cols="12" md="6">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-2">{{ t("pages.systemSettings.localesTitle") }}</VCardTitle>
          <VCardText>
            <VDataTable :headers="localeHeaders" :items="locales" :items-per-page="10" density="comfortable">
              <template #item.is_default="{ item }">
                <div class="d-flex justify-center">
                  <VChip size="small" variant="tonal" class="font-weight-bold" :color="item.is_default ? 'success' : 'secondary'">
                    {{ item.is_default ? t("bool.true") : t("bool.false") }}
                  </VChip>
                </div>
              </template>
            </VDataTable>
          </VCardText>
        </VCard>
      </VCol>
      <VCol cols="12" md="6">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-2">{{ t("pages.systemSettings.uiPrefsTitle") }}</VCardTitle>
          <VCardText>
            <VRow density="compact">
              <VCol cols="12">
                <VSelect
                  v-model="prefsDensity"
                  :items="densityOptions"
                  :label="t('pages.systemSettings.density')"
                  hide-details
                />
              </VCol>
              <VCol cols="12">
                <VSelect
                  v-model="prefsDateFormat"
                  :items="dateFormatOptions"
                  :label="t('pages.systemSettings.dateTimeFormat')"
                  hide-details
                />
              </VCol>
              <VCol cols="12">
                <VSwitch v-model="prefsDebugHints" :label="t('pages.systemSettings.debugHints')" hide-details />
              </VCol>
            </VRow>

            <VAlert type="info" variant="tonal" class="mt-3">
              <div class="text-body-2">
                {{ t("pages.systemSettings.scaffoldNote") }}
              </div>
            </VAlert>
          </VCardText>
        </VCard>
      </VCol>
    </VRow>

    <VDialog v-model="addLocaleDialogOpen" max-width="520">
      <VCard>
        <VCardTitle class="text-subtitle-1">{{ t("pages.systemSettings.addLocaleTitle") }}</VCardTitle>
        <VCardText>
          <VRow density="compact">
            <VCol cols="12" sm="4">
              <VTextField v-model.trim="newLocaleCode" :label="t('pages.systemSettings.localeCode')" hide-details />
            </VCol>
            <VCol cols="12" sm="8">
              <VTextField v-model.trim="newLocaleName" :label="t('pages.systemSettings.localeName')" hide-details />
            </VCol>
          </VRow>
          <VAlert type="warning" variant="tonal" class="mt-3">
            <div class="text-body-2">
              {{ t("pages.systemSettings.addLocaleScaffoldWarning") }}
            </div>
          </VAlert>
        </VCardText>
        <VCardActions>
          <VSpacer />
          <VBtn variant="text" @click="addLocaleDialogOpen = false">{{ t("common.cancel") }}</VBtn>
          <VBtn variant="tonal" @click="mockAddLocale">{{ t("common.save") }}</VBtn>
        </VCardActions>
      </VCard>
    </VDialog>
  </div>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные локали (backend API + хранение), а также persist UI prefs (per-user settings).
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

import PageHeader from "../../shared/ui/PageHeader.vue";
import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import { formatDateTime } from "../../shared/lib/datetime";
import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { listConfigEntries, upsertConfigEntry } from "../../features/config/api";
import type { ConfigEntry } from "../../features/config/types";

type LocaleRow = {
  code: string;
  name: string;
  is_default: boolean;
};

const { t, locale } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();

type PlatformTokenField = {
  key: string;
  titleKey: string;
  kind: "variable" | "secret";
  value: string;
  currentValue: string;
  updatedAt: string;
  configured: boolean;
  syncTargets: string[];
};

const platformTokensLoading = ref(false);
const platformTokensSaving = ref(false);
const platformTokensError = ref<ApiError | null>(null);
const platformEntries = ref<ConfigEntry[]>([]);

const platformTokenKeys = [
  { key: "CODEXK8S_GITHUB_PAT", titleKey: "pages.systemSettings.platformTokenGitHubPat", kind: "secret" as const },
  { key: "CODEXK8S_GIT_BOT_TOKEN", titleKey: "pages.systemSettings.platformTokenGitBotToken", kind: "secret" as const },
  { key: "CODEXK8S_GIT_BOT_USERNAME", titleKey: "pages.systemSettings.platformTokenGitBotUsername", kind: "variable" as const },
  { key: "CODEXK8S_GIT_BOT_MAIL", titleKey: "pages.systemSettings.platformTokenGitBotMail", kind: "variable" as const },
] as const;

const platformTokenFields = ref<PlatformTokenField[]>([]);

function buildPlatformTokenFields(items: ConfigEntry[]): PlatformTokenField[] {
  const byKey = new Map<string, ConfigEntry>();
  for (const item of items) {
    const k = String(item?.key || "").trim();
    if (!k) continue;
    byKey.set(k, item);
  }

  return platformTokenKeys.map((spec) => {
    const item = byKey.get(spec.key);
    const updatedAt = item?.updated_at ? formatDateTime(item.updated_at, locale.value) : "";
    return {
      key: spec.key,
      titleKey: spec.titleKey,
      kind: spec.kind,
      value: "",
      currentValue: spec.kind === "variable" ? String(item?.value || "").trim() : "",
      updatedAt,
      configured: Boolean(item && (item.kind === "variable" ? String(item.value || "").trim() !== "" : true)),
      syncTargets: item?.sync_targets ? [...item.sync_targets] : [],
    };
  });
}

async function loadPlatformTokens(): Promise<void> {
  platformTokensLoading.value = true;
  platformTokensError.value = null;
  try {
    const items = await listConfigEntries({ scope: "platform", limit: 200 });
    platformEntries.value = items;
    platformTokenFields.value = buildPlatformTokenFields(items);
  } catch (e) {
    platformTokensError.value = normalizeApiError(e);
  } finally {
    platformTokensLoading.value = false;
  }
}

function canSaveTokenField(f: PlatformTokenField): boolean {
  if (platformTokensLoading.value || platformTokensSaving.value) return false;
  const v = String(f.value || "").trim();
  if (!v) return false;
  if (f.kind === "variable" && v === String(f.currentValue || "").trim()) return false;
  return true;
}

async function saveTokenField(f: PlatformTokenField): Promise<void> {
  const value = String(f.value || "").trim();
  if (!value) return;

  platformTokensSaving.value = true;
  platformTokensError.value = null;
  try {
    await upsertConfigEntry({
      scope: "platform",
      kind: f.kind,
      projectId: null,
      repositoryId: null,
      key: f.key,
      valuePlain: f.kind === "variable" ? value : null,
      valueSecret: f.kind === "secret" ? value : null,
      syncTargets: f.syncTargets,
      mutability: "runtime_mutable",
      isDangerous: false,
      dangerousConfirmed: false,
    });

    snackbar.success(t("common.saved"));
    f.value = "";
    await loadPlatformTokens();
  } catch (e) {
    platformTokensError.value = normalizeApiError(e);
  } finally {
    platformTokensSaving.value = false;
  }
}

const localeHeaders = [
  { title: t("table.fields.code"), key: "code", width: 120, align: "start" },
  { title: t("table.fields.name"), key: "name", align: "center" },
  { title: t("table.fields.is_default"), key: "is_default", width: 160, align: "center" },
] as const;

const locales = ref<LocaleRow[]>([
  { code: "en", name: "English", is_default: true },
  { code: "ru", name: "Русский", is_default: false },
]);

const addLocaleDialogOpen = ref(false);
const newLocaleCode = ref("");
const newLocaleName = ref("");

const densityOptions = ["default", "comfortable", "compact"] as const;
const dateFormatOptions = ["local", "iso", "relative"] as const;

const prefsDensity = ref<(typeof densityOptions)[number]>("comfortable");
const prefsDateFormat = ref<(typeof dateFormatOptions)[number]>("local");
const prefsDebugHints = ref(false);

function mockAddLocale(): void {
  // scaffold: no persistence
  const code = newLocaleCode.value.trim().toLowerCase();
  const name = newLocaleName.value.trim();
  if (code && name) {
    locales.value = [...locales.value, { code, name, is_default: false }];
  }
  newLocaleCode.value = "";
  newLocaleName.value = "";
  addLocaleDialogOpen.value = false;
}

onMounted(() => {
  void loadPlatformTokens();
});
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
