<template>
  <section class="card">
    <div class="row">
      <div>
        <h2>{{ t("pages.projectRepositories.title") }}</h2>
        <div class="muted mono">{{ t("pages.projectRepositories.projectId") }}: {{ projectId }}</div>
      </div>
      <div class="actions">
        <RouterLink class="btn" :to="{ name: 'projects' }">{{ t("common.back") }}</RouterLink>
        <button class="btn" type="button" @click="load" :disabled="repos.loading">{{ t("common.refresh") }}</button>
      </div>
    </div>

    <div v-if="repos.error" class="err">{{ t(repos.error.messageKey) }}</div>

    <table v-if="repos.items.length" class="tbl">
      <thead>
        <tr>
          <th>{{ t("pages.projectRepositories.provider") }}</th>
          <th>{{ t("pages.projectRepositories.repo") }}</th>
          <th>{{ t("pages.projectRepositories.servicesYaml") }}</th>
          <th>{{ t("pages.projectRepositories.externalId") }}</th>
          <th>{{ t("pages.projectRepositories.id") }}</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in repos.items" :key="r.id">
          <td>{{ r.provider }}</td>
          <td class="mono">{{ r.owner }}/{{ r.name }}</td>
          <td class="mono">{{ r.servicesYamlPath }}</td>
          <td class="mono">{{ r.externalId }}</td>
          <td class="mono">{{ r.id }}</td>
          <td class="right">
            <button class="btn danger" type="button" @click="remove(r.id)" :disabled="repos.removing">
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
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import { useProjectRepositoriesStore } from "../features/projects/repositories-store";

const props = defineProps<{ projectId: string }>();

const { t } = useI18n({ useScope: "global" });
const repos = useProjectRepositoriesStore();

const owner = ref("");
const name = ref("");
const servicesYamlPath = ref("services.yaml");
const token = ref("");

async function load() {
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

async function remove(repositoryId: string) {
  await repos.remove(repositoryId);
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
