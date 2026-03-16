<template>
  <div>
    <PageHeader :title="t('pages.systemSettings.title')" :hint="t('pages.systemSettings.hint')">
      <template #actions>
        <AdaptiveBtn
          variant="tonal"
          icon="mdi-refresh"
          :label="t('common.refresh')"
          :loading="loading"
          @click="void refreshPage()"
        />
      </template>
    </PageHeader>

    <VRow class="mt-4" density="compact">
      <VCol cols="12">
        <VCard variant="outlined" class="system-settings-summary">
          <VCardText class="d-flex flex-wrap align-center ga-3">
            <VChip
              size="small"
              variant="tonal"
              :color="realtimeStateColor"
              prepend-icon="mdi-access-point"
            >
              {{ t("pages.systemSettings.realtime") }}: {{ realtimeStateLabel }}
            </VChip>
            <div class="text-body-2 text-medium-emphasis">
              {{ t("pages.systemSettings.liveHint") }}
            </div>
          </VCardText>
        </VCard>
      </VCol>

      <VCol v-if="error" cols="12">
        <VAlert type="error" variant="tonal">
          {{ t(error.messageKey) }}
        </VAlert>
      </VCol>

      <VCol v-if="!loading && settings.length === 0 && !error" cols="12">
        <VAlert type="info" variant="tonal">
          {{ t("pages.systemSettings.empty") }}
        </VAlert>
      </VCol>

      <VCol v-for="item in settings" :key="item.key" cols="12">
        <VCard variant="outlined" class="system-settings-card">
          <VCardTitle class="d-flex flex-wrap align-start justify-space-between ga-3">
            <div>
              <div class="text-subtitle-1 font-weight-bold">
                {{ settingTitle(item.key) }}
              </div>
              <div class="text-body-2 text-medium-emphasis mt-1">
                {{ settingDescription(item.key) }}
              </div>
            </div>
            <VChip
              size="small"
              variant="tonal"
              :color="currentBooleanValue(item) ? 'success' : 'secondary'"
              class="font-weight-bold"
            >
              {{ currentBooleanValue(item) ? t("pages.systemSettings.enabled") : t("pages.systemSettings.disabled") }}
            </VChip>
          </VCardTitle>

          <VCardText>
            <div class="d-flex flex-wrap ga-2 mb-4">
              <VChip size="small" variant="outlined">
                {{ sectionLabel(item.section) }}
              </VChip>
              <VChip size="small" variant="tonal" color="primary">
                {{ sourceLabel(item.source) }}
              </VChip>
              <VChip size="small" variant="tonal" color="info">
                {{ reloadLabel(item.reload_semantics) }}
              </VChip>
              <VChip size="small" variant="tonal" color="grey-darken-1">
                v{{ item.version }}
              </VChip>
            </div>

            <VSwitch
              :model-value="currentBooleanValue(item)"
              :label="t('pages.systemSettings.fields.currentValue')"
              :loading="savingKey === item.key"
              :disabled="isBusy(item.key)"
              color="success"
              inset
              hide-details
              @update:model-value="(value) => void handleBooleanToggle(item, Boolean(value))"
            />

            <VList density="compact" class="bg-transparent px-0 mt-3">
              <VListItem :title="t('pages.systemSettings.fields.defaultValue')" :subtitle="item.default_boolean_value ? t('pages.systemSettings.enabled') : t('pages.systemSettings.disabled')" />
              <VListItem :title="t('pages.systemSettings.fields.updatedAt')" :subtitle="formatSettingDate(item.updated_at)" />
              <VListItem :title="t('pages.systemSettings.fields.updatedBy')" :subtitle="updatedByLabel(item)" />
              <VListItem :title="t('pages.systemSettings.fields.visibility')" :subtitle="visibilityLabel(item.visibility)" />
              <VListItem :title="t('pages.systemSettings.fields.key')" :subtitle="item.key" />
            </VList>
          </VCardText>

          <VCardActions>
            <VSpacer />
            <VBtn
              variant="text"
              color="secondary"
              :loading="resettingKey === item.key"
              :disabled="isBusy(item.key) || (item.source === 'default' && item.boolean_value === item.default_boolean_value)"
              @click="void handleReset(item)"
            >
              {{ t("common.reset") }}
            </VBtn>
          </VCardActions>
        </VCard>
      </VCol>
    </VRow>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

import type { ApiError } from "../../shared/api/errors";
import { ApiError as LocalApiError, normalizeApiError } from "../../shared/api/errors";
import { formatDateTime } from "../../shared/lib/datetime";
import { bindRealtimePageLifecycle } from "../../shared/ws/lifecycle";
import PageHeader from "../../shared/ui/PageHeader.vue";
import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { listSystemSettings, resetSystemSetting, updateSystemSettingBoolean } from "../../features/system-settings/api";
import { subscribeSystemSettingsRealtime } from "../../features/system-settings/realtime";
import type { SystemSetting, SystemSettingsRealtimeMessage, SystemSettingsRealtimeState } from "../../features/system-settings/types";

const { t, locale } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();

const loading = ref(true);
const error = ref<ApiError | null>(null);
const settings = ref<SystemSetting[]>([]);
const savingKey = ref("");
const resettingKey = ref("");
const draftValues = ref<Record<string, boolean>>({});
const realtimeState = ref<SystemSettingsRealtimeState | "closed">("closed");
const stopRealtimeRef = ref<(() => void) | null>(null);
const stopLifecycleBindingRef = ref<(() => void) | null>(null);

const realtimeStateLabel = computed(() => {
  switch (realtimeState.value) {
    case "connected":
      return t("pages.systemSettings.realtimeConnected");
    case "reconnecting":
      return t("pages.systemSettings.realtimeReconnecting");
    default:
      return t("pages.systemSettings.realtimeConnecting");
  }
});

