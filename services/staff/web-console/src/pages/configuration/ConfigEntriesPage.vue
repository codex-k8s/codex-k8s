<template>
  <div>
    <PageHeader :title="t('pages.configEntries.title')" :hint="t('pages.configEntries.hint')">
      <template #actions>
        <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :loading="store.loading" @click="load" />
      </template>
    </PageHeader>

    <VAlert v-if="store.error" type="error" variant="tonal" class="mt-4">
      {{ t(store.error.messageKey) }}
    </VAlert>
    <VAlert v-if="store.saveError" type="error" variant="tonal" class="mt-4">
      {{ t(store.saveError.messageKey) }}
    </VAlert>
    <VAlert v-if="store.deleteError" type="error" variant="tonal" class="mt-4">
      {{ t(store.deleteError.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardTitle class="text-subtitle-1">{{ t("pages.configEntries.filtersTitle") }}</VCardTitle>
      <VCardText>
        <VRow density="compact" class="align-end">
          <VCol cols="12" md="3">
            <VSelect v-model="scope" :items="scopeItems" :label="t('pages.configEntries.scope')" hide-details @update:model-value="load" />
          </VCol>
          <VCol cols="12" md="3">
            <VTextField v-model.trim="projectId" :label="t('pages.configEntries.projectId')" hide-details />
          </VCol>
          <VCol cols="12" md="3">
            <VTextField v-model.trim="repositoryId" :label="t('pages.configEntries.repositoryId')" hide-details />
          </VCol>
          <VCol cols="12" md="3">
            <AdaptiveBtn variant="tonal" icon="mdi-magnify" :label="t('common.refresh')" :loading="store.loading" @click="load" />
          </VCol>
        </VRow>
        <div class="text-caption text-medium-emphasis mt-2">
          {{ t("pages.configEntries.syncTargetsHint") }}
        </div>
      </VCardText>
    </VCard>

    <VCard class="mt-4" variant="outlined">
      <VCardTitle class="text-subtitle-1">{{ t("pages.configEntries.listTitle") }}</VCardTitle>
      <VCardText>
        <VDataTable :headers="headers" :items="store.items" :loading="store.loading" :items-per-page="10" hover>
          <template #item.value="{ item }">
            <span class="mono text-medium-emphasis">{{ item.kind === "secret" ? "********" : item.value ?? "-" }}</span>
          </template>
          <template #item.sync_targets="{ item }">
            <div class="d-flex justify-center flex-wrap ga-1">
              <VChip v-for="tgt in item.sync_targets" :key="tgt" size="x-small" variant="tonal" color="secondary">
                <span class="mono">{{ tgt }}</span>
              </VChip>
            </div>
          </template>
          <template #item.updated_at="{ item }">
            <span class="text-medium-emphasis">{{ fmtDateTime(item.updated_at) }}</span>
          </template>
          <template #item.actions="{ item }">
            <div class="d-flex justify-end">
              <VTooltip :text="t('common.delete')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    color="error"
                    variant="tonal"
                    icon="mdi-delete-outline"
                    :loading="store.deleting"
                    @click="askDelete(item.id, item.key)"
                  />
                </template>
              </VTooltip>
            </div>
          </template>
          <template #no-data>
            <div class="py-8 text-medium-emphasis">
              {{ t("states.noData") }}
            </div>
          </template>
        </VDataTable>
      </VCardText>
    </VCard>

    <VCard class="mt-6" variant="outlined">
      <VCardTitle class="text-subtitle-1">{{ t("pages.configEntries.upsertTitle") }}</VCardTitle>
      <VCardText>
        <VRow density="compact" class="align-end">
          <VCol cols="12" md="3">
            <VSelect v-model="formScope" :items="scopeItems" :label="t('pages.configEntries.scope')" hide-details />
          </VCol>
          <VCol cols="12" md="3">
            <VSelect v-model="kind" :items="kindItems" :label="t('pages.configEntries.kind')" hide-details />
          </VCol>
          <VCol cols="12" md="3">
            <VSelect v-model="mutability" :items="mutabilityItems" :label="t('pages.configEntries.mutability')" hide-details />
          </VCol>
          <VCol cols="12" md="3">
            <VSwitch v-model="isDangerous" :label="t('pages.configEntries.isDangerous')" hide-details />
          </VCol>
          <VCol cols="12" md="6">
            <VTextField v-model.trim="formProjectId" :label="t('pages.configEntries.projectId')" hide-details />
          </VCol>
          <VCol cols="12" md="6">
            <VTextField v-model.trim="formRepositoryId" :label="t('pages.configEntries.repositoryId')" hide-details />
          </VCol>
          <VCol cols="12">
            <VTextField v-model.trim="key" :label="t('pages.configEntries.key')" hide-details />
          </VCol>
          <VCol cols="12">
            <VTextField
              v-if="kind === 'variable'"
              v-model="valuePlain"
              :label="t('pages.configEntries.valuePlain')"
              hide-details
            />
            <VTextField
              v-else
              v-model="valueSecret"
              :label="t('pages.configEntries.valueSecret')"
              type="password"
              hide-details
            />
          </VCol>
          <VCol cols="12">
            <VTextField v-model="syncTargetsRaw" :label="t('pages.configEntries.syncTargets')" hide-details />
          </VCol>
          <VCol cols="12">
            <VAlert v-if="isDangerous" type="warning" variant="tonal">
              <div class="text-body-2">{{ t("pages.configEntries.dangerousWarning") }}</div>
            </VAlert>
            <VCheckbox
              v-if="isDangerous"
              v-model="dangerousConfirmed"
              class="mt-2"
              density="compact"
              :label="t('pages.configEntries.dangerousConfirm')"
              hide-details
            />
          </VCol>
          <VCol cols="12">
            <VBtn color="primary" variant="tonal" :loading="store.saving" @click="save">
              {{ t("common.save") }}
            </VBtn>
          </VCol>
        </VRow>
      </VCardText>
    </VCard>
  </div>

  <ConfirmDialog
    v-model="confirmDeleteOpen"
    :title="t('common.delete')"
    :message="confirmDeleteLabel"
    :confirm-text="t('common.delete')"
    :cancel-text="t('common.cancel')"
    danger
    @confirm="doDelete"
  />
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import ConfirmDialog from "../../shared/ui/ConfirmDialog.vue";
import PageHeader from "../../shared/ui/PageHeader.vue";
import { formatDateTime } from "../../shared/lib/datetime";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { useConfigEntriesStore } from "../../features/config/config-entries-store";

const { t, locale } = useI18n({ useScope: "global" });
const store = useConfigEntriesStore();
const snackbar = useSnackbarStore();

const scopeItems = ["platform", "project", "repository"] as const;
const kindItems = ["variable", "secret"] as const;
const mutabilityItems = ["startup_required", "runtime_mutable"] as const;

const scope = ref<(typeof scopeItems)[number]>("platform");
const projectId = ref("");
const repositoryId = ref("");

const headers = [
  { title: t("pages.configEntries.scope"), key: "scope", width: 140, align: "center" },
  { title: t("pages.configEntries.kind"), key: "kind", width: 140, align: "center" },
  { title: t("pages.configEntries.key"), key: "key", align: "start" },
  { title: t("pages.configEntries.value"), key: "value", align: "center" },
  { title: t("pages.configEntries.mutability"), key: "mutability", width: 200, align: "center" },
  { title: t("pages.configEntries.isDangerous"), key: "is_dangerous", width: 140, align: "center" },
  { title: t("pages.configEntries.syncTargets"), key: "sync_targets", align: "center" },
  { title: t("pages.configEntries.updatedAt"), key: "updated_at", width: 220, align: "center" },
  { title: "", key: "actions", sortable: false, width: 72, align: "end" },
] as const;

function fmtDateTime(value: string | null | undefined): string {
  return formatDateTime(value, locale.value);
}

async function load() {
  await store.load({
    scope: scope.value,
    projectId: projectId.value.trim() || undefined,
    repositoryId: repositoryId.value.trim() || undefined,
    limit: 200,
  });
}

const formScope = ref<(typeof scopeItems)[number]>("platform");
const kind = ref<(typeof kindItems)[number]>("variable");
const mutability = ref<(typeof mutabilityItems)[number]>("startup_required");
const isDangerous = ref(false);
const dangerousConfirmed = ref(false);
const formProjectId = ref("");
const formRepositoryId = ref("");
const key = ref("");
const valuePlain = ref("");
const valueSecret = ref("");
const syncTargetsRaw = ref("");

function parseTargets(raw: string): string[] {
  const parts = raw.split(",").map((p) => p.trim()).filter((p) => p.length > 0);
  return Array.from(new Set(parts));
}

async function save() {
  if (isDangerous.value && !dangerousConfirmed.value) {
    snackbar.error(t("pages.configEntries.dangerousConfirmRequired"));
    return;
  }
  const item = await store.upsert({
    scope: formScope.value,
    kind: kind.value,
    projectId: formProjectId.value.trim() ? formProjectId.value.trim() : null,
    repositoryId: formRepositoryId.value.trim() ? formRepositoryId.value.trim() : null,
    key: key.value.trim(),
    valuePlain: kind.value === "variable" ? (valuePlain.value === "" ? null : valuePlain.value) : null,
    valueSecret: kind.value === "secret" ? (valueSecret.value === "" ? null : valueSecret.value) : null,
    syncTargets: parseTargets(syncTargetsRaw.value),
    mutability: mutability.value,
    isDangerous: isDangerous.value,
    dangerousConfirmed: dangerousConfirmed.value,
  });
  if (item) {
    snackbar.success(t("common.saved"));
    await load();
    valuePlain.value = "";
    valueSecret.value = "";
    dangerousConfirmed.value = false;
  }
}

const confirmDeleteOpen = ref(false);
const confirmDeleteId = ref("");
const confirmDeleteLabel = ref("");

function askDelete(id: string, label: string) {
  confirmDeleteId.value = id;
  confirmDeleteLabel.value = label;
  confirmDeleteOpen.value = true;
}

async function doDelete() {
  const id = confirmDeleteId.value;
  confirmDeleteId.value = "";
  if (!id) return;
  await store.remove(id);
  if (!store.deleteError) {
    snackbar.success(t("common.deleted"));
    await load();
  }
}

onMounted(() => void load());
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
