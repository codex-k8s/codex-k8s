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
            <th class="center">{{ t("pages.users.github") }}</th>
            <th class="center">{{ t("pages.users.admin") }}</th>
            <th class="center">{{ t("common.id") }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users.items" :key="u.id">
            <td>{{ u.email }}</td>
            <td class="mono center">{{ u.github_login || "-" }}</td>
            <td class="center">{{ u.is_platform_admin ? t("pages.users.yes") : t("pages.users.no") }}</td>
            <td class="mono center">{{ u.id }}</td>
            <td class="right">
              <button v-if="canDelete(u.id, u.is_platform_admin, u.is_platform_owner)" class="btn danger" type="button" @click="askRemove(u.id, u.email)" :disabled="users.deleting">
                {{ t("common.delete") }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="muted">{{ t("states.noUsers") }}</div>

      <div v-if="users.deleteError" class="err">{{ t(users.deleteError.messageKey) }}</div>
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

  <ConfirmModal
    :open="confirmOpen"
    :title="t('common.delete')"
    :message="confirmName"
    :confirmText="t('common.delete')"
    :cancelText="t('common.cancel')"
    danger
    @cancel="confirmOpen = false"
    @confirm="doRemove"
  />
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

import ConfirmModal from "../shared/ui/ConfirmModal.vue";
import { useAuthStore } from "../features/auth/store";
import { useUsersStore } from "../features/users/store";

const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();
const users = useUsersStore();

const email = ref("");
const isAdmin = ref(false);

const confirmOpen = ref(false);
const confirmUserId = ref("");
const confirmName = ref("");

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

function canDelete(userId: string, isPlatformAdmin: boolean, isPlatformOwner: boolean): boolean {
  if (!auth.me) return false;
  if (userId === auth.me.id) return false;
  if (auth.isPlatformOwner) return true;
  if (!auth.isPlatformAdmin) return false;
  // Platform admin cannot delete other admins/owner.
  if (isPlatformOwner || isPlatformAdmin) return false;
  return true;
}

function askRemove(userId: string, emailLabel: string) {
  confirmUserId.value = userId;
  confirmName.value = emailLabel;
  confirmOpen.value = true;
}

async function doRemove() {
  const id = confirmUserId.value;
  confirmOpen.value = false;
  confirmUserId.value = "";
  if (!id) return;
  await users.remove(id);
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
