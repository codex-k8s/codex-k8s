<template>
  <div>
    <PageHeader :title="t('pages.projectDetails.title')">
      <template #actions>
        <CopyChip :label="t('pages.projectDetails.projectId')" :value="projectId" icon="mdi-identifier" />
        <VBtn variant="tonal" prepend-icon="mdi-arrow-left" :to="{ name: 'projects' }">
          {{ t("common.back") }}
        </VBtn>
        <VBtn variant="tonal" prepend-icon="mdi-refresh" :loading="details.loading" @click="load">
          {{ t("common.refresh") }}
        </VBtn>
        <VBtn v-if="auth.isPlatformOwner" color="error" variant="tonal" prepend-icon="mdi-delete-outline" @click="confirmOpen = true">
          {{ t("common.delete") }}
        </VBtn>
      </template>
    </PageHeader>

    <VAlert v-if="details.error" type="error" variant="tonal" class="mt-4">
      {{ t(details.error.messageKey) }}
    </VAlert>
    <VAlert v-if="projects.deleteError" type="error" variant="tonal" class="mt-4">
      {{ t(projects.deleteError.messageKey) }}
    </VAlert>

    <VRow class="mt-4" density="compact">
      <VCol cols="12" md="6">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-1">{{ t("pages.projectDetails.slug") }}</VCardTitle>
          <VCardText>
            <span class="mono">{{ details.item?.slug || "-" }}</span>
          </VCardText>
        </VCard>
      </VCol>
      <VCol cols="12" md="6">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-1">{{ t("pages.projectDetails.name") }}</VCardTitle>
          <VCardText>
            <span>{{ details.item?.name || "-" }}</span>
          </VCardText>
        </VCard>
      </VCol>
    </VRow>

    <VCard class="mt-4" variant="outlined">
      <VCardText class="d-flex ga-2 flex-wrap">
        <VBtn variant="tonal" prepend-icon="mdi-source-repository" :to="{ name: 'project-repositories', params: { projectId } }">
          {{ t("pages.projects.repos") }}
        </VBtn>
        <VBtn variant="tonal" prepend-icon="mdi-account-group-outline" :to="{ name: 'project-members', params: { projectId } }">
          {{ t("pages.projects.members") }}
        </VBtn>
      </VCardText>
    </VCard>
  </div>

  <ConfirmDialog
    v-model="confirmOpen"
    :title="t('common.delete')"
    :message="details.item ? details.item.name : projectId"
    :confirm-text="t('common.delete')"
    :cancel-text="t('common.cancel')"
    danger
    @confirm="doDelete"
  />
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

import ConfirmDialog from "../shared/ui/ConfirmDialog.vue";
import CopyChip from "../shared/ui/CopyChip.vue";
import PageHeader from "../shared/ui/PageHeader.vue";
import { useSnackbarStore } from "../shared/ui/feedback/snackbar-store";
import { useAuthStore } from "../features/auth/store";
import { useProjectsStore } from "../features/projects/projects-store";
import { useProjectDetailsStore } from "../features/projects/details-store";

const props = defineProps<{ projectId: string }>();

const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();
const projects = useProjectsStore();
const details = useProjectDetailsStore();
const router = useRouter();
const snackbar = useSnackbarStore();

const confirmOpen = ref(false);

async function load() {
  await details.load(props.projectId);
}

async function doDelete() {
  await projects.remove(props.projectId);
  if (!projects.deleteError) {
    snackbar.success(t("common.deleted"));
    await router.push({ name: "projects" });
  }
}

onMounted(() => void load());
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>

