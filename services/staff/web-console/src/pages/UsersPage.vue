<template>
  <section class="grid">
    <div class="card">
      <div class="row">
        <h2>{{ t("pages.users.title") }}</h2>
        <button class="btn" type="button" @click="load" :disabled="users.loading">
          {{ t("common.refresh") }}
        </button>
      </div>

      <div v-if="users.error" class="err">{{ t(users.error.messageKey) }}</div>

      <table v-if="users.items.length" class="tbl">
        <thead>
          <tr>
            <th>{{ t("pages.users.email") }}</th>
            <th>{{ t("pages.users.github") }}</th>
            <th>{{ t("pages.users.admin") }}</th>
            <th>{{ t("common.id") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users.items" :key="u.id">
            <td>{{ u.email }}</td>
            <td class="mono">{{ u.githubLogin || "-" }}</td>
            <td>{{ u.isPlatformAdmin ? t("pages.users.yes") : t("pages.users.no") }}</td>
            <td class="mono">{{ u.id }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="muted">{{ t("states.noUsers") }}</div>
    </div>

    <div class="card">
      <h2>{{ t("pages.users.addAllowedUser") }}</h2>
      <div class="muted">{{ t("pages.users.addAllowedUserHint") }}</div>

      <div class="form">
        <label class="email">
          <div class="lbl">{{ t("pages.users.email") }}</div>
          <input v-model="email" class="inp" :placeholder="t('placeholders.userEmail')" />
        </label>

        <button class="btn primary" type="button" @click="create" :disabled="users.creating">
          {{ t("common.createOrUpdate") }}
        </button>

        <label class="chk">
          <input type="checkbox" v-model="isAdmin" />
          <span>{{ t("pages.users.platformAdmin") }}</span>
        </label>

        <div v-if="users.createError" class="err">{{ t(users.createError.messageKey) }}</div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

import { useUsersStore } from "../features/users/store";

const { t } = useI18n({ useScope: "global" });
const users = useUsersStore();

const email = ref("");
const isAdmin = ref(false);

async function load() {
  await users.load();
}

async function create() {
  await users.create(email.value, isAdmin.value);
  if (!users.createError) {
    email.value = "";
    isAdmin.value = false;
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
h2 {
  margin: 0;
  letter-spacing: -0.01em;
}
.form {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 12px;
  margin-top: 12px;
  align-items: end;
}
.email {
  min-width: 0;
}
.chk {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 800;
  opacity: 0.9;
}
.err {
  grid-column: 1 / -1;
}
@media (max-width: 960px) {
  .form {
    grid-template-columns: 1fr;
  }
}
</style>
