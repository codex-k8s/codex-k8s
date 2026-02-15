<template>
  <div>
    <PageHeader :title="t('pages.agentDetails.title', { name: agentName })" :hint="t('pages.agentDetails.hint')">
      <template #leading>
        <VBtn variant="text" icon="mdi-arrow-left" :title="t('common.back')" :to="{ name: 'agents' }" />
      </template>
      <template #actions>
        <CopyChip :label="t('pages.agentDetails.agent')" :value="agentName" icon="mdi-robot-outline" />
      </template>
    </PageHeader>

    <VTabs v-model="tab" class="mt-4">
      <VTab value="settings">{{ t("pages.agentDetails.tabs.settings") }}</VTab>
      <VTab value="templates">{{ t("pages.agentDetails.tabs.templates") }}</VTab>
      <VTab value="history">{{ t("pages.agentDetails.tabs.history") }}</VTab>
    </VTabs>

    <VWindow v-model="tab" class="mt-2">
      <VWindowItem value="settings">
        <VAlert type="info" variant="tonal">
          <div class="text-body-2">{{ t("pages.agentDetails.scaffoldNote") }}</div>
        </VAlert>
      </VWindowItem>

      <VWindowItem value="templates">
        <VRow density="compact">
          <VCol cols="12" md="4">
            <VCard variant="outlined">
              <VCardTitle class="text-subtitle-1">{{ t("pages.agentDetails.templates.controlsTitle") }}</VCardTitle>
              <VCardText>
                <VBtnToggle v-model="selectedLocale" divided density="compact" mandatory>
                  <VBtn value="ru">ru</VBtn>
                  <VBtn value="en">en</VBtn>
                </VBtnToggle>

                <VSelect
                  v-model="templateKind"
                  class="mt-4"
                  :items="templateKindOptions"
                  :label="t('pages.agentDetails.templates.kind')"
                  hide-details
                />

                <VAlert type="warning" variant="tonal" class="mt-4">
                  <div class="text-body-2">{{ t("pages.agentDetails.templates.scaffoldWarning") }}</div>
                </VAlert>
              </VCardText>
            </VCard>
          </VCol>

          <VCol cols="12" md="8">
            <VCard variant="outlined">
              <VCardTitle class="text-subtitle-1 d-flex align-center justify-space-between ga-2 flex-wrap">
                {{ t("pages.agentDetails.templates.editorTitle") }}
                <VBtn variant="tonal" prepend-icon="mdi-content-save-outline" @click="mockSave">
                  {{ t("common.save") }}
                </VBtn>
              </VCardTitle>
              <VCardText>
                <MonacoEditor v-model="templateText" language="markdown" height="520px" />
              </VCardText>
            </VCard>
          </VCol>
        </VRow>

        <VRow class="mt-4" density="compact">
          <VCol cols="12" md="6">
            <VCard variant="outlined">
              <VCardTitle class="text-subtitle-2">{{ t("pages.agentDetails.templates.diffBase") }}</VCardTitle>
              <VCardText>
                <pre class="pre">{{ baseTemplate }}</pre>
              </VCardText>
            </VCard>
          </VCol>
          <VCol cols="12" md="6">
            <VCard variant="outlined">
              <VCardTitle class="text-subtitle-2">{{ t("pages.agentDetails.templates.diffEdited") }}</VCardTitle>
              <VCardText>
                <pre class="pre">{{ templateText }}</pre>
              </VCardText>
            </VCard>
          </VCol>
        </VRow>

        <VCard class="mt-4" variant="outlined">
          <VCardTitle class="text-subtitle-1">{{ t("pages.agentDetails.templates.effectiveTitle") }}</VCardTitle>
          <VCardText>
            <MonacoEditor v-model="effectiveTemplate" language="markdown" height="360px" read-only />
          </VCardText>
        </VCard>
      </VWindowItem>

      <VWindowItem value="history">
        <VAlert type="info" variant="tonal">
          <div class="text-body-2">{{ t("pages.agentDetails.history.scaffoldNote") }}</div>
        </VAlert>
      </VWindowItem>
    </VWindow>
  </div>
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные prompt templates (ru/en), diff/preview по версиям, effective template computation и историю изменений с аудитом.
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

import CopyChip from "../../shared/ui/CopyChip.vue";
import MonacoEditor from "../../shared/ui/monaco/MonacoEditor.vue";
import PageHeader from "../../shared/ui/PageHeader.vue";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";

const props = defineProps<{ agentName: string }>();

const { t } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();

const agentName = computed(() => props.agentName);

const tab = ref<"settings" | "templates" | "history">("templates");
const selectedLocale = ref<"ru" | "en">("ru");
const templateKind = ref<"work" | "review">("work");

const templateKindOptions = [
  { title: "work", value: "work" },
  { title: "review", value: "review" },
] as const;

const baseTemplates: Record<"ru" | "en", Record<"work" | "review", string>> = {
  en: {
    work: "# Work template\n\n- Goal\n- Constraints\n- Steps\n",
    review: "# Review template\n\n- Findings\n- Risks\n- Tests\n",
  },
  ru: {
    work: "# Шаблон работы\n\n- Цель\n- Ограничения\n- Шаги\n",
    review: "# Шаблон ревью\n\n- Замечания\n- Риски\n- Тесты\n",
  },
};

const baseTemplate = computed(() => baseTemplates[selectedLocale.value][templateKind.value]);
const templateText = ref(baseTemplate.value);

watch([selectedLocale, templateKind], () => {
  templateText.value = baseTemplate.value;
});

const effectiveTemplate = computed(() => {
  return `<!-- effective template (scaffold) -->\n\n${templateText.value}`;
});

function mockSave(): void {
  snackbar.success(t("common.saved"));
}
</script>

<style scoped>
.pre {
  margin: 0;
  white-space: pre-wrap;
  font-size: 12px;
  opacity: 0.95;
}
</style>
