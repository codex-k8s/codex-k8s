<template>
  <section class="card">
    <div class="row">
      <div>
        <h2>{{ t("pages.runDetails.title") }}</h2>
        <div class="muted mono">
          {{ t("pages.runDetails.runId") }}: {{ runId }}
          <span v-if="details.run?.correlation_id"> · {{ t("pages.runDetails.correlation") }}: {{ details.run.correlation_id }}</span>
        </div>
        <div v-if="details.run?.project_id" class="muted">
          {{ t("pages.runDetails.project") }}:
          <RouterLink class="lnk" :to="{ name: 'project-details', params: { projectId: details.run.project_id } }">
            {{ details.run.project_name || details.run.project_slug || details.run.project_id }}
          </RouterLink>
        </div>
        <div v-if="details.run?.issue_url && details.run?.issue_number" class="muted">
          {{ t("pages.runDetails.issue") }}:
          <a class="lnk mono" :href="details.run.issue_url" target="_blank" rel="noopener noreferrer">#{{ details.run.issue_number }}</a>
        </div>
        <div v-if="details.run?.pr_url && details.run?.pr_number" class="muted">
          {{ t("pages.runDetails.pr") }}:
          <a class="lnk mono" :href="details.run.pr_url" target="_blank" rel="noopener noreferrer">#{{ details.run.pr_number }}</a>
        </div>
        <div class="muted">
          {{ t("pages.runDetails.triggerKind") }}: <span class="mono">{{ details.run?.trigger_kind || "-" }}</span>
          ·
          {{ t("pages.runDetails.triggerLabel") }}: <span class="mono">{{ details.run?.trigger_label || "-" }}</span>
        </div>
      </div>
      <div class="actions">
        <RouterLink class="btn equal" :to="{ name: 'runs' }">{{ t("common.back") }}</RouterLink>
        <button class="btn equal" type="button" @click="loadAll" :disabled="details.loading">{{ t("common.refresh") }}</button>
        <button
          v-if="canDeleteNamespace"
          class="btn equal danger"
          type="button"
          @click="askDeleteNamespace"
          :disabled="details.deletingNamespace"
        >
          {{ t("pages.runDetails.deleteNamespace") }}
        </button>
      </div>
    </div>

    <div v-if="details.error" class="err">{{ t(details.error.messageKey) }}</div>
    <div v-if="details.deleteNamespaceError" class="err">{{ t(details.deleteNamespaceError.messageKey) }}</div>
    <div v-if="details.namespaceDeleteResult" class="muted mono">
      {{ t("pages.runDetails.namespace") }}: {{ details.namespaceDeleteResult.namespace }} ·
      {{
        details.namespaceDeleteResult.already_deleted
          ? t("pages.runDetails.namespaceAlreadyDeleted")
          : t("pages.runDetails.namespaceDeleted")
      }}
    </div>

    <div class="pane runtime">
      <div class="pane-h">{{ t("pages.runDetails.job") }}</div>
      <div class="muted mono">
        {{ t("pages.runDetails.jobNamespace") }}: {{ details.run?.job_namespace || details.run?.namespace || "-" }}
      </div>
      <div class="muted mono">
        {{ t("pages.runDetails.job") }}: {{ details.run?.job_name || "-" }}
      </div>
      <div class="muted">
        <span v-if="details.run?.job_exists" class="pill">active</span>
        <span v-else>{{ t("pages.runDetails.noJob") }}</span>
      </div>
    </div>

    <div class="grid">
      <div class="pane">
        <div class="pane-h">{{ t("pages.runDetails.flowEvents") }}</div>
        <div v-if="details.events.length" class="list">
          <div v-for="e in details.events" :key="e.created_at + ':' + e.event_type" class="item">
            <div class="topline">
              <span class="pill">{{ e.event_type }}</span>
              <span class="mono muted">{{ formatDateTime(e.created_at, locale) }}</span>
            </div>
            <pre class="pre">{{ e.payload_json }}</pre>
          </div>
        </div>
        <div v-else class="muted">{{ t("states.noEvents") }}</div>
      </div>
    </div>
  </section>

  <ConfirmModal
    :open="confirmDeleteNamespaceOpen"
    :title="t('pages.runDetails.deleteNamespace')"
    :message="t('pages.runDetails.deleteNamespaceConfirm')"
    :confirmText="t('pages.runDetails.deleteNamespace')"
    :cancelText="t('common.cancel')"
    danger
    @cancel="confirmDeleteNamespaceOpen = false"
    @confirm="doDeleteNamespace"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import ConfirmModal from "../shared/ui/ConfirmModal.vue";
import { formatDateTime } from "../shared/lib/datetime";
import { useRunDetailsStore } from "../features/runs/store";

const props = defineProps<{ runId: string }>();

const { t, locale } = useI18n({ useScope: "global" });
const details = useRunDetailsStore();
const confirmDeleteNamespaceOpen = ref(false);
const canDeleteNamespace = computed(() => Boolean(details.run?.job_exists && details.run?.namespace));

async function loadAll() {
  await details.load(props.runId);
}

function askDeleteNamespace() {
  confirmDeleteNamespaceOpen.value = true;
}

async function doDeleteNamespace() {
  confirmDeleteNamespaceOpen.value = false;
  await details.deleteNamespace(props.runId);
}

onMounted(() => void loadAll());
</script>

<style scoped>
h2 {
  margin: 0;
  letter-spacing: -0.01em;
}
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  margin-top: 12px;
}
.pane {
  border: 1px solid rgba(17, 24, 39, 0.1);
  border-radius: 14px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.6);
}
.runtime {
  margin-top: 12px;
}
.pane-h {
  font-weight: 900;
  letter-spacing: -0.01em;
  margin-bottom: 10px;
}
.list {
  display: grid;
  gap: 10px;
}
.item {
  border: 1px solid rgba(17, 24, 39, 0.1);
  border-radius: 12px;
  padding: 10px;
  background: rgba(255, 255, 255, 0.7);
}
.topline {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: center;
  margin-bottom: 8px;
}
.pre {
  margin: 0;
  white-space: pre-wrap;
  font-size: 12px;
  opacity: 0.9;
}
@media (max-width: 960px) {
  .grid {
    grid-template-columns: 1fr;
  }
}
</style>
