<template>
  <TableScaffoldPage
    titleKey="pages.cluster.jobs.title"
    hintKey="pages.cluster.jobs.hint"
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
        :to="{ name: 'cluster-jobs-details', params: { name: String(item.name) } }"
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
              :to="{ name: 'cluster-jobs-details', params: { name: String(item.name) } }"
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
        <VTooltip :text="t('common.stop')">
          <template #activator="{ props: tipProps }">
            <VBtn
              v-bind="tipProps"
              size="small"
              variant="text"
              icon="mdi-stop-circle-outline"
              :disabled="destructiveDisabled"
              @click="askAction('stop', String(item.name))"
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
    :danger="confirmAction === 'delete' || confirmAction === 'stop'"
    @confirm="confirmActionNow"
  />
</template>

<script setup lang="ts">
// TODO(#19): Подключить реальные Jobs + логи контейнеров (logs viewer), табы Overview/YAML/Events/Related/Logs и action preview.
import { computed, ref } from "vue";
import AdminClusterContextBar from "../../shared/ui/AdminClusterContextBar.vue";
import ConfirmDialog from "../../shared/ui/ConfirmDialog.vue";
import DismissibleWarningAlert from "../../shared/ui/DismissibleWarningAlert.vue";
import TableScaffoldPage from "../../shared/ui/scaffold/TableScaffoldPage.vue";
import { useI18n } from "vue-i18n";
import { useSnackbarStore } from "../../shared/ui/feedback/snackbar-store";
import { useUiContextStore } from "../../features/ui-context/store";

type JobRow = {
  name: string;
  namespace: string;
  completions: string;
  duration: string;
  age: string;
};

const { t } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();
const uiContext = useUiContextStore();

const headers = [
  { key: "name" },
  { key: "namespace", width: 220 },
  { key: "completions", width: 140 },
  { key: "duration", width: 140 },
  { key: "age", width: 120 },
  { key: "actions", sortable: false, width: 176 },
] as const;

const rows: JobRow[] = [
  { name: "agent-runner-27184120", namespace: "codex-k8s-dev-1", completions: "0/1", duration: "15m", age: "15m" },
  { name: "db-migrate-27184001", namespace: "codex-k8s-prod", completions: "1/1", duration: "32s", age: "10d" },
  { name: "smoke-check-27183012", namespace: "codex-k8s-prod", completions: "1/1", duration: "58s", age: "10d" },
];

const destructiveDisabled = computed(() => uiContext.clusterMode === "view-only");

const confirmOpen = ref(false);
const confirmAction = ref<"restart" | "stop" | "delete">("restart");
const confirmName = ref("");

const confirmTitle = computed(() => {
  switch (confirmAction.value) {
    case "delete":
      return t("common.delete");
    case "stop":
      return t("common.stop");
    default:
      return t("common.restart");
  }
});
const confirmConfirmText = computed(() => confirmTitle.value);

function askAction(action: "restart" | "stop" | "delete", name: string) {
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
