<template>
  <section class="card">
    <div class="row">
      <h2>{{ t("pages.projects.title") }}</h2>
      <button class="btn" type="button" @click="load" :disabled="projects.loading">
        {{ t("common.refresh") }}
      </button>
    </div>

    <div v-if="projects.error" class="err">{{ t(projects.error.messageKey) }}</div>

    <table v-if="projects.items.length" class="tbl">
      <thead>
        <tr>
          <th>{{ t("pages.projects.slug") }}</th>
          <th>{{ t("pages.projects.name") }}</th>
          <th class="center">{{ t("pages.projects.role") }}</th>
          <th v-if="auth.isPlatformAdmin" class="center">{{ t("pages.projects.manage") }}</th>
          <th class="center">{{ t("pages.projects.id") }}</th>
          <th v-if="auth.isPlatformOwner"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in projects.items" :key="p.id">
          <td>{{ p.slug }}</td>
          <td>
            <RouterLink class="lnk" :to="{ name: 'project-details', params: { projectId: p.id } }">
              {{ p.name }}
            </RouterLink>
          </td>
          <td class="center">{{ roleLabel(p.role) }}</td>
          <td v-if="auth.isPlatformAdmin" class="manage">
            <RouterLink class="lnk" :to="{ name: 'project-repositories', params: { projectId: p.id } }">
              {{ t("pages.projects.repos") }}
            </RouterLink>
            <RouterLink class="lnk" :to="{ name: 'project-members', params: { projectId: p.id } }">
              {{ t("pages.projects.members") }}
            </RouterLink>
          </td>
          <td class="mono center">{{ p.id }}</td>
          <td v-if="auth.isPlatformOwner" class="right">
            <button class="btn danger" type="button" @click="askDelete(p.id, p.name)" :disabled="projects.deleting">
              {{ t("common.delete") }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-else class="muted">{{ t("states.noProjects") }}</div>

    <template v-if="auth.isPlatformAdmin">
      <div class="sep"></div>

      <div class="row">
        <h3>{{ t("pages.projects.createTitle") }}</h3>
      </div>

      <div class="form">
        <label>
          <div class="lbl">{{ t("pages.projects.slug") }}</div>
          <input class="inp mono" v-model="slug" :placeholder="t('placeholders.projectSlug')" />
        </label>
        <label>
          <div class="lbl">{{ t("pages.projects.name") }}</div>
          <input class="inp" v-model="name" :placeholder="t('placeholders.projectName')" />
        </label>
        <button class="btn primary" type="button" @click="createOrUpdate" :disabled="projects.saving">
          {{ t("common.createOrUpdate") }}
        </button>
      </div>

      <div v-if="projects.saveError" class="err">{{ t(projects.saveError.messageKey) }}</div>
      <div v-if="projects.deleteError" class="err">{{ t(projects.deleteError.messageKey) }}</div>
    </template>

    <template v-else>
      <div class="sep"></div>
      <div class="muted">{{ t("pages.projects.adminOnlyHint") }}</div>
    </template>
  </section>

  <ConfirmModal
    :open="confirmOpen"
    :title="t('common.delete')"
    :message="confirmName"
    :confirmText="t('common.delete')"
    :cancelText="t('common.cancel')"
    danger
    @cancel="confirmOpen = false"
    @confirm="doDelete"
  />
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import ConfirmModal from "../shared/ui/ConfirmModal.vue";
import { useAuthStore } from "../features/auth/store";
import { useProjectsStore } from "../features/projects/projects-store";

const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();
const projects = useProjectsStore();

const slug = ref("");
const name = ref("");

const confirmOpen = ref(false);
const confirmProjectId = ref<string>("");
const confirmName = ref<string>("");

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
}

onMounted(() => void load());
</script>

<style scoped>
h2 {
  margin: 0;
  letter-spacing: -0.01em;
}
h3 {
  margin: 0;
  letter-spacing: -0.01em;
  font-size: 14px;
  opacity: 0.9;
}
.manage {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
  justify-content: center;
}
.form {
  display: grid;
  grid-template-columns: minmax(180px, 1fr) minmax(180px, 1fr) auto;
  gap: 14px;
  margin-top: 12px;
  align-items: end;
}
@media (max-width: 840px) {
  .form {
    grid-template-columns: 1fr;
  }
}
</style>
