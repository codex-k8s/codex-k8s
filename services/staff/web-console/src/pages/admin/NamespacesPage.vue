<template>
  <TableScaffoldPage
    titleKey="pages.cluster.namespaces.title"
    hintKey="pages.cluster.namespaces.hint"
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
        :to="{ name: 'cluster-namespaces-details', params: { name: String(item.name) } }"
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
              :to="{ name: 'cluster-namespaces-details', params: { name: String(item.name) } }"
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
// TODO(#19): Подключить реальные namespaces из backend (k8s API + RBAC + audit) и правила режимов:
// - ai-staging/prod для platform resources (app.kubernetes.io/part-of=codex-k8s) = view-only
// - ai env = destructive actions через backend dry-run, с явным feedback
import { computed, ref } from "vue";
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import ConfirmDialog from "../../shared/ui/ConfirmDialog.vue";
import DismissibleWarningAlert from "../../shared/ui/DismissibleWarningAlert.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { useUiContextStore } from "../../features/ui-context/store";

type NamespaceRow = {
  name: string;
  status: "Active" | "Terminating";
  age: string;
  part_of: "codex-k8s" | "-";
};

const { t } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();
const uiContext = useUiContextStore();

const headers = [
  { key: "name" },
  { key: "status", width: 160 },
  { key: "age", width: 120 },
  { key: "part_of", width: 140 },
  { key: "actions", sortable: false, width: 96 },
] as const;

const rows: NamespaceRow[] = [
  { name: "codex-k8s-dev-1", status: "Active", age: "3h", part_of: "-" },
  { name: "codex-k8s-dev-2", status: "Active", age: "1d", part_of: "-" },
  { name: "codex-k8s-ai-staging", status: "Active", age: "12d", part_of: "codex-k8s" },
  { name: "codex-k8s-prod", status: "Active", age: "45d", part_of: "codex-k8s" },
];

const destructiveDisabled = computed(() => uiContext.clusterMode === "view-only");

const confirmOpen = ref(false);
const confirmName = ref("");
const confirmTargetName = ref("");

function askDelete(name: string) {
  confirmTargetName.value = name;
  confirmName.value = name;
  confirmOpen.value = true;
}

function confirmDelete() {
  confirmOpen.value = false;
  confirmName.value = "";
  confirmTargetName.value = "";
  snackbar.info(t("common.notImplementedYet"));
}
</script>
