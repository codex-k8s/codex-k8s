<template>
  <section class="card">
    <div class="row">
      <div>
        <h2>{{ t("pages.projectMembers.title") }}</h2>
        <div v-if="details.item" class="muted">
          <RouterLink class="lnk" :to="{ name: 'project-details', params: { projectId } }">{{ details.item.name }}</RouterLink>
        </div>
        <div v-else class="muted mono">{{ t("pages.projectMembers.projectId") }}: {{ projectId }}</div>
      </div>
      <div class="actions">
        <RouterLink class="btn equal" :to="{ name: 'projects' }">{{ t("common.back") }}</RouterLink>
        <button class="btn equal" type="button" @click="load" :disabled="members.loading">{{ t("common.refresh") }}</button>
      </div>
    </div>

    <div v-if="members.error" class="err">{{ t(members.error.messageKey) }}</div>

    <table v-if="members.items.length" class="tbl">
      <thead>
        <tr>
          <th>{{ t("pages.projectMembers.email") }}</th>
          <th class="center">{{ t("pages.projectMembers.userId") }}</th>
          <th class="center">{{ t("pages.projectMembers.role") }}</th>
          <th class="center">{{ t("pages.projectMembers.learningOverride") }}</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="m in members.items" :key="m.user_id">
          <td>{{ m.email }}</td>
          <td class="mono center">{{ m.user_id }}</td>
          <td class="center">
            <select class="sel mono" v-model="m.role">
              <option value="read">{{ t("roles.read") }}</option>
              <option value="read_write">{{ t("roles.readWrite") }}</option>
              <option value="admin">{{ t("roles.admin") }}</option>
            </select>
          </td>
          <td class="center">
            <select class="sel mono" v-model="m.learning_mode_override">
              <option :value="null">{{ t("pages.projectMembers.inherit") }}</option>
              <option :value="true">{{ t("bool.true") }}</option>
              <option :value="false">{{ t("bool.false") }}</option>
            </select>
          </td>
          <td class="right">
            <div class="actions-row">
              <button class="btn primary" type="button" @click="save(m)" :disabled="members.saving">
                {{ t("common.save") }}
              </button>
              <button
                v-if="auth.isPlatformOwner"
                class="btn danger"
                type="button"
                @click="askRemove(m.user_id, m.email)"
                :disabled="members.removing"
              >
                {{ t("common.delete") }}
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-else class="muted">{{ t("states.noMembers") }}</div>

    <template v-if="auth.isPlatformOwner">
      <div class="sep"></div>

      <div class="row">
        <h3>{{ t("pages.projectMembers.addTitle") }}</h3>
      </div>

      <div class="form">
        <label>
          <div class="lbl">{{ t("pages.projectMembers.email") }}</div>
          <input v-model="newEmail" class="inp" :placeholder="t('placeholders.userEmail')" />
        </label>

        <label>
          <div class="lbl">{{ t("pages.projectMembers.role") }}</div>
          <select class="sel mono" v-model="newRole">
            <option value="read">{{ t("roles.read") }}</option>
            <option value="read_write">{{ t("roles.readWrite") }}</option>
            <option value="admin">{{ t("roles.admin") }}</option>
          </select>
        </label>

        <button class="btn primary" type="button" @click="add" :disabled="members.adding">
          {{ t("common.createOrUpdate") }}
        </button>
      </div>

      <div v-if="members.addError" class="err">{{ t(members.addError.messageKey) }}</div>
    </template>
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
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import ConfirmModal from "../shared/ui/ConfirmModal.vue";
import { useAuthStore } from "../features/auth/store";
import { useProjectMembersStore } from "../features/projects/members-store";
import { useProjectDetailsStore } from "../features/projects/details-store";
import type { ProjectMember } from "../features/projects/types";

const props = defineProps<{ projectId: string }>();

const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();
const members = useProjectMembersStore();
const details = useProjectDetailsStore();

const newEmail = ref("");
const newRole = ref<"read" | "read_write" | "admin">("read");

const confirmOpen = ref(false);
const confirmUserId = ref("");
const confirmName = ref("");

async function load() {
  await details.load(props.projectId);
  await members.load(props.projectId);
}

async function save(m: ProjectMember) {
  await members.save({
    user_id: m.user_id,
    role: m.role,
    learning_mode_override: m.learning_mode_override ?? null,
  });
}

async function add() {
  await members.addByEmail(newEmail.value, newRole.value);
  if (!members.addError) {
    newEmail.value = "";
    newRole.value = "read";
  }
}

function askRemove(userId: string, email: string) {
  confirmUserId.value = userId;
  confirmName.value = email;
  confirmOpen.value = true;
}

async function doRemove() {
  const id = confirmUserId.value;
  confirmOpen.value = false;
  confirmUserId.value = "";
  if (!id) return;
  await members.remove(id);
}

onMounted(() => void load());
</script>

<style scoped>
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
.actions-row {
  display: inline-flex;
  gap: 10px;
  align-items: center;
}
.form {
  display: grid;
  grid-template-columns: 1fr 200px auto;
  gap: 12px;
  margin-top: 12px;
  align-items: end;
}
@media (max-width: 960px) {
  .form {
    grid-template-columns: 1fr;
  }
}
</style>
