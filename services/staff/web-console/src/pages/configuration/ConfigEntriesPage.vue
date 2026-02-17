<template>
  <div>
    <PageHeader :title="t('pages.configEntries.title')" :hint="t('pages.configEntries.hint')">
      <template #actions>
        <AdaptiveBtn
          color="primary"
          variant="tonal"
          icon="mdi-plus"
          :label="t('pages.configEntries.add')"
          :disabled="store.loading"
          class="mr-2"
          @click="openCreateDialog"
        />
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
            <VSelect
              v-model="scope"
              :items="scopeItems"
              :label="t('pages.configEntries.scope')"
              hide-details
            />
          </VCol>
          <VCol cols="12" md="4">
            <VSelect
              v-model="projectId"
              :items="projectOptions"
              :label="t('context.project')"
              hide-details
              clearable
              :disabled="!projectFilterEnabled"
            />
          </VCol>
          <VCol cols="12" md="4">
            <VSelect
              v-model="repositoryId"
              :items="repositoryOptions"
              :label="t('pages.configEntries.repository')"
              hide-details
              clearable
              :loading="repositoriesLoading"
              :disabled="!repositoryFilterEnabled"
            />
          </VCol>
          <VCol cols="12" md="1" class="d-flex justify-end">
            <VTooltip :text="t('common.reset')">
              <template #activator="{ props: tipProps }">
                <VBtn
                  v-bind="tipProps"
                  icon="mdi-backspace-outline"
                  variant="text"
                  :disabled="store.loading"
                  @click="resetFilters"
                />
              </template>
            </VTooltip>
          </VCol>
        </VRow>
      </VCardText>
    </VCard>

    <VCard class="mt-4" variant="outlined">
      <VCardTitle class="text-subtitle-1">{{ t("pages.configEntries.listTitle") }}</VCardTitle>
      <VCardText>
        <VDataTable :headers="headers" :items="store.items" :loading="store.loading" :items-per-page="10" hover>
          <template #item.scope="{ item }">
            <div class="d-flex justify-center">
              <VChip size="x-small" variant="tonal" class="font-weight-bold" color="secondary">
                {{ scopeLabel(item.scope) }}
              </VChip>
            </div>
          </template>
          <template #item.value="{ item }">
            <template v-if="item.kind === 'secret'">
              <span class="mono text-medium-emphasis">********</span>
            </template>
            <template v-else>
              <div class="d-flex justify-center">
                <div class="value-cell">
                  <div
                    class="mono text-medium-emphasis value-preview"
                    :class="{ 'value-preview--collapsed': !isValueExpanded(item.id) }"
                  >
                    {{ item.value ?? "-" }}
                  </div>
                  <VBtn
                    v-if="isMultilineValue(item.value)"
                    size="x-small"
                    variant="text"
                    class="value-toggle"
                    @click="toggleValueExpanded(item.id)"
                  >
                    {{ isValueExpanded(item.id) ? t("common.collapse") : t("common.expand") }}
                  </VBtn>
                </div>
              </div>
            </template>
          </template>
          <template #item.mutability="{ item }">
            <span class="text-medium-emphasis">{{ mutabilityLabel(item.mutability) }}</span>
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
              <VTooltip :text="t('scaffold.rowActions.edit')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    variant="text"
                    icon="mdi-pencil-outline"
                    :disabled="store.saving || store.deleting"
                    @click="openEditDialog(item)"
                  />
                </template>
              </VTooltip>
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

  <VDialog v-model="upsertDialogOpen" max-width="760">
    <VCard>
      <VCardTitle class="text-subtitle-1">
        {{ upsertMode === "create" ? t("pages.configEntries.createTitle") : t("pages.configEntries.editTitle") }}
      </VCardTitle>
      <VCardText>
        <VRow density="compact" class="align-end">
          <VCol cols="12" md="4">
            <VSelect v-model="formScope" :items="formScopeItems" :label="t('pages.configEntries.scope')" hide-details />
          </VCol>
          <VCol cols="12" md="4">
            <VSelect v-model="kind" :items="kindItems" :label="t('pages.configEntries.kind')" hide-details />
          </VCol>
          <VCol cols="12" md="4">
            <VSelect v-model="mutability" :items="mutabilityItems" :label="t('pages.configEntries.mutability')" hide-details />
          </VCol>
          <VCol cols="12" md="4">
            <VSwitch v-model="isDangerous" :label="t('pages.configEntries.isDangerous')" hide-details />
          </VCol>
          <VCol cols="12" md="4">
            <VSelect
              v-model="formProjectId"
              :items="projectOptions"
              :label="t('context.project')"
              hide-details
              clearable
              :disabled="formScope === 'platform'"
            />
          </VCol>
          <VCol cols="12" md="4">
            <VSelect
              v-model="formRepositoryId"
              :items="repositoryOptions"
              :label="t('pages.configEntries.repository')"
              hide-details
              clearable
              :loading="repositoriesLoading"
              :disabled="formScope !== 'repository'"
            />
          </VCol>
          <VCol cols="12">
            <VTextField v-model.trim="key" :label="t('pages.configEntries.key')" hide-details />
          </VCol>
          <VCol cols="12">
            <VTextarea
              v-if="kind === 'variable'"
              v-model="valuePlain"
              :label="t('pages.configEntries.valuePlain')"
              rows="5"
              auto-grow
              hide-details
            />
            <VTextarea
              v-else
              v-model="valueSecret"
              :label="t('pages.configEntries.valueSecret')"
              rows="5"
              auto-grow
              hide-details
              :class="{ 'secret-masked': !showSecretValue }"
              :append-inner-icon="showSecretValue ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
              @click:append-inner="showSecretValue = !showSecretValue"
            />
          </VCol>
          <VCol cols="12">
            <VTextField v-model="syncTargetsRaw" :label="t('pages.configEntries.syncTargets')" hide-details />
            <div class="text-caption text-medium-emphasis mt-2">
              {{ t("pages.configEntries.syncTargetsHint") }}
            </div>
          </VCol>
          <VCol cols="12">
            <VAlert v-if="isDangerous" :type="dangerousAlertType" variant="tonal">
              <div class="text-body-2">{{ dangerousAlertText }}</div>
            </VAlert>
            <VCheckbox
              v-if="isDangerous"
              v-model="dangerousConfirmed"
              class="mt-2"
              density="compact"
              :label="dangerousConfirmText"
              hide-details
            />
          </VCol>
          <VCol cols="12">
            <VAlert v-if="store.saveError" type="error" variant="tonal">
              {{ t(store.saveError.messageKey) }}
            </VAlert>
          </VCol>
        </VRow>
      </VCardText>
      <VCardActions>
        <VSpacer />
        <VBtn variant="text" :disabled="store.saving" @click="upsertDialogOpen = false">{{ t("common.cancel") }}</VBtn>
        <VBtn color="primary" variant="tonal" :disabled="!canSave" :loading="store.saving" @click="save">
          {{ t("common.save") }}
        </VBtn>
      </VCardActions>
    </VCard>
  </VDialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import ConfirmDialog from "../../shared/ui/ConfirmDialog.vue";
