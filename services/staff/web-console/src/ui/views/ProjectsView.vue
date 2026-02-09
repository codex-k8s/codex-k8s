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
          <th>ID</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in items" :key="p.id">
          <td>{{ p.slug }}</td>
          <td>{{ p.name }}</td>
          <td>{{ p.role }}</td>
          <td class="mono">{{ p.id }}</td>
        </tr>
      </tbody>
    </table>
    <div v-else class="muted">No projects yet. Trigger a webhook run first.</div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { api } from "../api";

type Project = { id: string; slug: string; name: string; role: string };

const items = ref<Project[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

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
.muted {
  margin-top: 12px;
  opacity: 0.7;
}
.err {
  margin-top: 10px;
  color: #b42318;
  font-weight: 700;
}
</style>

