<template>
  <div>
    <PageHeader :title="t('pages.projects.title')">
      <template #actions>
        <AdaptiveBtn
          v-if="auth.isPlatformAdmin"
          class="mr-2"
          color="primary"
          variant="tonal"
          icon="mdi-plus"
          :label="t('pages.projects.createProject')"
          :disabled="projects.loading"
          @click="openCreateDialog"
        />
        <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :loading="projects.loading" @click="refreshProjects" />
      </template>
    </PageHeader>

    <VAlert v-if="projects.error" type="error" variant="tonal" class="mt-4">
      {{ t(projects.error.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable v-model:page="tablePage" :headers="headers" :items="projects.items" :loading="projects.loading" :items-per-page="itemsPerPage" hover>
          <template #item.name="{ item }">
            <div class="d-flex justify-center">
              <RouterLink class="text-primary font-weight-bold text-decoration-none" :to="{ name: 'project-details', params: { projectId: item.id } }">
                {{ item.name }}
              </RouterLink>
            </div>
          </template>

          <template #item.select="{ item }">
            <div class="d-flex justify-center">
              <VTooltip :text="t('pages.projects.makeCurrent')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    variant="tonal"
                    :color="uiContext.projectId === item.id ? 'success' : 'primary'"
                    :icon="uiContext.projectId === item.id ? 'mdi-check-circle-outline' : 'mdi-check'"
                    :disabled="uiContext.projectId === item.id"
                    @click="selectProject(item.id)"
                  />
                </template>
              </VTooltip>
            </div>
          </template>

          <template #item.role="{ item }">
            <div class="d-flex justify-center">
              <VChip size="small" variant="tonal" class="font-weight-bold" :color="colorForProjectRole(item.role)">
                {{ roleLabel(item.role) }}
              </VChip>
            </div>
          </template>

          <template #item.manage="{ item }">
            <div class="d-flex ga-2 justify-center flex-wrap">
              <VTooltip v-if="auth.isPlatformAdmin" :text="t('scaffold.rowActions.edit')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    variant="text"
                    icon="mdi-pencil-outline"
                    :disabled="projects.saving"
                    @click="openEditDialog(item.slug, item.name)"
                  />
                </template>
              </VTooltip>
              <VTooltip :text="t('pages.projects.repos')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    variant="text"
                    icon="mdi-source-repository"
                    :to="{ name: 'project-repositories', params: { projectId: item.id } }"
                  />
                </template>
              </VTooltip>
              <VTooltip :text="t('pages.projects.members')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    variant="text"
                    icon="mdi-account-group-outline"
                    :to="{ name: 'project-members', params: { projectId: item.id } }"
                  />
                </template>
              </VTooltip>
            </div>
          </template>

          <template #item.actions="{ item }">
            <div class="d-flex justify-end">
              <VTooltip v-if="auth.isPlatformOwner" :text="t('common.delete')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    color="error"
                    variant="tonal"
                    icon="mdi-delete-outline"
                    :loading="projects.deleting"
                    @click="askDelete(item.id, item.name)"
                  />
                </template>
              </VTooltip>
            </div>
          </template>

          <template #no-data>
            <div class="py-8 text-medium-emphasis">
              {{ t("states.noProjects") }}
            </div>
          </template>
        </VDataTable>
      </VCardText>
    </VCard>
  </div>

  <VDialog v-model="upsertDialogOpen" max-width="720">
    <VCard>
      <VCardTitle class="text-subtitle-1">
        {{ upsertMode === "create" ? t("pages.projects.createProject") : t("pages.projects.editProject") }}
      </VCardTitle>
      <VCardText>
        <VRow density="compact">
          <VCol cols="12" md="4">
            <VTextField
              v-model.trim="formSlug"
              :label="t('pages.projects.slug')"
              :placeholder="t('placeholders.projectSlug')"
              :disabled="upsertMode === 'edit'"
              hide-details
            />
          </VCol>
          <VCol cols="12" md="8">
            <VTextField v-model.trim="formName" :label="t('pages.projects.name')" :placeholder="t('placeholders.projectName')" hide-details />
          </VCol>
        </VRow>

        <VAlert v-if="projects.saveError" type="error" variant="tonal" class="mt-4">
          {{ t(projects.saveError.messageKey) }}
        </VAlert>
        <VAlert v-if="projects.deleteError" type="error" variant="tonal" class="mt-4">
          {{ t(projects.deleteError.messageKey) }}
        </VAlert>
      </VCardText>
      <VCardActions>
        <VSpacer />
        <VBtn variant="text" :disabled="projects.saving" @click="upsertDialogOpen = false">{{ t("common.cancel") }}</VBtn>
        <VBtn color="primary" variant="tonal" :disabled="!canSave" :loading="projects.saving" @click="saveProject">
          {{ t("common.save") }}
        </VBtn>
      </VCardActions>
    </VCard>
  </VDialog>

  <ConfirmDialog
    v-model="confirmOpen"
    :title="t('common.delete')"
    :message="confirmName"
    :confirm-text="t('common.delete')"
    :cancel-text="t('common.cancel')"
    danger
    @confirm="doDelete"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import PageHeader from "../shared/ui/PageHeader.vue";
import AdaptiveBtn from "../shared/ui/AdaptiveBtn.vue";
import ConfirmDialog from "../shared/ui/ConfirmDialog.vue";
import { useSnackbarStore } from "../shared/ui/feedback/snackbar-store";
import { useAuthStore } from "../features/auth/store";
import { useProjectsStore } from "../features/projects/projects-store";
import { useUiContextStore } from "../features/ui-context/store";
import { colorForProjectRole } from "../shared/lib/chips";
import { createProgressiveTableState } from "../shared/lib/progressive-table";

const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();
const projects = useProjectsStore();
const snackbar = useSnackbarStore();
const uiContext = useUiContextStore();
const itemsPerPage = 10;
const paging = createProgressiveTableState({ itemsPerPage });
const tablePage = paging.page;

const confirmOpen = ref(false);
const confirmProjectId = ref<string>("");
const confirmName = ref<string>("");

const headers = computed(() => ([
  { title: t("pages.projects.slug"), key: "slug", width: 220, align: "start" },
  { title: t("pages.projects.name"), key: "name", align: "center" },
  { title: "", key: "select", width: 72, sortable: false, align: "center" },
  { title: t("pages.projects.role"), key: "role", width: 160, sortable: false, align: "center" },
  { title: t("pages.projects.manage"), key: "manage", width: 180, sortable: false, align: "center" },
  { title: "", key: "actions", sortable: false, width: 72, align: "end" },
]) as const);

function roleLabel(role: string): string {
  const normalized = role.trim();
  if (normalized === "read") return t("roles.read");
  if (normalized === "read_write") return t("roles.readWrite");
  if (normalized === "admin") return t("roles.admin");
  return normalized;
}

async function load() {
  await projects.load(paging.limit.value);
  paging.markLoaded(projects.items.length);
}

async function refreshProjects(): Promise<void> {
  paging.reset();
  await load();
}

async function loadMoreProjectsIfNeeded(nextPage: number, prevPage: number): Promise<void> {
  if (projects.loading) {
    return;
  }
  if (!paging.shouldGrowForPage(projects.items.length, nextPage, prevPage)) {
    return;
  }
  await load();
}

function selectProject(projectId: string): void {
  uiContext.setProjectId(projectId);
}

function askDelete(projectId: string, projectName: string) {
  confirmProjectId.value = projectId;
  confirmName.value = projectName;
  confirmOpen.value = true;
}

async function doDelete() {
  const id = confirmProjectId.value;
  confirmOpen.value = false;
  confirmProjectId.value = "";
  if (!id) return;
  await projects.remove(id, paging.limit.value);
  paging.markLoaded(projects.items.length);
  if (!projects.deleteError) {
    snackbar.success(t("common.deleted"));
  }
}

watch(
  tablePage,
  (nextPage, prevPage) => void loadMoreProjectsIfNeeded(nextPage, prevPage),
);

onMounted(() => void refreshProjects());

const upsertDialogOpen = ref(false);
const upsertMode = ref<"create" | "edit">("create");
const formSlug = ref("");
const formName = ref("");
const initialSlug = ref("");
const initialName = ref("");

function openCreateDialog(): void {
  upsertMode.value = "create";
  formSlug.value = "";
  formName.value = "";
  initialSlug.value = "";
  initialName.value = "";
  projects.saveError = null;
  upsertDialogOpen.value = true;
}

function openEditDialog(slug: string, name: string): void {
  upsertMode.value = "edit";
  formSlug.value = slug;
  formName.value = name;
  initialSlug.value = slug;
  initialName.value = name;
  projects.saveError = null;
  upsertDialogOpen.value = true;
}

const canSave = computed(() => {
  const slug = formSlug.value.trim();
  const name = formName.value.trim();
  if (!slug || !name) return false;
  if (upsertMode.value === "create") return true;
  return slug === initialSlug.value && name !== initialName.value;
});

async function saveProject(): Promise<void> {
  await projects.createOrUpdate(formSlug.value, formName.value, paging.limit.value);
  paging.markLoaded(projects.items.length);
  if (!projects.saveError) {
    upsertDialogOpen.value = false;
    snackbar.success(t("common.saved"));
  }
}
</script>
