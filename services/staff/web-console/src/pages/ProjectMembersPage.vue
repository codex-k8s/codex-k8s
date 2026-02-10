<template>
  <section class="card">
    <div class="row">
      <div>
        <h2>{{ t("pages.projectMembers.title") }}</h2>
        <div class="muted mono">{{ t("pages.projectMembers.projectId") }}: {{ projectId }}</div>
      </div>
      <div class="actions">
        <RouterLink class="btn" :to="{ name: 'projects' }">{{ t("common.back") }}</RouterLink>
        <button class="btn" type="button" @click="load" :disabled="members.loading">{{ t("common.refresh") }}</button>
      </div>
    </div>

    <div v-if="members.error" class="err">{{ t(members.error.messageKey) }}</div>

    <table v-if="members.items.length" class="tbl">
      <thead>
        <tr>
          <th>{{ t("pages.projectMembers.email") }}</th>
          <th>{{ t("pages.projectMembers.userId") }}</th>
          <th>{{ t("pages.projectMembers.role") }}</th>
          <th>{{ t("pages.projectMembers.learningOverride") }}</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="m in members.items" :key="m.userId">
          <td>{{ m.email }}</td>
          <td class="mono">{{ m.userId }}</td>
          <td>
            <select class="sel mono" v-model="m.role">
              <option value="read">{{ t("roles.read") }}</option>
              <option value="read_write">{{ t("roles.readWrite") }}</option>
              <option value="admin">{{ t("roles.admin") }}</option>
            </select>
          </td>
          <td>
            <select class="sel mono" v-model="m.learningModeOverride">
              <option :value="null">{{ t("pages.projectMembers.inherit") }}</option>
              <option :value="true">{{ t("bool.true") }}</option>
              <option :value="false">{{ t("bool.false") }}</option>
            </select>
          </td>
          <td class="right">
            <button class="btn primary" type="button" @click="save(m)" :disabled="members.saving">
              {{ t("common.save") }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-else class="muted">{{ t("states.noMembers") }}</div>
  </section>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import { useProjectMembersStore } from "../features/projects/members-store";
import type { ProjectMember } from "../features/projects/types";

const props = defineProps<{ projectId: string }>();

const { t } = useI18n({ useScope: "global" });
const members = useProjectMembersStore();

async function load() {
  await members.load(props.projectId);
}

async function save(m: ProjectMember) {
  await members.save({ userId: m.userId, role: m.role, learningModeOverride: m.learningModeOverride });
}

onMounted(() => void load());
</script>

<style scoped>
h2 {
  margin: 0;
  letter-spacing: -0.01em;
}
</style>