import PageHeader from "../../shared/ui/PageHeader.vue";
import { formatDateTime } from "../../shared/lib/datetime";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { useConfigEntriesStore } from "../../features/config/config-entries-store";
import { useProjectsStore } from "../../features/projects/projects-store";
import { useUiContextStore } from "../../features/ui-context/store";
import { listProjectRepositories } from "../../features/projects/api";
import type { RepositoryBinding } from "../../features/projects/types";
import type { ConfigEntry } from "../../features/config/types";

type ScopeFilter = "all" | "platform" | "project" | "repository";
type Scope = Exclude<ScopeFilter, "all">;
type Kind = "variable" | "secret";
type Mutability = "startup_required" | "runtime_mutable";

const { t, locale } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();
const store = useConfigEntriesStore();
const projects = useProjectsStore();
const uiContext = useUiContextStore();

const scope = ref<ScopeFilter>("platform");
const projectId = ref(uiContext.projectId || "");
const repositoryId = ref("");

const repositories = ref<RepositoryBinding[]>([]);
const repositoriesLoading = ref(false);

const scopeItems = computed(() => ([
  { title: t("pages.configEntries.scopeAll"), value: "all" },
  { title: t("pages.configEntries.scopePlatform"), value: "platform" },
  { title: t("pages.configEntries.scopeProject"), value: "project" },
  { title: t("pages.configEntries.scopeRepository"), value: "repository" },
] as const));

const projectOptions = computed(() => ([
  { title: t("context.allObjects"), value: "" },
  ...projects.items.map((p) => ({ title: p.name || p.slug || p.id, value: p.id })),
]));

const repositoryOptions = computed(() => ([
  { title: t("context.allObjects"), value: "" },
  ...repositories.value.map((r) => ({ title: `${r.owner}/${r.name}`, value: r.id })),
]));

