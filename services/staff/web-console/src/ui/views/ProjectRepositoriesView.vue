<template>
  <section class="card">
    <div class="row">
      <div>
        <h2>Project Repositories</h2>
        <div class="muted mono">project_id: {{ projectId }}</div>
      </div>
      <div class="actions">
        <RouterLink class="btn" to="/">Back</RouterLink>
        <button class="btn" @click="load" :disabled="loading">Refresh</button>
      </div>
    </div>

    <div v-if="error" class="err">{{ error }}</div>

    <table v-if="items.length" class="tbl">
      <thead>
        <tr>
          <th>Provider</th>
          <th>Repo</th>
          <th>services.yaml</th>
          <th>External ID</th>
          <th>ID</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in items" :key="r.id">
          <td>{{ r.provider }}</td>
          <td class="mono">{{ r.owner }}/{{ r.name }}</td>
          <td class="mono">{{ r.services_yaml_path }}</td>
          <td class="mono">{{ r.external_id }}</td>
          <td class="mono">{{ r.id }}</td>
          <td class="right">
            <button class="btn danger" @click="remove(r.id)" :disabled="removing">Delete</button>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-else class="muted">No repositories attached yet.</div>

    <div class="sep"></div>

    <div class="row">
      <h3>Attach GitHub Repository</h3>
    </div>

    <div class="form">
      <label>
        <div class="lbl">Owner</div>
        <input class="inp mono" v-model="owner" placeholder="codex-k8s" />
      </label>
      <label>
        <div class="lbl">Name</div>
        <input class="inp mono" v-model="name" placeholder="codex-k8s" />
      </label>
      <label>
        <div class="lbl">services.yaml path</div>
        <input class="inp mono" v-model="servicesYamlPath" placeholder="services.yaml" />
      </label>
      <label class="wide">
        <div class="lbl">Repo token (stored encrypted in DB)</div>
        <input class="inp mono" type="password" v-model="token" placeholder="ghp_... / fine-grained token" />
      </label>
      <button class="btn primary" @click="attach" :disabled="attaching">Attach + Ensure Webhook</button>
    </div>
    <div v-if="attachError" class="err">{{ attachError }}</div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { api } from "../api";

type RepoBinding = {
  id: string;
  project_id: string;
  provider: string;
  external_id: number;
  owner: string;
  name: string;
  services_yaml_path: string;
};

const route = useRoute();
const projectId = String(route.params.project_id || "");

const items = ref<RepoBinding[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

const owner = ref("");
const name = ref("");
const servicesYamlPath = ref("services.yaml");
const token = ref("");
const attaching = ref(false);
const attachError = ref<string | null>(null);
const removing = ref(false);

async function load() {
  loading.value = true;
  error.value = null;
  try {
    const resp = await api.get(`/api/v1/staff/projects/${projectId}/repositories`);
    items.value = (resp.data.items ?? []) as RepoBinding[];
  } catch (e: any) {
    error.value = e?.message ?? "Failed to load";
  } finally {
    loading.value = false;
  }
}

async function attach() {
  attaching.value = true;
  attachError.value = null;
  try {
    await api.post(`/api/v1/staff/projects/${projectId}/repositories`, {
      provider: "github",
      owner: owner.value,
      name: name.value,
      token: token.value,
      services_yaml_path: servicesYamlPath.value,
    });
    owner.value = "";
    name.value = "";
    token.value = "";
    servicesYamlPath.value = "services.yaml";
    await load();
  } catch (e: any) {
    attachError.value = e?.message ?? "Failed to attach repository";
  } finally {
    attaching.value = false;
  }
}

async function remove(repositoryId: string) {
  removing.value = true;
  try {
    await api.delete(`/api/v1/staff/projects/${projectId}/repositories/${repositoryId}`);
    await load();
  } catch (e: any) {
    error.value = e?.message ?? "Failed to delete repository";
  } finally {
    removing.value = false;
  }
}

onMounted(() => void load());
</script>

<style scoped>
.card {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(17, 24, 39, 0.1);
  border-radius: 14px;
  padding: 16px;
}
.row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}
.actions {
  display: flex;
  gap: 10px;
  align-items: center;
}
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
.tbl {
  width: 100%;
  border-collapse: collapse;
  margin-top: 12px;
  font-size: 13px;
}
.tbl th,
.tbl td {
  border-top: 1px solid rgba(17, 24, 39, 0.1);
  padding: 10px 8px;
  vertical-align: top;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  opacity: 0.9;
}
.btn {
  border: 1px solid rgba(17, 24, 39, 0.14);
  background: rgba(255, 255, 255, 0.75);
  padding: 8px 10px;
  border-radius: 10px;
  font-weight: 800;
  cursor: pointer;
  text-decoration: none;
  color: inherit;
}
.btn.primary {
  border-color: rgba(17, 24, 39, 0.25);
  background: #111827;
  color: #fff;
}
.btn.danger {
  border-color: rgba(180, 35, 24, 0.35);
  color: #b42318;
}
.muted {
  margin-top: 6px;
  opacity: 0.7;
}
.err {
  margin-top: 10px;
  color: #b42318;
  font-weight: 700;
}
.sep {
  height: 1px;
  background: rgba(17, 24, 39, 0.08);
  margin: 16px 0;
}
.form {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 10px;
  margin-top: 10px;
  align-items: end;
}
.wide {
  grid-column: 1 / -1;
}
.lbl {
  font-size: 12px;
  opacity: 0.75;
  font-weight: 800;
  margin-bottom: 6px;
}
.inp {
  width: 100%;
  border: 1px solid rgba(17, 24, 39, 0.14);
  border-radius: 10px;
  padding: 9px 10px;
  background: rgba(255, 255, 255, 0.85);
}
.right {
  text-align: right;
}
@media (max-width: 900px) {
  .form {
    grid-template-columns: 1fr;
  }
}
</style>

