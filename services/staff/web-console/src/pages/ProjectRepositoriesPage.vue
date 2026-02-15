<template>
  <div>
    <PageHeader :title="t('pages.projectRepositories.title')">
      <template #leading>
        <AdaptiveBtn variant="text" icon="mdi-arrow-left" :label="t('common.back')" :to="{ name: 'projects' }" />
      </template>
      <template #actions>
        <CopyChip :label="t('pages.projectRepositories.projectId')" :value="projectId" icon="mdi-identifier" />
        <AdaptiveBtn variant="tonal" icon="mdi-refresh" :label="t('common.refresh')" :loading="repos.loading" @click="load" />
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

    <VAlert v-if="repos.error" type="error" variant="tonal" class="mt-4">
      {{ t(repos.error.messageKey) }}
    </VAlert>
    <VAlert v-if="repos.attachError" type="error" variant="tonal" class="mt-4">
      {{ t(repos.attachError.messageKey) }}
    </VAlert>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VDataTable :headers="headers" :items="repos.items" :loading="repos.loading" :items-per-page="10" hover>
          <template #item.repo="{ item }">
            <span class="mono text-medium-emphasis">{{ item.owner }}/{{ item.name }}</span>
          </template>

          <template #item.services_yaml_path="{ item }">
            <span class="mono text-medium-emphasis">{{ item.services_yaml_path }}</span>
          </template>

          <template #item.actions="{ item }">
            <div class="d-flex justify-end">
              <VTooltip :text="t('common.delete')">
                <template #activator="{ props: tipProps }">
                  <VBtn
                    v-bind="tipProps"
                    size="small"
                    color="error"
                    variant="tonal"
                    icon="mdi-delete-outline"
                    :loading="repos.removing"
                    @click="askRemove(item.id, `${item.owner}/${item.name}`)"
                  />
                </template>
              </VTooltip>
            </div>
          </template>

          <template #no-data>
            <div class="py-8 text-medium-emphasis">
              {{ t("states.noRepos") }}
            </div>
          </template>
        </VDataTable>
      </VCardText>
    </VCard>

    <VCard class="mt-6" variant="outlined">
      <VCardTitle class="text-subtitle-1">{{ t("pages.projectRepositories.attachTitle") }}</VCardTitle>
      <VCardText>
        <VRow density="compact" class="align-end">
          <VCol cols="12" md="4">
            <VTextField v-model.trim="owner" :label="t('pages.projectRepositories.owner')" :placeholder="t('placeholders.repoOwner')" />
          </VCol>
          <VCol cols="12" md="4">
            <VTextField v-model.trim="name" :label="t('pages.projectRepositories.name')" :placeholder="t('placeholders.repoName')" />
          </VCol>
          <VCol cols="12" md="4">
            <VTextField
              v-model.trim="servicesYamlPath"
              :label="t('pages.projectRepositories.servicesYamlPath')"
              :placeholder="t('placeholders.servicesYamlPath')"
            />
          </VCol>
          <VCol cols="12">
            <VTextField
              v-model="token"
              :label="t('pages.projectRepositories.repoToken')"
              :placeholder="t('placeholders.repoToken')"
              type="password"
            />
          </VCol>
          <VCol cols="12">
            <VBtn color="primary" variant="tonal" :loading="repos.attaching" @click="attach">
              {{ t("common.attachEnsureWebhook") }}
            </VBtn>
          </VCol>
        </VRow>
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
import AdaptiveBtn from "../shared/ui/AdaptiveBtn.vue";
import { useSnackbarStore } from "../shared/ui/feedback/snackbar-store";
import { useProjectRepositoriesStore } from "../features/projects/repositories-store";
import { useProjectDetailsStore } from "../features/projects/details-store";

const props = defineProps<{ projectId: string }>();

const { t } = useI18n({ useScope: "global" });
const repos = useProjectRepositoriesStore();
const details = useProjectDetailsStore();
const snackbar = useSnackbarStore();

const owner = ref("");
const name = ref("");
const servicesYamlPath = ref("services.yaml");
const token = ref("");

const confirmOpen = ref(false);
const confirmRepoId = ref("");
const confirmName = ref("");

const headers = [
  { title: t("pages.projectRepositories.provider"), key: "provider", width: 160, align: "start" },
  { title: t("pages.projectRepositories.repo"), key: "repo", sortable: false, width: 260, align: "center" },
  { title: t("pages.projectRepositories.servicesYaml"), key: "services_yaml_path", width: 220, align: "center" },
  { title: "", key: "actions", sortable: false, width: 72, align: "end" },
] as const;

async function load() {
  await details.load(props.projectId);
  await repos.load(props.projectId);
}

async function attach() {
  await repos.attach({ owner: owner.value, name: name.value, token: token.value, servicesYamlPath: servicesYamlPath.value });
  if (!repos.attachError) {
    owner.value = "";
    name.value = "";
    token.value = "";
    servicesYamlPath.value = "services.yaml";
    snackbar.success(t("common.saved"));
  }
}

function askRemove(repositoryId: string, label: string) {
  confirmRepoId.value = repositoryId;
  confirmName.value = label;
  confirmOpen.value = true;
}

async function doRemove() {
  const id = confirmRepoId.value;
  confirmRepoId.value = "";
  if (!id) return;
  await repos.remove(id);
  snackbar.success(t("common.deleted"));
}

onMounted(() => void load());
</script>

<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
