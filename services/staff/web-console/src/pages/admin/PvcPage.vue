<template>
  <TableScaffoldPage
    titleKey="pages.cluster.pvc.title"
    hintKey="pages.cluster.pvc.hint"
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
        :to="{ name: 'cluster-pvc-details', params: { name: String(item.name) } }"
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
              :to="{ name: 'cluster-pvc-details', params: { name: String(item.name) } }"
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
// TODO(#19): Подключить реальные PVC (list/get) и правила view-only/dry-run для окружений.
import { computed, ref } from "vue";
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import ConfirmDialog from "../../shared/ui/ConfirmDialog.vue";
import DismissibleWarningAlert from "../../shared/ui/DismissibleWarningAlert.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { useUiContextStore } from "../../features/ui-context/store";

type PvcRow = {
  name: string;
  namespace: string;
  status: string;
  volume: string;
  capacity: string;
  accessModes: string;
  storageClass: string;
  age: string;
};

const { t } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();
const uiContext = useUiContextStore();

const headers = [
  { key: "name" },
  { key: "namespace", width: 220 },
  { key: "status", width: 140 },
  { key: "volume", width: 180 },
  { key: "capacity", width: 120 },
  { key: "accessModes", width: 120 },
  { key: "storageClass", width: 160 },
  { key: "age", width: 120 },
  { key: "actions", sortable: false, width: 96 },
] as const;

const rows: PvcRow[] = [
  {
    name: "postgres-data",
    namespace: "codex-k8s-prod",
    status: "Bound",
    volume: "pvc-7b2f...",
    capacity: "20Gi",
    accessModes: "RWO",
    storageClass: "standard",
    age: "10d",
  },
  {
    name: "pgvector-data",
    namespace: "codex-k8s-prod",
    status: "Bound",
    volume: "pvc-3c19...",
    capacity: "50Gi",
    accessModes: "RWO",
    storageClass: "standard",
    age: "10d",
  },
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