const realtimeStateColor = computed(() => {
  switch (realtimeState.value) {
    case "connected":
      return "success";
    case "reconnecting":
      return "warning";
    default:
      return "info";
  }
});

function sortSettings(items: SystemSetting[]): SystemSetting[] {
  return [...items].sort((left, right) => left.key.localeCompare(right.key));
}

function replaceSettings(items: SystemSetting[]): void {
  settings.value = sortSettings(items);
  draftValues.value = {};
}

function upsertSetting(item: SystemSetting): void {
  const next = settings.value.filter((current) => current.key !== item.key);
  next.push(item);
  settings.value = sortSettings(next);
}

function clearDraftValue(settingKey: string): void {
  const next = { ...draftValues.value };
  delete next[settingKey];
  draftValues.value = next;
}

function currentBooleanValue(item: SystemSetting): boolean {
  if (Object.prototype.hasOwnProperty.call(draftValues.value, item.key)) {
    return draftValues.value[item.key];
  }
  return item.boolean_value;
}

function isBusy(settingKey: string): boolean {
  return savingKey.value === settingKey || resettingKey.value === settingKey;
}

function settingTitle(settingKey: string): string {
  return t(`pages.systemSettings.settings.${settingKey}.title`);
}

function settingDescription(settingKey: string): string {
  return t(`pages.systemSettings.settings.${settingKey}.description`);
}

function sectionLabel(section: string): string {
  return t(`pages.systemSettings.sections.${section}`);
}

function sourceLabel(source: string): string {
  return t(`pages.systemSettings.sources.${source}`);
}

function reloadLabel(reloadSemantics: string): string {
  return t(`pages.systemSettings.reload.${reloadSemantics}`);
}

function visibilityLabel(visibility: string): string {
  return t(`pages.systemSettings.visibility.${visibility}`);
}

function formatSettingDate(value: string | null | undefined): string {
  return formatDateTime(value, locale.value);
}

function updatedByLabel(item: SystemSetting): string {
  const email = String(item.updated_by_email || "").trim();
  if (email) return email;
  const userId = String(item.updated_by_user_id || "").trim();
  if (userId) return userId;
  return t("pages.systemSettings.noActor");
}

async function loadSettings(): Promise<void> {
  loading.value = true;
  try {
    replaceSettings(await listSystemSettings());
    error.value = null;
  } catch (err) {
    error.value = normalizeApiError(err);
  } finally {
    loading.value = false;
  }
}

function stopRealtime(): void {
  stopRealtimeRef.value?.();
  stopRealtimeRef.value = null;
  realtimeState.value = "closed";
}

function applyRealtimeMessage(message: SystemSettingsRealtimeMessage): void {
  if (message.type === "error") {
    if (settings.value.length === 0) {
      error.value = new LocalApiError({ kind: "unknown", messageKey: "errors.unknown" });
      loading.value = false;
    }
    return;
  }
  replaceSettings(message.items ?? []);
  error.value = null;
  loading.value = false;
}

function startRealtime(): void {
  stopRealtime();
  realtimeState.value = "connecting";
  stopRealtimeRef.value = subscribeSystemSettingsRealtime({
    onMessage: applyRealtimeMessage,
    onStateChange: (state) => {
      realtimeState.value = state;
    },
    onInitialMessageTimeout: () => {
      if (settings.value.length === 0) {
        error.value = new LocalApiError({ kind: "network", messageKey: "errors.realtimeUnavailable" });
        loading.value = false;
      }
    },
  });
}

async function refreshPage(): Promise<void> {
  await loadSettings();
  startRealtime();
}

async function handleBooleanToggle(item: SystemSetting, nextValue: boolean): Promise<void> {
  if (isBusy(item.key) || nextValue === item.boolean_value) {
    clearDraftValue(item.key);
    return;
  }

  draftValues.value = { ...draftValues.value, [item.key]: nextValue };
  savingKey.value = item.key;
  try {
    const updated = await updateSystemSettingBoolean(item.key, nextValue);
    upsertSetting(updated);
    clearDraftValue(item.key);
    snackbar.success(t("pages.systemSettings.saveSuccess"));
  } catch (err) {
    draftValues.value = { ...draftValues.value, [item.key]: item.boolean_value };
    snackbar.error(t(normalizeApiError(err).messageKey));
  } finally {
    savingKey.value = "";
  }
}

async function handleReset(item: SystemSetting): Promise<void> {
  if (isBusy(item.key)) return;

  resettingKey.value = item.key;
  try {
    const updated = await resetSystemSetting(item.key);
    upsertSetting(updated);
    clearDraftValue(item.key);
    snackbar.success(t("pages.systemSettings.resetSuccess"));
  } catch (err) {
    snackbar.error(t(normalizeApiError(err).messageKey));
  } finally {
    resettingKey.value = "";
  }
}

onMounted(() => {
  void refreshPage();
  stopLifecycleBindingRef.value = bindRealtimePageLifecycle({
    onResume: () => {
      void refreshPage();
    },
    onSuspend: () => {
      stopRealtime();
    },
  });
});

onBeforeUnmount(() => {
  stopLifecycleBindingRef.value?.();
  stopLifecycleBindingRef.value = null;
  stopRealtime();
});
</script>

<style scoped>
.system-settings-summary {
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(246, 248, 252, 0.98)),
    radial-gradient(280px 160px at 100% 0%, rgba(205, 231, 255, 0.45), transparent 60%);
}

.system-settings-card {
  border-radius: 28px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 1), rgba(248, 250, 252, 1)),
    radial-gradient(260px 140px at 100% 0%, rgba(225, 239, 255, 0.5), transparent 60%);
}
</style>
