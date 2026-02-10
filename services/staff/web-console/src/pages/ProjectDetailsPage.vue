<template>
  <section class="card">
    <div class="row">
      <div>
        <h2>{{ t("pages.projectDetails.title") }}</h2>
        <div class="muted mono">{{ t("pages.projectDetails.projectId") }}: {{ projectId }}</div>
      </div>
      <div class="actions">
        <RouterLink class="btn equal" :to="{ name: 'projects' }">{{ t("common.back") }}</RouterLink>
        <button class="btn equal" type="button" @click="load" :disabled="details.loading">{{ t("common.refresh") }}</button>
      </div>
    </div>

    <div v-if="details.error" class="err">{{ t(details.error.messageKey) }}</div>

    <div v-if="details.item" class="grid">
      <div class="kv">
        <div class="k">{{ t("pages.projectDetails.slug") }}</div>
        <div class="v mono">{{ details.item.slug }}</div>
      </div>
      <div class="kv">
        <div class="k">{{ t("pages.projectDetails.name") }}</div>
        <div class="v">{{ details.item.name }}</div>
      </div>
    </div>

    <div class="sep"></div>

    <div class="row">
      <div class="actions">
        <RouterLink class="btn" :to="{ name: 'project-repositories', params: { projectId } }">
          {{ t("pages.projects.repos") }}
        </RouterLink>
        <RouterLink class="btn" :to="{ name: 'project-members', params: { projectId } }">
          {{ t("pages.projects.members") }}
        </RouterLink>
      </div>

      <button v-if="auth.isPlatformOwner" class="btn danger" type="button" @click="askDelete">
        {{ t("common.delete") }}
      </button>
    </div>

    <div v-if="projects.deleteError" class="err">{{ t(projects.deleteError.messageKey) }}</div>

    <ConfirmModal
      :open="confirmOpen"
      :title="t('common.delete')"
      :message="details.item ? details.item.name : projectId"
      :confirmText="t('common.delete')"
      :cancelText="t('common.cancel')"
      danger
      @cancel="confirmOpen = false"
      @confirm="doDelete"
    />
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

import ConfirmModal from "../shared/ui/ConfirmModal.vue";
import { useAuthStore } from "../features/auth/store";
import { useProjectsStore } from "../features/projects/projects-store";
import { useProjectDetailsStore } from "../features/projects/details-store";

const props = defineProps<{ projectId: string }>();

const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();
const projects = useProjectsStore();
const details = useProjectDetailsStore();
const router = useRouter();

const confirmOpen = ref(false);

async function load() {
  await details.load(props.projectId);
}

function askDelete() {
  confirmOpen.value = true;
}

async function doDelete() {
  confirmOpen.value = false;
  await projects.remove(props.projectId);
  if (!projects.deleteError) {
    await router.push({ name: "projects" });
  }
}

onMounted(() => void load());
</script>

<style scoped>
h2 {
  margin: 0;
  letter-spacing: -0.01em;
}
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-top: 12px;
}
.kv {
  border: 1px solid rgba(17, 24, 39, 0.1);
  background: rgba(255, 255, 255, 0.6);
  border-radius: 12px;
  padding: 10px;
}
.k {
  font-weight: 900;
  font-size: 12px;
  opacity: 0.75;
}
.v {
  margin-top: 6px;
  font-weight: 800;
}
@media (max-width: 960px) {
  .grid {
    grid-template-columns: 1fr;
  }
}
</style>
