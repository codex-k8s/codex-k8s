<template>
  <TableScaffoldPage
    titleKey="pages.cluster.secrets.title"
    hintKey="pages.cluster.secrets.hint"
    :headers="headers"
    :items="rows"
  >
    <template #below-header>
      <DismissibleWarningAlert
        alert-id="admin_cluster"
        :title="t('warnings.adminCluster.title')"
        :text="t('warnings.adminCluster.text')"
      />
      <AdminClusterContextBar />
    </template>
    <template #item.name="{ item }">
      <RouterLink
        class="text-primary font-weight-bold text-decoration-none"
        :to="{ name: 'cluster-secrets-details', params: { name: String(item.name) } }"
      >
        {{ item.name }}
      </RouterLink>
    </template>
    <template #row-actions="{ item }">
      <div class="d-flex ga-1 justify-end">
        <VTooltip :text="t('scaffold.rowActions.view')">
          <template #activator="{ props: tipProps }">
            <VBtn
              v-bind="tipProps"
              size="small"
              variant="text"
              icon="mdi-open-in-new"
              :to="{ name: 'cluster-secrets-details', params: { name: String(item.name) } }"
            />
          </template>
        </VTooltip>
        <VTooltip :text="t('common.delete')">
          <template #activator="{ props: tipProps }">
            <VBtn
              v-bind="tipProps"
              size="small"
              variant="text"
              color="error"
              icon="mdi-delete-outline"
              :disabled="destructiveDisabled"
              @click="askDelete(String(item.name))"
            />
          </template>
        </VTooltip>
      </div>
    </template>
  </TableScaffoldPage>

  <ConfirmDialog
    v-model="confirmOpen"
    :title="t('common.delete')"
    :message="confirmName"
    :confirm-text="t('common.delete')"
    :cancel-text="t('common.cancel')"
    danger
    @confirm="confirmDelete"
  />
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные Secrets (metadata-only по умолчанию), reveal как отдельное осознанное действие, YAML view/edit на Monaco.
import { computed, ref } from "vue";
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import ConfirmDialog from "../../shared/ui/ConfirmDialog.vue";
import DismissibleWarningAlert from "../../shared/ui/DismissibleWarningAlert.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { useUiContextStore } from "../../features/ui-context/store";

type SecretRow = {
  name: string;
  namespace: string;
  type: string;
  keys: number;
  age: string;
};

const { t } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();
const uiContext = useUiContextStore();

const headers = [
  { key: "name" },
  { key: "namespace", width: 220 },
  { key: "type", width: 180 },
  { key: "keys", width: 120 },
  { key: "age", width: 120 },
  { key: "actions", sortable: false, width: 96 },
] as const;

const rows: SecretRow[] = [
  { name: "postgres-credentials", namespace: "codex-k8s-ai-staging", type: "Opaque", keys: 2, age: "10d" },
  { name: "github-oauth", namespace: "codex-k8s-ai-staging", type: "Opaque", keys: 2, age: "10d" },
  { name: "runner-token", namespace: "codex-k8s-dev-1", type: "Opaque", keys: 1, age: "2h" },
  { name: "tls-cert", namespace: "codex-k8s-prod", type: "kubernetes.io/tls", keys: 2, age: "45d" },
];

const destructiveDisabled = computed(() => uiContext.clusterMode === "view-only");

const confirmOpen = ref(false);
const confirmName = ref("");

function askDelete(name: string) {
  confirmName.value = name;
  confirmOpen.value = true;
}

function confirmDelete() {
  confirmOpen.value = false;
  confirmName.value = "";
  snackbar.info(t("common.notImplementedYet"));
}
</script>