const projectFilterEnabled = computed(() => scope.value !== "platform");
const repositoryFilterEnabled = computed(() =>
  (scope.value === "repository" || scope.value === "all") && Boolean(projectId.value),
);

watch(
  () => projectId.value,
  async (next) => {
    repositories.value = [];
    repositoryId.value = "";
    const trimmed = String(next || "").trim();
    if (!trimmed) return;

    repositoriesLoading.value = true;
    try {
      repositories.value = await listProjectRepositories(trimmed);
    } catch (e) {
      snackbar.error(t("errors.unknown"));
    } finally {
      repositoriesLoading.value = false;
    }
  },
  { immediate: true },
);

watch(
  () => scope.value,
  (next) => {
    if (next === "platform") {
      projectId.value = "";
      repositoryId.value = "";
      return;
    }
    if (next === "project") {
      repositoryId.value = "";
    }
  },
);

const headers = computed(() => ([
  { title: t("pages.configEntries.scope"), key: "scope", width: 140, align: "center" },
  { title: t("pages.configEntries.kind"), key: "kind", width: 140, align: "center" },
  { title: t("pages.configEntries.key"), key: "key", align: "start" },
  { title: t("pages.configEntries.value"), key: "value", align: "center" },
  { title: t("pages.configEntries.mutability"), key: "mutability", width: 200, align: "center" },
  { title: t("pages.configEntries.isDangerous"), key: "is_dangerous", width: 140, align: "center" },
  { title: t("pages.configEntries.syncTargets"), key: "sync_targets", align: "center" },
  { title: t("pages.configEntries.updatedAt"), key: "updated_at", width: 220, align: "center" },
  { title: "", key: "actions", sortable: false, width: 88, align: "end" },
] as const));

function fmtDateTime(value: string | null | undefined): string {
  return formatDateTime(value, locale.value);
}

function scopeLabel(value: string): string {
  switch (String(value || "")) {
    case "platform":
      return t("pages.configEntries.scopePlatform");
    case "project":
      return t("pages.configEntries.scopeProject");
    case "repository":
      return t("pages.configEntries.scopeRepository");
    default:
      return String(value || "-");
  }
}

function mutabilityLabel(value: string): string {
  switch (String(value || "")) {
    case "startup_required":
      return t("pages.configEntries.mutabilityStartupRequired");
    case "runtime_mutable":
      return t("pages.configEntries.mutabilityRuntimeMutable");
    default:
      return String(value || "-");
  }
}

async function load(): Promise<void> {
  if (scope.value === "project" && !projectId.value) {
    store.items = [];
    store.error = null;
    return;
  }
  if (scope.value === "repository" && !repositoryId.value) {
    store.items = [];
    store.error = null;
    return;
  }
  const limit = 200;
  switch (scope.value) {
    case "platform":
      await store.load({ scope: "platform", limit });
      return;
    case "project":
      await store.load({ scope: "project", projectId: projectId.value, limit });
      return;
    case "repository":
      await store.load({ scope: "repository", repositoryId: repositoryId.value, limit });
      return;
    default:
      await store.load({
        scope: "all",
        projectId: projectId.value || undefined,
        repositoryId: repositoryId.value || undefined,
        limit,
      });
      return;
  }
}

function resetFilters(): void {
  scope.value = "platform";
  projectId.value = "";
  repositoryId.value = "";
}

const confirmDeleteOpen = ref(false);
const confirmDeleteId = ref("");
const confirmDeleteLabel = ref("");

const expandedValueByID = ref<Record<string, boolean>>({});

function isMultilineValue(value: string | null | undefined): boolean {
  return String(value || "").includes("\n");
}

function isValueExpanded(id: string): boolean {
  return Boolean(expandedValueByID.value[id]);
}

function toggleValueExpanded(id: string): void {
  expandedValueByID.value[id] = !expandedValueByID.value[id];
}

function askDelete(id: string, label: string): void {
  confirmDeleteId.value = id;
  confirmDeleteLabel.value = label;
  confirmDeleteOpen.value = true;
}

async function doDelete(): Promise<void> {
  const id = confirmDeleteId.value;
  confirmDeleteId.value = "";
  if (!id) return;
  await store.remove(id);
  if (!store.deleteError) {
    snackbar.success(t("common.deleted"));
    await load();
  }
}

const upsertDialogOpen = ref(false);
const upsertMode = ref<"create" | "edit">("create");

