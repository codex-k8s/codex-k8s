<template>
  <div>
    <PageHeader :title="t('pages.projects.title')">
      <template #actions>
        <VBtn variant="tonal" prepend-icon="mdi-refresh" :loading="projects.loading" @click="load">
          {{ t("common.refresh") }}
        </VBtn>
      </template>
    </PageHeader>

    <VAlert v-if="projects.error" type="error" variant="tonal" class="mt-4">
      {{ t(projects.error.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable :headers="headers" :items="projects.items" :loading="projects.loading" :items-per-page="10" hover>
          <template #item.name="{ item }">
            <RouterLink class="text-primary font-weight-bold text-decoration-none" :to="{ name: 'project-details', params: { projectId: item.id } }">
              {{ item.name }}
            </RouterLink>
          </template>

          <template #item.role="{ item }">
            <VChip size="small" variant="tonal" class="font-weight-bold" :color="colorForProjectRole(item.role)">
              {{ roleLabel(item.role) }}
            </VChip>
          </template>

          <template #item.manage="{ item }">
            <div class="d-flex ga-2 justify-center flex-wrap">
              <VBtn size="small" variant="text" :to="{ name: 'project-repositories', params: { projectId: item.id } }">
                {{ t("pages.projects.repos") }}
              </VBtn>
              <VBtn size="small" variant="text" :to="{ name: 'project-members', params: { projectId: item.id } }">
                {{ t("pages.projects.members") }}
              </VBtn>
            </div>
          </template>

          <template #item.actions="{ item }">
            <div class="d-flex justify-end">
              <VBtn
                v-if="auth.isPlatformOwner"
                size="small"
                color="error"
                variant="tonal"
                :loading="projects.deleting"
                @click="askDelete(item.id, item.name)"
              >
                {{ t("common.delete") }}
              </VBtn>
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
  { title: t("pages.projects.slug"), key: "slug", width: 220 },
  { title: t("pages.projects.name"), key: "name" },
  { title: t("pages.projects.role"), key: "role", width: 160, sortable: false },
  { title: t("pages.projects.manage"), key: "manage", width: 220, sortable: false },
  { title: t("pages.projects.id"), key: "id", width: 240 },
  { title: "", key: "actions", sortable: false, width: 140 },
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
