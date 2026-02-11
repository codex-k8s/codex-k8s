<template>
  <section class="card">
    <div class="row">
      <div>
        <h2>{{ t("pages.projectRepositories.title") }}</h2>
        <div v-if="details.item" class="muted">
          <RouterLink class="lnk" :to="{ name: 'project-details', params: { projectId } }">{{ details.item.name }}</RouterLink>
        </div>
        <div v-else class="muted mono">{{ t("pages.projectRepositories.projectId") }}: {{ projectId }}</div>
      </div>
      <div class="actions">
        <RouterLink class="btn equal" :to="{ name: 'projects' }">{{ t("common.back") }}</RouterLink>
        <button class="btn equal" type="button" @click="load" :disabled="repos.loading">{{ t("common.refresh") }}</button>
      </div>
    </div>

    <div v-if="repos.error" class="err">{{ t(repos.error.messageKey) }}</div>

    <table v-if="repos.items.length" class="tbl">
      <thead>
        <tr>
          <th class="center">{{ t("pages.projectRepositories.provider") }}</th>
          <th class="center">{{ t("pages.projectRepositories.repo") }}</th>
          <th class="center">{{ t("pages.projectRepositories.servicesYaml") }}</th>
          <th class="center">{{ t("pages.projectRepositories.externalId") }}</th>
          <th class="center">{{ t("pages.projectRepositories.id") }}</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in repos.items" :key="r.id">
          <td class="center">{{ r.provider }}</td>
          <td class="mono center">{{ r.owner }}/{{ r.name }}</td>
          <td class="mono center">{{ r.services_yaml_path }}</td>
          <td class="mono center">{{ r.external_id }}</td>
          <td class="mono center">{{ r.id }}</td>
          <td class="right">
            <button class="btn danger" type="button" @click="askRemove(r.id, r.owner + '/' + r.name)" :disabled="repos.removing">
              {{ t("common.delete") }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-else class="muted">{{ t("states.noRepos") }}</div>

    <div class="sep"></div>

    <div class="row">
      <h3>{{ t("pages.projectRepositories.attachTitle") }}</h3>
    </div>

    <div class="form">
      <label>
        <div class="lbl">{{ t("pages.projectRepositories.owner") }}</div>
        <input class="inp mono" v-model="owner" :placeholder="t('placeholders.repoOwner')" />
      </label>
      <label>
        <div class="lbl">{{ t("pages.projectRepositories.name") }}</div>
        <input class="inp mono" v-model="name" :placeholder="t('placeholders.repoName')" />
      </label>
      <label>
        <div class="lbl">{{ t("pages.projectRepositories.servicesYamlPath") }}</div>
        <input class="inp mono" v-model="servicesYamlPath" :placeholder="t('placeholders.servicesYamlPath')" />
      </label>
      <label class="wide">
        <div class="lbl">{{ t("pages.projectRepositories.repoToken") }}</div>
        <input class="inp mono" type="password" v-model="token" :placeholder="t('placeholders.repoToken')" />
      </label>
      <button class="btn primary" type="button" @click="attach" :disabled="repos.attaching">
        {{ t("common.attachEnsureWebhook") }}
      </button>
    </div>

    <div v-if="repos.attachError" class="err">{{ t(repos.attachError.messageKey) }}</div>
  </section>

  <ConfirmModal
    :open="confirmOpen"
    :title="t('common.delete')"
    :message="confirmName"
    :confirmText="t('common.delete')"
    :cancelText="t('common.cancel')"
    danger
    @cancel="confirmOpen = false"
    @confirm="doRemove"
  />
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import ConfirmModal from "../shared/ui/ConfirmModal.vue";
import { useProjectRepositoriesStore } from "../features/projects/repositories-store";
import { useProjectDetailsStore } from "../features/projects/details-store";

const props = defineProps<{ projectId: string }>();

const { t } = useI18n({ useScope: "global" });
const repos = useProjectRepositoriesStore();
const details = useProjectDetailsStore();

const owner = ref("");
const name = ref("");
const servicesYamlPath = ref("services.yaml");
const token = ref("");

const confirmOpen = ref(false);
const confirmRepoId = ref("");
const confirmName = ref("");

async function load() {
  await details.load(props.projectId);
  await repos.load(props.projectId);
}

async function attach() {
  await repos.attach({ owner: owner.value, name: name.value, token: token.value, servicesYamlPath: servicesYamlPath.value });
  if (!repos.attachError) {
    owner.value = "";
    name.value = "";
    token.value = "";
    servicesYamlPath.value = "services.yaml";
  }
}

function askRemove(repositoryId: string, label: string) {
  confirmRepoId.value = repositoryId;
  confirmName.value = label;
  confirmOpen.value = true;
}

async function doRemove() {
  const id = confirmRepoId.value;
  confirmOpen.value = false;
  confirmRepoId.value = "";
  if (!id) return;
  await repos.remove(id);
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
.form {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 12px;
  margin-top: 12px;
  align-items: end;
}
.wide {
  grid-column: 1 / -1;
}
@media (max-width: 960px) {
  .form {
    grid-template-columns: 1fr;
  }
}
</style>
