<template>
  <div>
    <PageHeader :title="t('pages.users.title')">
      <template #actions>
        <VBtn variant="tonal" prepend-icon="mdi-refresh" :loading="users.loading" @click="load">
          {{ t("common.refresh") }}
        </VBtn>
      </template>
    </PageHeader>

    <VRow class="mt-4" density="compact">
      <VCol cols="12" md="8">
        <VAlert v-if="users.error" type="error" variant="tonal" class="mb-4">
          {{ t(users.error.messageKey) }}
        </VAlert>
        <VAlert v-if="users.deleteError" type="error" variant="tonal" class="mb-4">
          {{ t(users.deleteError.messageKey) }}
        </VAlert>

        <VCard variant="outlined">
          <VCardText>
            <VDataTable :headers="headers" :items="users.items" :loading="users.loading" :items-per-page="10" hover>
              <template #item.github_login="{ item }">
                <span class="mono text-medium-emphasis">{{ item.github_login || "-" }}</span>
              </template>

              <template #item.is_platform_admin="{ item }">
                <VChip size="small" variant="tonal" class="font-weight-bold">
                  {{ item.is_platform_admin ? t("pages.users.yes") : t("pages.users.no") }}
                </VChip>
              </template>

              <template #item.id="{ item }">
                <span class="mono text-medium-emphasis">{{ item.id }}</span>
              </template>

              <template #item.actions="{ item }">
                <div class="d-flex justify-end">
                  <VBtn
                    v-if="canDelete(item.id, item.is_platform_admin, item.is_platform_owner)"
                    size="small"
                    color="error"
                    variant="tonal"
                    :loading="users.deleting"
                    @click="askRemove(item.id, item.email)"
                  >
                    {{ t("common.delete") }}
                  </VBtn>
                </div>
              </template>

              <template #no-data>
                <div class="py-8 text-medium-emphasis">
                  {{ t("states.noUsers") }}
                </div>
              </template>
            </VDataTable>
          </VCardText>
        </VCard>
      </VCol>

      <VCol cols="12" md="4">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-1">{{ t("pages.users.addAllowedUser") }}</VCardTitle>
          <VCardText>
            <div class="text-body-2 text-medium-emphasis mb-4">
              {{ t("pages.users.addAllowedUserHint") }}
            </div>

            <VTextField v-model.trim="email" :label="t('pages.users.email')" :placeholder="t('placeholders.userEmail')" />
            <VCheckbox v-model="isAdmin" :label="t('pages.users.platformAdmin')" />

            <VBtn class="mt-2" color="primary" variant="tonal" :loading="users.creating" @click="create">
              {{ t("common.createOrUpdate") }}
            </VBtn>

            <VAlert v-if="users.createError" type="error" variant="tonal" class="mt-4">
              {{ t(users.createError.messageKey) }}
            </VAlert>
          </VCardText>
        </VCard>
      </VCol>
    </VRow>
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
import { useI18n } from "vue-i18n";

import ConfirmDialog from "../shared/ui/ConfirmDialog.vue";
import PageHeader from "../shared/ui/PageHeader.vue";
import { useSnackbarStore } from "../shared/ui/feedback/snackbar-store";
import { useAuthStore } from "../features/auth/store";
import { useUsersStore } from "../features/users/store";

const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();
const users = useUsersStore();
const snackbar = useSnackbarStore();

const email = ref("");
const isAdmin = ref(false);

const confirmOpen = ref(false);
const confirmUserId = ref("");
const confirmName = ref("");

const headers = [
  { title: t("pages.users.email"), key: "email" },
  { title: t("pages.users.github"), key: "github_login", width: 220 },
  { title: t("pages.users.admin"), key: "is_platform_admin", width: 160 },
  { title: t("common.id"), key: "id", width: 240 },
  { title: "", key: "actions", sortable: false, width: 140 },
] as const;

async function load() {
  await users.load();
}

async function create() {
  await users.create(email.value, isAdmin.value);
  if (!users.createError) {
    email.value = "";
    isAdmin.value = false;
    snackbar.success(t("common.saved"));
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
  confirmUserId.value = "";
  if (!id) return;
  await users.remove(id);
  if (!users.deleteError) {
    snackbar.success(t("common.deleted"));
  }
}

onMounted(() => void load());
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>

