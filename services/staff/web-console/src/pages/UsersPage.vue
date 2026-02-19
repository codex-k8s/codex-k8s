<template>
  <div>
    <PageHeader :title="t('pages.users.title')">
      <template #actions>
        <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :loading="users.loading" @click="refreshUsers" />
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
            <VDataTable v-model:page="tablePage" :headers="headers" :items="users.items" :loading="users.loading" :items-per-page="itemsPerPage" hover>
              <template #item.github_login="{ item }">
                <span class="mono text-medium-emphasis">{{ item.github_login || "-" }}</span>
              </template>

              <template #item.is_platform_admin="{ item }">
                <div class="d-flex justify-center">
                  <VChip size="small" variant="tonal" class="font-weight-bold" :color="item.is_platform_admin ? 'warning' : 'secondary'">
                    {{ item.is_platform_admin ? t("pages.users.yes") : t("pages.users.no") }}
                  </VChip>
                </div>
              </template>

              <template #item.actions="{ item }">
                <div class="d-flex justify-end">
                  <VTooltip v-if="canEdit(item.id, item.is_platform_owner)" :text="t('scaffold.rowActions.edit')">
                    <template #activator="{ props: tipProps }">
                      <VBtn
                        v-bind="tipProps"
                        size="small"
                        variant="text"
                        icon="mdi-pencil-outline"
                        :disabled="users.deleting || users.creating"
                        @click="openEditDialog(item.email, item.is_platform_admin)"
                      />
                    </template>
                  </VTooltip>
                  <VTooltip v-if="canDelete(item.id, item.is_platform_admin, item.is_platform_owner)" :text="t('common.delete')">
                    <template #activator="{ props: tipProps }">
                      <VBtn
                        v-bind="tipProps"
                        size="small"
                        color="error"
                        variant="tonal"
                        icon="mdi-delete-outline"
                        :loading="users.deleting"
                        @click="askRemove(item.id, item.email)"
                      />
                    </template>
                  </VTooltip>
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

  <VDialog v-model="editDialogOpen" max-width="520">
    <VCard>
      <VCardTitle class="text-subtitle-1">{{ t("pages.users.editTitle") }}</VCardTitle>
      <VCardText>
        <VTextField v-model="editEmail" :label="t('pages.users.email')" disabled />
        <VCheckbox v-model="editIsAdmin" :label="t('pages.users.platformAdmin')" />

        <VAlert v-if="users.createError" type="error" variant="tonal" class="mt-4">
          {{ t(users.createError.messageKey) }}
        </VAlert>
      </VCardText>
      <VCardActions>
        <VSpacer />
        <VBtn variant="text" :disabled="users.creating" @click="editDialogOpen = false">{{ t("common.cancel") }}</VBtn>
        <VBtn color="primary" variant="tonal" :loading="users.creating" @click="saveEdit">{{ t("common.save") }}</VBtn>
      </VCardActions>
    </VCard>
  </VDialog>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

import ConfirmDialog from "../shared/ui/ConfirmDialog.vue";
import PageHeader from "../shared/ui/PageHeader.vue";
import AdaptiveBtn from "../shared/ui/AdaptiveBtn.vue";
import { useSnackbarStore } from "../shared/ui/feedback/snackbar-store";
import { createProgressiveTableState } from "../shared/lib/progressive-table";
import { useAuthStore } from "../features/auth/store";
import { useUsersStore } from "../features/users/store";

const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();
const users = useUsersStore();
const snackbar = useSnackbarStore();
const itemsPerPage = 10;
const paging = createProgressiveTableState({ itemsPerPage });
const tablePage = paging.page;

const email = ref("");
const isAdmin = ref(false);

const confirmOpen = ref(false);
const confirmUserId = ref("");
const confirmName = ref("");

const editDialogOpen = ref(false);
const editEmail = ref("");
const editIsAdmin = ref(false);

const headers = [
  { title: t("pages.users.email"), key: "email", align: "start" },
  { title: t("pages.users.github"), key: "github_login", width: 220, align: "center" },
  { title: t("pages.users.admin"), key: "is_platform_admin", width: 160, align: "center" },
  { title: "", key: "actions", sortable: false, width: 112, align: "end" },
] as const;

async function load() {
  await users.load(paging.limit.value);
  paging.markLoaded(users.items.length);
}

async function refreshUsers(): Promise<void> {
  paging.reset();
  await load();
}

async function loadMoreUsersIfNeeded(nextPage: number, prevPage: number): Promise<void> {
  if (users.loading) {
    return;
  }
  if (!paging.shouldGrowForPage(users.items.length, nextPage, prevPage)) {
    return;
  }
  await load();
}

async function create() {
  await users.create(email.value, isAdmin.value, paging.limit.value);
  paging.markLoaded(users.items.length);
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

function canEdit(userId: string, isPlatformOwner: boolean): boolean {
  if (!auth.me) return false;
  if (!auth.isPlatformAdmin) return false;
  if (isPlatformOwner) return false;
  if (userId === auth.me.id) return false;
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
  await users.remove(id, paging.limit.value);
  paging.markLoaded(users.items.length);
  if (!users.deleteError) {
    snackbar.success(t("common.deleted"));
  }
}

function openEditDialog(userEmail: string, isPlatformAdmin: boolean): void {
  users.createError = null;
  editEmail.value = userEmail;
  editIsAdmin.value = isPlatformAdmin;
  editDialogOpen.value = true;
}

async function saveEdit(): Promise<void> {
  await users.create(editEmail.value, editIsAdmin.value, paging.limit.value);
  paging.markLoaded(users.items.length);
  if (!users.createError) {
    editDialogOpen.value = false;
    snackbar.success(t("common.saved"));
  }
}

watch(
  tablePage,
  (nextPage, prevPage) => void loadMoreUsersIfNeeded(nextPage, prevPage),
);

onMounted(() => void refreshUsers());
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
