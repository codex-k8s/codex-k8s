<template>
  <div class="page">
    <header class="top">
      <div class="brand">
        <div class="title">codex-k8s</div>
        <div class="subtitle">staff console</div>
      </div>

      <div class="actions">
        <template v-if="authState === 'authed'">
          <div class="who">
            <div class="email">{{ me?.user.email }}</div>
            <div class="meta">
              <span v-if="me?.user.github_login">@{{ me.user.github_login }}</span>
              <span v-if="me?.user.is_platform_admin" class="pill">admin</span>
            </div>
          </div>
          <button class="btn" @click="logout">Logout</button>
        </template>

        <template v-else-if="authState === 'anon'">
          <a class="btn primary" href="/api/v1/auth/github/login">Login with GitHub</a>
        </template>
      </div>
    </header>

    <nav v-if="authState === 'authed'" class="nav">
      <RouterLink to="/">Projects</RouterLink>
      <RouterLink to="/runs">Runs</RouterLink>
      <RouterLink to="/users">Users</RouterLink>
    </nav>

    <main class="main">
      <div v-if="authState === 'loading'" class="card">Loading...</div>
      <div v-else-if="authState === 'anon'" class="card">
        <div class="h">Access required</div>
        <div class="p">
          This console is protected by GitHub OAuth. Ask your admin to allow your email before first login.
        </div>
      </div>
      <RouterView v-else />
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink, RouterView } from "vue-router";
import { api, getMe, type MeResponse } from "./api";

type AuthState = "loading" | "authed" | "anon";

const authState = ref<AuthState>("loading");
const me = ref<MeResponse | null>(null);

async function refresh() {
  authState.value = "loading";
  try {
    me.value = await getMe();
    authState.value = "authed";
  } catch {
    me.value = null;
    authState.value = "anon";
  }
}

async function logout() {
  await api.post("/api/v1/auth/logout");
  await refresh();
}

onMounted(() => {
  void refresh();
});
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: radial-gradient(1200px 500px at 20% 0%, #fff6e5, transparent 55%),
    radial-gradient(900px 450px at 90% 10%, #e8f6ff, transparent 55%),
    linear-gradient(180deg, #fbfbfc, #f4f5f7);
  color: #111827;
  font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif;
}
.top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 20px;
  border-bottom: 1px solid rgba(17, 24, 39, 0.08);
  backdrop-filter: blur(10px);
}
.brand .title {
  font-weight: 800;
  letter-spacing: -0.02em;
}
.brand .subtitle {
  font-size: 12px;
  opacity: 0.7;
}
.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.who {
  text-align: right;
  line-height: 1.1;
}
.who .email {
  font-size: 13px;
  font-weight: 700;
}
.who .meta {
  font-size: 12px;
  opacity: 0.7;
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  align-items: center;
}
.pill {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px solid rgba(17, 24, 39, 0.14);
  background: rgba(255, 255, 255, 0.7);
}
.btn {
  border: 1px solid rgba(17, 24, 39, 0.14);
  background: rgba(255, 255, 255, 0.75);
  padding: 9px 12px;
  border-radius: 10px;
  font-weight: 700;
  cursor: pointer;
  text-decoration: none;
  color: inherit;
}
.btn.primary {
  border-color: rgba(17, 24, 39, 0.25);
  background: #111827;
  color: #fff;
}
.nav {
  display: flex;
  gap: 10px;
  padding: 12px 20px;
}
.nav a {
  padding: 8px 10px;
  border-radius: 10px;
  text-decoration: none;
  color: inherit;
  font-weight: 700;
  opacity: 0.8;
}
.nav a.router-link-active {
  opacity: 1;
  background: rgba(17, 24, 39, 0.08);
}
.main {
  padding: 20px;
  max-width: 1100px;
  margin: 0 auto;
}
.card {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(17, 24, 39, 0.1);
  border-radius: 14px;
  padding: 16px;
}
.h {
  font-weight: 800;
  margin-bottom: 6px;
}
.p {
  opacity: 0.8;
}
</style>

