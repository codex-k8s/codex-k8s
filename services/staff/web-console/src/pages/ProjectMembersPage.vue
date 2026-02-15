<template>
  <div>
    <PageHeader :title="t('pages.projectMembers.title')">
      <template #actions>
        <CopyChip :label="t('pages.projectMembers.projectId')" :value="projectId" icon="mdi-identifier" />
        <VBtn variant="tonal" prepend-icon="mdi-arrow-left" :to="{ name: 'projects' }">
          {{ t("common.back") }}
        </VBtn>
        <VBtn variant="tonal" prepend-icon="mdi-refresh" :loading="members.loading" @click="load">
          {{ t("common.refresh") }}
        </VBtn>
      </template>
    </PageHeader>

    <div class="mt-2 text-body-2 text-medium-emphasis">
      <RouterLink
        v-if="details.item"
        class="text-primary font-weight-bold text-decoration-none"
        :to="{ name: 'project-details', params: { projectId } }"
      >
        {{ details.item.name }}
      </RouterLink>
    </div>

    <VAlert v-if="members.error" type="error" variant="tonal" class="mt-4">
      {{ t(members.error.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable :headers="headers" :items="members.items" :loading="members.loading" :items-per-page="10" hover>
          <template #item.user_id="{ item }">
            <span class="mono text-medium-emphasis">{{ item.user_id }}</span>
          </template>

          <template #item.role="{ item }">
            <VSelect v-model="item.role" :items="roleOptions" density="compact" hide-details />
          </template>

          <template #item.learning_mode_override="{ item }">
            <VSelect v-model="item.learning_mode_override" :items="learningOptions" density="compact" hide-details />
          </template>

          <template #item.actions="{ item }">
            <div class="d-flex ga-2 justify-end flex-wrap">
              <VBtn size="small" color="primary" variant="tonal" :loading="members.saving" @click="save(item)">
                {{ t("common.save") }}
              </VBtn>
              <VBtn
                v-if="auth.isPlatformOwner"
                size="small"
                color="error"
                variant="tonal"
                :loading="members.removing"
                @click="askRemove(item.user_id, item.email)"
              >
                {{ t("common.delete") }}
              </VBtn>
            </div>
          </template>

          <template #no-data>
            <div class="py-8 text-medium-emphasis">
              {{ t("states.noMembers") }}
            </div>
          </template>
        </VDataTable>
      </VCardText>
    </VCard>

    <VCard v-if="auth.isPlatformOwner" class="mt-6" variant="outlined">
      <VCardTitle class="text-subtitle-1">{{ t("pages.projectMembers.addTitle") }}</VCardTitle>
      <VCardText>
        <VRow density="compact" class="align-end">
          <VCol cols="12" md="6">
            <VTextField v-model.trim="newEmail" :label="t('pages.projectMembers.email')" :placeholder="t('placeholders.userEmail')" />
          </VCol>
          <VCol cols="12" md="4">
            <VSelect v-model="newRole" :items="roleOptions" :label="t('pages.projectMembers.role')" />
          </VCol>
          <VCol cols="12" md="2">
            <VBtn class="w-100" color="primary" variant="tonal" :loading="members.adding" @click="add">
              {{ t("common.createOrUpdate") }}
            </VBtn>
          </VCol>
        </VRow>

        <VAlert v-if="members.addError" type="error" variant="tonal" class="mt-4">
          {{ t(members.addError.messageKey) }}
        </VAlert>
      </VCardText>
    </VCard>
  </div>

  <ConfirmDialog
    v-model="confirmOpen"
    :title="t('common.delete')"
    :message="confirmName"
    :confirm-text="t('common.delete')"
    :cancel-text="t('common.cancel')"
    danger
    @confirm="doRemove"
  />
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import ConfirmDialog from "../shared/ui/ConfirmDialog.vue";
import CopyChip from "../shared/ui/CopyChip.vue";
import PageHeader from "../shared/ui/PageHeader.vue";
import { useSnackbarStore } from "../shared/ui/feedback/snackbar-store";
import { useAuthStore } from "../features/auth/store";
import { useProjectMembersStore } from "../features/projects/members-store";
import { useProjectDetailsStore } from "../features/projects/details-store";
import type { ProjectMember } from "../features/projects/types";

const props = defineProps<{ projectId: string }>();

const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();
const members = useProjectMembersStore();
const details = useProjectDetailsStore();
const snackbar = useSnackbarStore();

const newEmail = ref("");
const newRole = ref<"read" | "read_write" | "admin">("read");

const confirmOpen = ref(false);
const confirmUserId = ref("");
const confirmName = ref("");

const roleOptions = [
  { title: t("roles.read"), value: "read" },
  { title: t("roles.readWrite"), value: "read_write" },
  { title: t("roles.admin"), value: "admin" },
] as const;

const learningOptions = [
  { title: t("pages.projectMembers.inherit"), value: null },
  { title: t("bool.true"), value: true },
  { title: t("bool.false"), value: false },
] as const;

const headers = [
  { title: t("pages.projectMembers.email"), key: "email" },
  { title: t("pages.projectMembers.userId"), key: "user_id", width: 260 },
  { title: t("pages.projectMembers.role"), key: "role", width: 220, sortable: false },
  { title: t("pages.projectMembers.learningOverride"), key: "learning_mode_override", width: 240, sortable: false },
  { title: "", key: "actions", sortable: false, width: 220 },
] as const;

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
  snackbar.success(t("common.saved"));
}

async function add() {
  await members.addByEmail(newEmail.value, newRole.value);
  if (!members.addError) {
    newEmail.value = "";
    newRole.value = "read";
    snackbar.success(t("common.saved"));
  }
}

function askRemove(userId: string, email: string) {
  confirmUserId.value = userId;
  confirmName.value = email;
  confirmOpen.value = true;
}

async function doRemove() {
  const id = confirmUserId.value;
  confirmUserId.value = "";
  if (!id) return;
  await members.remove(id);
  snackbar.success(t("common.deleted"));
}

onMounted(() => void load());
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>

