<template>
  <TableScaffoldPage
    titleKey="pages.cluster.pods.title"
    hintKey="pages.cluster.pods.hint"
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
        :to="{ name: 'cluster-pods-details', params: { name: String(item.name) } }"
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
              :to="{ name: 'cluster-pods-details', params: { name: String(item.name) } }"
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
// TODO(#19): Подключить реальные Pods + отдельный экран/таб Logs (Monaco не нужен; для логов использовать logs viewer),
// а также action preview перед destructive действиями.
import { computed, ref } from "vue";
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import ConfirmDialog from "../../shared/ui/ConfirmDialog.vue";
import DismissibleWarningAlert from "../../shared/ui/DismissibleWarningAlert.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { useUiContextStore } from "../../features/ui-context/store";

type PodRow = {
  name: string;
  namespace: string;
  ready: string;
  status: string;
  restarts: number;
  age: string;
};

const { t } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();
const uiContext = useUiContextStore();

const headers = [
  { key: "name" },
  { key: "namespace", width: 220 },
  { key: "ready", width: 120 },
  { key: "status", width: 160 },
  { key: "restarts", width: 120 },
  { key: "age", width: 120 },
  { key: "actions", sortable: false, width: 132 },
] as const;

const rows: PodRow[] = [
  { name: "codex-k8s-6b7c79d7b9-2rj6q", namespace: "codex-k8s-prod", ready: "1/1", status: "Running", restarts: 0, age: "2h" },
  { name: "api-gateway-7f7cc8dbb7-7d2s9", namespace: "codex-k8s-prod", ready: "1/1", status: "Running", restarts: 1, age: "2h" },
  { name: "web-console-6d6fd8c8d6-h9v2p", namespace: "codex-k8s-prod", ready: "1/1", status: "Running", restarts: 0, age: "2h" },
  { name: "codex-k8s-worker-6bfbd7b8f9-6k1z2", namespace: "codex-k8s-prod", ready: "1/1", status: "Running", restarts: 0, age: "2h" },
  { name: "agent-runner-27184120-zzr8m", namespace: "codex-k8s-dev-1", ready: "1/1", status: "Running", restarts: 0, age: "15m" },
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
