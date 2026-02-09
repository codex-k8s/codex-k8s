<template>
  <section class="card">
    <div class="row">
      <h2>Projects</h2>
      <button class="btn" @click="load" :disabled="loading">Refresh</button>
    </div>
    <div v-if="error" class="err">{{ error }}</div>
    <table v-if="items.length" class="tbl">
      <thead>
        <tr>
          <th>Slug</th>
          <th>Name</th>
          <th>Role</th>
          <th>Manage</th>
          <th>ID</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in items" :key="p.id">
          <td>{{ p.slug }}</td>
          <td>{{ p.name }}</td>
          <td>{{ p.role }}</td>
          <td class="manage">
            <RouterLink class="lnk" :to="`/projects/${p.id}/repositories`">Repos</RouterLink>
            <RouterLink class="lnk" :to="`/projects/${p.id}/members`">Members</RouterLink>
          </td>
          <td class="mono">{{ p.id }}</td>
        </tr>
      </tbody>
    </table>
    <div v-else class="muted">No projects yet. Trigger a webhook run first.</div>

    <div class="sep"></div>

    <div class="row">
      <h3>Create / Update Project</h3>
    </div>
    <div class="form">
      <label>
        <div class="lbl">Slug</div>
        <input class="inp mono" v-model="newSlug" placeholder="codex-k8s" />
      </label>
      <label>
        <div class="lbl">Name</div>
        <input class="inp" v-model="newName" placeholder="codex-k8s" />
      </label>
      <button class="btn primary" @click="createOrUpdate" :disabled="creating">Create / Update</button>
    </div>
    <div v-if="createError" class="err">{{ createError }}</div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { api } from "../api";

type Project = { id: string; slug: string; name: string; role: string };

const items = ref<Project[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const newSlug = ref("");
const newName = ref("");
const creating = ref(false);
const createError = ref<string | null>(null);

async function load() {
  loading.value = true;
  error.value = null;
  try {
    const resp = await api.get("/api/v1/staff/projects");
    items.value = (resp.data.items ?? []) as Project[];
  } catch (e: any) {
    error.value = e?.message ?? "Failed to load";
  } finally {
    loading.value = false;
  }
}

async function createOrUpdate() {
  creating.value = true;
  createError.value = null;
  try {
    await api.post("/api/v1/staff/projects", { slug: newSlug.value, name: newName.value });
    newSlug.value = "";
    newName.value = "";
    await load();
  } catch (e: any) {
    createError.value = e?.message ?? "Failed to create project";
  } finally {
    creating.value = false;
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
  align-items: center;
  justify-content: space-between;
  gap: 10px;
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
  opacity: 0.85;
}
.btn {
  border: 1px solid rgba(17, 24, 39, 0.14);
  background: rgba(255, 255, 255, 0.75);
  padding: 8px 10px;
  border-radius: 10px;
  font-weight: 800;
  cursor: pointer;
}
.btn.primary {
  border-color: rgba(17, 24, 39, 0.25);
  background: #111827;
  color: #fff;
}
.muted {
  margin-top: 12px;
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
  grid-template-columns: 1fr 1fr auto;
  gap: 10px;
  margin-top: 10px;
  align-items: end;
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
.lnk {
  font-weight: 800;
  text-decoration: none;
  color: #111827;
  opacity: 0.8;
}
.manage {
  display: flex;
  gap: 10px;
}
.lnk:hover {
  opacity: 1;
  text-decoration: underline;
}
@media (max-width: 840px) {
  .form {
    grid-template-columns: 1fr;
  }
}
</style>
