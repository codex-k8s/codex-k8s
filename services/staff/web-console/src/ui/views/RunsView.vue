<template>
  <section class="card">
    <div class="row">
      <h2>Runs</h2>
      <button class="btn" @click="load" :disabled="loading">Refresh</button>
    </div>
    <div v-if="error" class="err">{{ error }}</div>
    <table v-if="items.length" class="tbl">
      <thead>
        <tr>
          <th>Status</th>
          <th>Project</th>
          <th>Correlation</th>
          <th>Created</th>
          <th>Started</th>
          <th>Finished</th>
          <th></th>
          <th>ID</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in items" :key="r.id">
          <td><span class="pill" :class="'s-' + r.status">{{ r.status }}</span></td>
          <td class="mono">{{ r.project_id || "-" }}</td>
          <td class="mono">{{ r.correlation_id }}</td>
          <td class="mono">{{ r.created_at }}</td>
          <td class="mono">{{ r.started_at || "-" }}</td>
          <td class="mono">{{ r.finished_at || "-" }}</td>
          <td><RouterLink class="lnk" :to="`/runs/${r.id}`">Details</RouterLink></td>
          <td class="mono">{{ r.id }}</td>
        </tr>
      </tbody>
    </table>
    <div v-else class="muted">No runs yet.</div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { api } from "../api";

type Run = {
  id: string;
  correlation_id: string;
  project_id: string;
  status: string;
  created_at: string;
  started_at: string;
  finished_at: string;
};

const items = ref<Run[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

async function load() {
  loading.value = true;
  error.value = null;
  try {
    const resp = await api.get("/api/v1/staff/runs");
    items.value = (resp.data.items ?? []) as Run[];
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
.pill {
  display: inline-block;
  font-size: 12px;
  font-weight: 800;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px solid rgba(17, 24, 39, 0.14);
  background: rgba(255, 255, 255, 0.7);
}
.s-succeeded {
  background: rgba(5, 150, 105, 0.12);
  border-color: rgba(5, 150, 105, 0.3);
}
.s-failed {
  background: rgba(180, 35, 24, 0.12);
  border-color: rgba(180, 35, 24, 0.3);
}
.s-running {
  background: rgba(37, 99, 235, 0.12);
  border-color: rgba(37, 99, 235, 0.3);
}
.lnk {
  font-weight: 800;
  text-decoration: none;
  color: #111827;
  opacity: 0.8;
}
.lnk:hover {
  opacity: 1;
  text-decoration: underline;
}
</style>
