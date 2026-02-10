<template>
  <div class="page">
    <header class="top">
      <div class="brand">
        <div class="title">{{ t("app.title") }}</div>
        <div class="subtitle">{{ t("app.subtitle") }}</div>
      </div>

      <div class="actions">
        <div class="lang" role="group" aria-label="Language selector">
          <button
            class="btn lang-btn"
            :class="{ active: locale === 'en' }"
            @click="setLocale('en')"
            type="button"
            :title="'en'"
          >
            {{ t("i18n.enFlag") }}
          </button>
          <button
            class="btn lang-btn"
            :class="{ active: locale === 'ru' }"
            @click="setLocale('ru')"
            type="button"
            :title="'ru'"
          >
            {{ t("i18n.ruFlag") }}
          </button>
        </div>

        <template v-if="auth.status === 'authed'">
          <div class="who">
            <div class="email">{{ auth.me?.email }}</div>
            <div class="meta">
              <span v-if="auth.me?.githubLogin" class="mono">@{{ auth.me.githubLogin }}</span>
              <span v-if="auth.me?.isPlatformAdmin" class="pill">{{ t("roles.admin") }}</span>
            </div>
          </div>
          <button class="btn" @click="logout" type="button">{{ t("common.logout") }}</button>
        </template>

        <template v-else-if="auth.status === 'anon'">
          <a class="btn primary" href="/oauth2/start">{{ t("common.loginWithGitHub") }}</a>
        </template>
      </div>
    </header>

    <nav v-if="auth.status === 'authed'" class="nav">
      <div class="tabs">
        <RouterLink :to="{ name: 'projects' }">{{ t("nav.projects") }}</RouterLink>
        <span v-if="crumbAfterProjectsKey" class="crumb"> &gt; {{ t(crumbAfterProjectsKey) }}</span>

        <RouterLink :to="{ name: 'runs' }">{{ t("nav.runs") }}</RouterLink>
        <span v-if="crumbAfterRunsKey" class="crumb"> &gt; {{ t(crumbAfterRunsKey) }}</span>

        <RouterLink v-if="auth.isPlatformAdmin" :to="{ name: 'users' }">{{ t("nav.users") }}</RouterLink>
      </div>
    </nav>

    <main class="main">
      <div v-if="auth.status === 'loading'" class="card">{{ t("common.loading") }}</div>
      <div v-else-if="auth.status === 'anon'" class="card">
        <div class="h">{{ t("states.accessRequiredTitle") }}</div>
        <div class="p">{{ t("states.accessRequiredText") }}</div>
      </div>
      <RouterView v-else />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";
import { useI18n } from "vue-i18n";

import { persistLocale, type Locale } from "../i18n/locale";
import { useAuthStore } from "../features/auth/store";

const auth = useAuthStore();
const { t, locale } = useI18n({ useScope: "global" });
const route = useRoute();

const crumbAfterProjectsKey = computed(() => {
  const meta = route.meta as Record<string, any>;
  if (meta?.section === "projects" && typeof meta?.crumbKey === "string" && meta.crumbKey) return meta.crumbKey as string;
  return "";
});

const crumbAfterRunsKey = computed(() => {
  const meta = route.meta as Record<string, any>;
  if (meta?.section === "runs" && typeof meta?.crumbKey === "string" && meta.crumbKey) return meta.crumbKey as string;
  return "";
});

function setLocale(next: Locale) {
  locale.value = next;
  persistLocale(next);
}

async function logout() {
  await auth.logout();
}

onMounted(() => {
  void auth.ensureLoaded();
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
  --link-color: #0b5a83;
  --link-color-hover: #084a6b;
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
.lang {
  display: inline-flex;
  gap: 6px;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
}
.lang-btn {
  padding: 7px 10px;
  border-radius: 999px;
  font-weight: 900;
}
.lang-btn.active {
  background: rgba(17, 24, 39, 0.08);
  border-color: rgba(17, 24, 39, 0.18);
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
.tabs {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.nav a {
  padding: 8px 10px;
  border-radius: 10px;
  text-decoration: none;
  color: var(--link-color);
  font-weight: 700;
  opacity: 0.8;
}
.crumb {
  opacity: 0.55;
  font-weight: 800;
  padding: 0 2px;
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
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}
</style>
