<template>
  <section class="card">
    <div class="row">
      <div>
        <h2>Project Members</h2>
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
          <th>Email</th>
          <th>User ID</th>
          <th>Role</th>
          <th>Learning override</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="m in items" :key="m.user_id">
          <td>{{ m.email }}</td>
          <td class="mono">{{ m.user_id }}</td>
          <td>
            <select class="sel mono" v-model="m.role">
              <option value="read">read</option>
              <option value="read_write">read_write</option>
              <option value="admin">admin</option>
            </select>
          </td>
          <td>
            <select class="sel mono" v-model="m.learning_mode_override">
              <option :value="''">inherit</option>
              <option :value="'true'">true</option>
              <option :value="'false'">false</option>
            </select>
          </td>
          <td class="right">
            <button class="btn primary" @click="save(m)" :disabled="saving">Save</button>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-else class="muted">No members.</div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { api } from "../api";

type MemberRow = {
  project_id: string;
  user_id: string;
  email: string;
  role: "read" | "read_write" | "admin";
  // UI-only tri-state string: '' (inherit), 'true', 'false'
  learning_mode_override: "" | "true" | "false";
};

const route = useRoute();
const projectId = String(route.params.project_id || "");

const items = ref<MemberRow[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const saving = ref(false);

async function load() {
  loading.value = true;
  error.value = null;
  try {
    const resp = await api.get(`/api/v1/staff/projects/${projectId}/members`);
    const raw = (resp.data.items ?? []) as any[];
    items.value = raw.map((m) => ({
      project_id: String(m.project_id),
      user_id: String(m.user_id),
      email: String(m.email),
      role: m.role as any,
      learning_mode_override: "",
    }));
  } catch (e: any) {
    error.value = e?.message ?? "Failed to load";
  } finally {
    loading.value = false;
  }
}

async function save(m: MemberRow) {
  saving.value = true;
  error.value = null;
  try {
    await api.post(`/api/v1/staff/projects/${projectId}/members`, {
      user_id: m.user_id,
      role: m.role,
    });

    let enabled: boolean | null = null;
    if (m.learning_mode_override === "true") enabled = true;
    if (m.learning_mode_override === "false") enabled = false;
    await api.put(`/api/v1/staff/projects/${projectId}/members/${m.user_id}/learning-mode`, {
      enabled,
    });
  } catch (e: any) {
    error.value = e?.message ?? "Failed to save";
  } finally {
    saving.value = false;
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
.sel {
  border: 1px solid rgba(17, 24, 39, 0.14);
  border-radius: 10px;
  padding: 8px 10px;
  background: rgba(255, 255, 255, 0.85);
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
.right {
  text-align: right;
}
</style>

