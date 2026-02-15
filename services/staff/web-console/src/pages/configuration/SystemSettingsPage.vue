<template>
  <div>
    <PageHeader :title="t('pages.systemSettings.title')" :hint="t('pages.systemSettings.hint')">
      <template #actions>
        <VBtn variant="tonal" prepend-icon="mdi-plus" @click="addLocaleDialogOpen = true">
          {{ t("pages.systemSettings.addLocale") }}
        </VBtn>
      </template>
    </PageHeader>

    <VRow class="mt-4" density="compact">
      <VCol cols="12" md="6">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-2">{{ t("pages.systemSettings.localesTitle") }}</VCardTitle>
          <VCardText>
            <VDataTable :headers="localeHeaders" :items="locales" :items-per-page="10" density="comfortable" />
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
import { ref } from "vue";
import { useI18n } from "vue-i18n";

import PageHeader from "../../shared/ui/PageHeader.vue";

type LocaleRow = {
  code: string;
  name: string;
  is_default: boolean;
};

const { t } = useI18n({ useScope: "global" });

const localeHeaders = [
  { title: "code", key: "code", width: 120 },
  { title: "name", key: "name" },
  { title: "default", key: "is_default", width: 120 },
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
</script>

