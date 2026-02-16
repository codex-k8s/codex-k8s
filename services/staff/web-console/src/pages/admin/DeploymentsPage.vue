<template>
  <TableScaffoldPage
    titleKey="pages.cluster.deployments.title"
    hintKey="pages.cluster.deployments.hint"
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
        :to="{ name: 'cluster-deployments-details', params: { name: String(item.name) } }"
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
              :to="{ name: 'cluster-deployments-details', params: { name: String(item.name) } }"
            />
          </template>
        </VTooltip>
        <VTooltip :text="t('common.restart')">
          <template #activator="{ props: tipProps }">
            <VBtn
              v-bind="tipProps"
              size="small"
              variant="text"
              icon="mdi-restart"
              :disabled="destructiveDisabled"
              @click="askAction('restart', String(item.name))"
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
              @click="askAction('delete', String(item.name))"
            />
          </template>
        </VTooltip>
      </div>
    </template>
  </TableScaffoldPage>

  <ConfirmDialog
    v-model="confirmOpen"
    :title="confirmTitle"
    :message="confirmName"
    :confirm-text="confirmConfirmText"
    :cancel-text="t('common.cancel')"
    :danger="confirmAction === 'delete'"
    @confirm="confirmActionNow"
  />
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные Deployments (list/get), табы Overview/YAML/Events/Related, и guardrails:
// - platform deployments (app.kubernetes.io/part-of=codex-k8s) в production/prod = view-only
import { computed, ref } from "vue";
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import ConfirmDialog from "../../shared/ui/ConfirmDialog.vue";
import DismissibleWarningAlert from "../../shared/ui/DismissibleWarningAlert.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { useUiContextStore } from "../../features/ui-context/store";

type DeploymentRow = {
  name: string;
  namespace: string;
  ready: string;
  updated: string;
  age: string;
};

const { t } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();
const uiContext = useUiContextStore();

const headers = [
  { key: "name" },
  { key: "namespace", width: 220 },
  { key: "ready", width: 120 },
  { key: "updated", width: 160 },
  { key: "age", width: 120 },
  { key: "actions", sortable: false, width: 132 },
] as const;

const rows: DeploymentRow[] = [
  { name: "codex-k8s", namespace: "codex-k8s-prod", ready: "1/1", updated: "2026-02-15", age: "10d" },
  { name: "codex-k8s-worker", namespace: "codex-k8s-prod", ready: "1/1", updated: "2026-02-15", age: "10d" },
  { name: "api-gateway", namespace: "codex-k8s-prod", ready: "1/1", updated: "2026-02-15", age: "10d" },
  { name: "web-console", namespace: "codex-k8s-prod", ready: "1/1", updated: "2026-02-15", age: "10d" },
];

const destructiveDisabled = computed(() => uiContext.clusterMode === "view-only");

const confirmOpen = ref(false);
const confirmAction = ref<"restart" | "delete">("restart");
const confirmName = ref("");

const confirmTitle = computed(() => (confirmAction.value === "delete" ? t("common.delete") : t("common.restart")));
const confirmConfirmText = computed(() => (confirmAction.value === "delete" ? t("common.delete") : t("common.restart")));

function askAction(action: "restart" | "delete", name: string) {
  confirmAction.value = action;
  confirmName.value = name;
  confirmOpen.value = true;
}

function confirmActionNow() {
  confirmOpen.value = false;
  confirmName.value = "";
  snackbar.info(t("common.notImplementedYet"));
}
</script>
