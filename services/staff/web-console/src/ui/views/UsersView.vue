<template>
  <section class="grid">
    <div class="card">
      <div class="row">
        <h2>Users</h2>
        <button class="btn" @click="load" :disabled="loading">Refresh</button>
      </div>
      <div v-if="error" class="err">{{ error }}</div>
      <table v-if="items.length" class="tbl">
        <thead>
          <tr>
            <th>Email</th>
            <th>GitHub</th>
            <th>Admin</th>
            <th>ID</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in items" :key="u.id">
            <td>{{ u.email }}</td>
            <td class="mono">{{ u.github_login || "-" }}</td>
            <td>{{ u.is_platform_admin ? "yes" : "no" }}</td>
            <td class="mono">{{ u.id }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="muted">No users.</div>
    </div>

    <div class="card">
      <h2>Add Allowed User</h2>
      <div class="muted">
        Registration is disabled. Adding an email here allows first GitHub OAuth login for that email.
      </div>
      <div class="form">
        <label>
          <div class="lbl">Email</div>
          <input v-model="newEmail" class="inp" placeholder="user@example.com" />
        </label>
        <label class="chk">
          <input type="checkbox" v-model="newAdmin" />
          <span>Platform admin</span>
        </label>
        <button class="btn primary" @click="createUser" :disabled="creating">Create / Update</button>
        <div v-if="createError" class="err">{{ createError }}</div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { api } from "../api";

type User = { id: string; email: string; github_login: string; is_platform_admin: boolean };

const items = ref<User[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

const newEmail = ref("");
const newAdmin = ref(false);
const creating = ref(false);
const createError = ref<string | null>(null);

async function load() {
  loading.value = true;
  error.value = null;
  try {
    const resp = await api.get("/api/v1/staff/users");
    items.value = (resp.data.items ?? []) as User[];
  } catch (e: any) {
    error.value = e?.message ?? "Failed to load";
  } finally {
    loading.value = false;
  }
}

async function createUser() {
  creating.value = true;
  createError.value = null;
  try {
    await api.post("/api/v1/staff/users", {
      email: newEmail.value,
      is_platform_admin: newAdmin.value,
    });
    newEmail.value = "";
    newAdmin.value = false;
    await load();
  } catch (e: any) {
    createError.value = e?.message ?? "Failed to create";
  } finally {
    creating.value = false;
  }
}

onMounted(() => void load());
</script>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: 1.2fr 0.8fr;
  gap: 14px;
}
@media (max-width: 960px) {
  .grid {
    grid-template-columns: 1fr;
  }
}
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
.btn.primary {
  background: #111827;
  color: #fff;
  border-color: rgba(17, 24, 39, 0.4);
}
.muted {
  margin-top: 10px;
  opacity: 0.7;
  font-size: 13px;
}
.err {
  margin-top: 10px;
  color: #b42318;
  font-weight: 700;
}
.form {
  display: grid;
  gap: 10px;
  margin-top: 12px;
}
.lbl {
  font-size: 12px;
  opacity: 0.8;
  margin-bottom: 4px;
}
.inp {
  width: 100%;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid rgba(17, 24, 39, 0.18);
  background: rgba(255, 255, 255, 0.9);
}
.chk {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  opacity: 0.9;
}
</style>