const formScope = ref<Scope>("platform");
const kind = ref<Kind>("variable");
const mutability = ref<Mutability>("startup_required");
const isDangerous = ref(false);
const dangerousConfirmed = ref(false);
const formProjectId = ref("");
const formRepositoryId = ref("");
const key = ref("");
const valuePlain = ref("");
const valueSecret = ref("");
const showSecretValue = ref(false);
const syncTargetsRaw = ref("");

const dangerousAlertType = computed(() => (upsertMode.value === "create" ? "info" : "warning"));
const dangerousAlertText = computed(() =>
  upsertMode.value === "create"
    ? t("pages.configEntries.dangerousWarningCreate")
    : t("pages.configEntries.dangerousWarningEdit"),
);
const dangerousConfirmText = computed(() =>
  upsertMode.value === "create"
    ? t("pages.configEntries.dangerousConfirmCreate")
    : t("pages.configEntries.dangerousConfirmEdit"),
);

const formScopeItems = computed(() => ([
  { title: t("pages.configEntries.scopePlatform"), value: "platform" },
  { title: t("pages.configEntries.scopeProject"), value: "project" },
  { title: t("pages.configEntries.scopeRepository"), value: "repository" },
] as const));

const kindItems = computed(() => ([
  { title: t("pages.configEntries.kindVariable"), value: "variable" },
  { title: t("pages.configEntries.kindSecret"), value: "secret" },
] as const));

const mutabilityItems = computed(() => ([
  { title: t("pages.configEntries.mutabilityStartupRequired"), value: "startup_required" },
  { title: t("pages.configEntries.mutabilityRuntimeMutable"), value: "runtime_mutable" },
] as const));

watch(
  () => formScope.value,
  (next) => {
    if (next === "platform") {
      formProjectId.value = "";
      formRepositoryId.value = "";
      return;
    }
    if (next === "project") {
      formRepositoryId.value = "";
    }
  },
);

function parseTargets(raw: string): string[] {
  const parts = raw
    .split(",")
    .map((p) => p.trim())
    .filter((p) => p.length > 0);
  return Array.from(new Set(parts));
}

const canSave = computed(() => {
  const k = key.value.trim();
  if (!k) return false;
  if (formScope.value === "project" && !formProjectId.value) return false;
  if (formScope.value === "repository" && !formRepositoryId.value) return false;
  if (kind.value === "secret" && valueSecret.value.trim() === "") return false;
  if (isDangerous.value && !dangerousConfirmed.value) return false;
  return true;
});

function openCreateDialog(): void {
  upsertMode.value = "create";
  store.saveError = null;
  formScope.value = scope.value === "repository" ? "repository" : scope.value === "project" ? "project" : "platform";
  kind.value = "variable";
  mutability.value = "startup_required";
  isDangerous.value = false;
  dangerousConfirmed.value = false;
  formProjectId.value = projectId.value;
  formRepositoryId.value = repositoryId.value;
  key.value = "";
  valuePlain.value = "";
  valueSecret.value = "";
  showSecretValue.value = false;
  syncTargetsRaw.value = "";
  upsertDialogOpen.value = true;
}

function openEditDialog(item: ConfigEntry): void {
  upsertMode.value = "edit";
  store.saveError = null;
  formScope.value = String(item.scope || "platform") as Scope;
  kind.value = String(item.kind || "variable") as Kind;
  mutability.value = String(item.mutability || "startup_required") as Mutability;
  isDangerous.value = Boolean(item.is_dangerous);
  dangerousConfirmed.value = false;
  formProjectId.value = String(item.project_id || "");
  formRepositoryId.value = String(item.repository_id || "");
  key.value = String(item.key || "");
  valuePlain.value = kind.value === "variable" ? String(item.value || "") : "";
  valueSecret.value = "";
  showSecretValue.value = false;
  syncTargetsRaw.value = (item.sync_targets || []).join(", ");
  upsertDialogOpen.value = true;
}

async function save(): Promise<void> {
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
  if (!item) return;

  snackbar.success(t("common.saved"));
  upsertDialogOpen.value = false;
  await load();
}

watch(
  () => [scope.value, projectId.value, repositoryId.value] as const,
  () => void load(),
  { immediate: true },
);

onMounted(() => {
  if (projects.items.length === 0) void projects.load();
});
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.value-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  max-width: 720px;
}

.value-preview {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.value-preview--collapsed {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.value-toggle {
  margin-top: 2px;
}

.secret-masked :deep(textarea) {
  -webkit-text-security: disc;
}
</style>
