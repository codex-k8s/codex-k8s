<template>
  <div>
    <PageHeader :title="t('pages.projects.title')">
      <template #actions>
        <VBtn variant="tonal" icon="mdi-refresh" :title="t('common.refresh')" :loading="projects.loading" @click="load" />
      </template>
    </PageHeader>

    <VAlert v-if="projects.error" type="error" variant="tonal" class="mt-4">
      {{ t(projects.error.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable :headers="headers" :items="projects.items" :loading="projects.loading" :items-per-page="10" hover>
          <template #item.name="{ item }">
            <div class="d-flex justify-center">
              <RouterLink class="text-primary font-weight-bold text-decoration-none" :to="{ name: 'project-details', params: { projectId: item.id } }">
                {{ item.name }}
              </RouterLink>
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

    <VCard v-if="auth.isPlatformAdmin" class="mt-6" variant="outlined">
      <VCardTitle class="text-subtitle-1">{{ t("pages.projects.createTitle") }}</VCardTitle>
      <VCardText>
        <VRow density="compact" class="align-end">
          <VCol cols="12" md="4">
            <VTextField v-model.trim="slug" :label="t('pages.projects.slug')" :placeholder="t('placeholders.projectSlug')" />
          </VCol>
          <VCol cols="12" md="6">
            <VTextField v-model.trim="name" :label="t('pages.projects.name')" :placeholder="t('placeholders.projectName')" />
          </VCol>
          <VCol cols="12" md="2">
            <VBtn class="w-100" color="primary" variant="tonal" :loading="projects.saving" @click="createOrUpdate">
              {{ t("common.createOrUpdate") }}
            </VBtn>
          </VCol>
        </VRow>

        <VAlert v-if="projects.saveError" type="error" variant="tonal" class="mt-4">
          {{ t(projects.saveError.messageKey) }}
        </VAlert>
        <VAlert v-if="projects.deleteError" type="error" variant="tonal" class="mt-4">
          {{ t(projects.deleteError.messageKey) }}
        </VAlert>
      </VCardText>
    </VCard>

    <VAlert v-else class="mt-6" type="info" variant="tonal">
      {{ t("pages.projects.adminOnlyHint") }}
    </VAlert>
  </div>

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
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import PageHeader from "../shared/ui/PageHeader.vue";
import ConfirmDialog from "../shared/ui/ConfirmDialog.vue";
import { useSnackbarStore } from "../shared/ui/feedback/snackbar-store";
import { useAuthStore } from "../features/auth/store";
import { useProjectsStore } from "../features/projects/projects-store";
import { colorForProjectRole } from "../shared/lib/chips";

const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();
const projects = useProjectsStore();
const snackbar = useSnackbarStore();

const slug = ref("");
const name = ref("");

const confirmOpen = ref(false);
const confirmProjectId = ref<string>("");
const confirmName = ref<string>("");

const headers = [
  { title: t("pages.projects.slug"), key: "slug", width: 220, align: "start" },
  { title: t("pages.projects.name"), key: "name", align: "center" },
  { title: t("pages.projects.role"), key: "role", width: 160, sortable: false, align: "center" },
  { title: t("pages.projects.manage"), key: "manage", width: 140, sortable: false, align: "center" },
  { title: "", key: "actions", sortable: false, width: 72, align: "end" },
] as const;

function roleLabel(role: string): string {
  const normalized = role.trim();
  if (normalized === "read") return t("roles.read");
  if (normalized === "read_write") return t("roles.readWrite");
  if (normalized === "admin") return t("roles.admin");
  return normalized;
}

async function load() {
  await projects.load();
}

async function createOrUpdate() {
  await projects.createOrUpdate(slug.value, name.value);
  if (!projects.saveError) {
    slug.value = "";
    name.value = "";
    snackbar.success(t("common.saved"));
  }
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
  await projects.remove(id);
  if (!projects.deleteError) {
    snackbar.success(t("common.deleted"));
  }
}

onMounted(() => void load());
</script>
