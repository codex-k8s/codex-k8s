<template>
  <section class="card">
    <div class="row">
      <h2>{{ t("pages.runs.title") }}</h2>
      <button class="btn" type="button" @click="loadAll" :disabled="runs.loading">
        {{ t("common.refresh") }}
      </button>
    </div>

    <div v-if="runs.error" class="err">{{ t(runs.error.messageKey) }}</div>
    <div v-if="runs.approvalsError" class="err">{{ t(runs.approvalsError.messageKey) }}</div>

    <table v-if="runs.items.length" class="tbl">
      <thead>
        <tr>
          <th class="center">{{ t("pages.runs.status") }}</th>
          <th class="center">{{ t("pages.runs.project") }}</th>
          <th class="center">{{ t("pages.runs.issue") }}</th>
          <th class="center">{{ t("pages.runs.pr") }}</th>
          <th class="center">{{ t("pages.runs.runType") }}</th>
          <th class="center">{{ t("pages.runs.triggerLabel") }}</th>
          <th class="center">{{ t("pages.runs.started") }}</th>
          <th class="center">{{ t("pages.runs.finished") }}</th>
          <th class="center"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in pageItems" :key="r.id">
          <td class="center">
            <span class="pill" :class="'s-' + r.status">{{ r.status }}</span>
          </td>
          <td class="center">
            <RouterLink v-if="r.project_id" class="lnk" :to="{ name: 'project-details', params: { projectId: r.project_id } }">
              {{ r.project_name || r.project_slug || r.project_id }}
            </RouterLink>
            <span v-else class="mono">-</span>
          </td>
          <td class="center">
            <a v-if="r.issue_url && r.issue_number" class="lnk mono" :href="r.issue_url" target="_blank" rel="noopener noreferrer">
              #{{ r.issue_number }}
            </a>
            <span v-else class="mono">-</span>
          </td>
          <td class="center">
            <a v-if="r.pr_url && r.pr_number" class="lnk mono" :href="r.pr_url" target="_blank" rel="noopener noreferrer">
              #{{ r.pr_number }}
            </a>
            <span v-else class="mono">-</span>
          </td>
          <td class="center">
            <span class="pill run-badge mono">{{ runBadgeValue(r.trigger_kind) }}</span>
          </td>
          <td class="center">
            <span class="pill run-badge mono">{{ runBadgeValue(r.trigger_label) }}</span>
          </td>
          <td class="mono center">{{ formatDateTime(r.started_at, locale) }}</td>
          <td class="mono center">{{ formatDateTime(r.finished_at, locale) }}</td>
          <td class="center">
            <RouterLink class="lnk" :to="{ name: 'run-details', params: { runId: r.id } }">
              {{ t("pages.runs.details") }}
            </RouterLink>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-else class="muted">{{ t("states.noRuns") }}</div>

    <div v-if="runs.items.length" class="row pager">
      <button class="btn equal" type="button" @click="prevPage" :disabled="currentPage <= 1">
        {{ t("pages.runs.prevPage") }}
      </button>
      <span class="mono">{{ t("pages.runs.pageInfo", { current: currentPage, total: totalPages }) }}</span>
      <button class="btn equal" type="button" @click="nextPage" :disabled="currentPage >= totalPages">
        {{ t("pages.runs.nextPage") }}
      </button>
    </div>

    <div class="pane runtime-pane">
      <div class="row">
        <h3>{{ t("pages.runs.runningJobs") }}</h3>
        <button class="btn" type="button" @click="runs.loadRunJobs()" :disabled="runs.jobsLoading">
          {{ t("common.refresh") }}
        </button>
      </div>
      <div class="row runtime-filters">
        <label class="runtime-filter">
          <span class="muted">{{ t("pages.runs.runType") }}</span>
          <input v-model.trim="runs.jobsFilters.triggerKind" class="in" type="text" />
        </label>
        <label class="runtime-filter">
          <span class="muted">{{ t("pages.runs.status") }}</span>
          <input v-model.trim="runs.jobsFilters.status" class="in" type="text" />
        </label>
        <label class="runtime-filter">
          <span class="muted">{{ t("pages.runs.agentKey") }}</span>
          <input v-model.trim="runs.jobsFilters.agentKey" class="in" type="text" />
        </label>
        <button class="btn" type="button" @click="runs.loadRunJobs()" :disabled="runs.jobsLoading">
          {{ t("common.refresh") }}
        </button>
      </div>
      <table v-if="runs.runningJobs.length" class="tbl">
        <thead>
          <tr>
            <th class="center">{{ t("pages.runs.status") }}</th>
            <th class="center">{{ t("pages.runs.project") }}</th>
            <th class="center">{{ t("pages.runs.issue") }}</th>
            <th class="center">{{ t("pages.runs.pr") }}</th>
            <th class="center">{{ t("pages.runs.runType") }}</th>
            <th class="center">{{ t("pages.runs.agentKey") }}</th>
            <th class="center">{{ t("pages.runs.jobNamespace") }}</th>
            <th class="center">{{ t("pages.runs.jobName") }}</th>
            <th class="center">{{ t("pages.runs.started") }}</th>
            <th class="center"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in runs.runningJobs" :key="`job-${r.id}`">
            <td class="center"><span class="pill" :class="'s-' + r.status">{{ r.status }}</span></td>
            <td class="center">
              <RouterLink v-if="r.project_id" class="lnk" :to="{ name: 'project-details', params: { projectId: r.project_id } }">
                {{ r.project_name || r.project_slug || r.project_id }}
              </RouterLink>
              <span v-else class="mono">-</span>
            </td>
            <td class="center">
              <a v-if="r.issue_url && r.issue_number" class="lnk mono" :href="r.issue_url" target="_blank" rel="noopener noreferrer">
                #{{ r.issue_number }}
              </a>
              <span v-else class="mono">-</span>
            </td>
            <td class="center">
              <a v-if="r.pr_url && r.pr_number" class="lnk mono" :href="r.pr_url" target="_blank" rel="noopener noreferrer">
                #{{ r.pr_number }}
              </a>
              <span v-else class="mono">-</span>
            </td>
            <td class="center"><span class="pill run-badge mono">{{ runBadgeValue(r.trigger_kind) }}</span></td>
            <td class="center"><span class="pill run-badge mono">{{ runBadgeValue(r.agent_key) }}</span></td>
            <td class="mono center">{{ runBadgeValue(r.job_namespace || r.namespace) }}</td>
            <td class="mono center">{{ runBadgeValue(r.job_name) }}</td>
            <td class="mono center">{{ formatDateTime(r.started_at, locale) }}</td>
            <td class="center">
              <RouterLink class="lnk" :to="{ name: 'run-details', params: { runId: r.id } }">
                {{ t("pages.runs.details") }}
              </RouterLink>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="muted">{{ t("states.noRunningJobs") }}</div>
    </div>

    <div class="pane runtime-pane">
      <div class="row">
        <h3>{{ t("pages.runs.waitQueue") }}</h3>
        <button class="btn" type="button" @click="runs.loadRunWaits()" :disabled="runs.waitsLoading">
          {{ t("common.refresh") }}
        </button>
      </div>
      <div class="row runtime-filters">
        <label class="runtime-filter">
          <span class="muted">{{ t("pages.runs.runType") }}</span>
          <input v-model.trim="runs.waitsFilters.triggerKind" class="in" type="text" />
        </label>
        <label class="runtime-filter">
          <span class="muted">{{ t("pages.runs.status") }}</span>
          <input v-model.trim="runs.waitsFilters.status" class="in" type="text" />
        </label>
        <label class="runtime-filter">
          <span class="muted">{{ t("pages.runs.agentKey") }}</span>
          <input v-model.trim="runs.waitsFilters.agentKey" class="in" type="text" />
        </label>
        <label class="runtime-filter">
          <span class="muted">{{ t("pages.runs.waitState") }}</span>
          <input v-model.trim="runs.waitsFilters.waitState" class="in" type="text" />
        </label>
        <button class="btn" type="button" @click="runs.loadRunWaits()" :disabled="runs.waitsLoading">
          {{ t("common.refresh") }}
        </button>
      </div>
      <table v-if="runs.waitQueue.length" class="tbl">
        <thead>
          <tr>
            <th class="center">{{ t("pages.runs.status") }}</th>
            <th class="center">{{ t("pages.runs.project") }}</th>
            <th class="center">{{ t("pages.runs.issue") }}</th>
            <th class="center">{{ t("pages.runs.pr") }}</th>
            <th class="center">{{ t("pages.runs.runType") }}</th>
            <th class="center">{{ t("pages.runs.agentKey") }}</th>
            <th class="center">{{ t("pages.runs.waitState") }}</th>
            <th class="center">{{ t("pages.runs.waitSince") }}</th>
            <th class="center">{{ t("pages.runs.waitSla") }}</th>
            <th class="center">{{ t("pages.runs.lastHeartbeatAt") }}</th>
            <th class="center"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in runs.waitQueue" :key="`wait-${r.id}`">
            <td class="center"><span class="pill" :class="'s-' + r.status">{{ r.status }}</span></td>
            <td class="center">
              <RouterLink v-if="r.project_id" class="lnk" :to="{ name: 'project-details', params: { projectId: r.project_id } }">
                {{ r.project_name || r.project_slug || r.project_id }}
              </RouterLink>
              <span v-else class="mono">-</span>
            </td>
            <td class="center">
              <a v-if="r.issue_url && r.issue_number" class="lnk mono" :href="r.issue_url" target="_blank" rel="noopener noreferrer">
                #{{ r.issue_number }}
              </a>
              <span v-else class="mono">-</span>
            </td>
            <td class="center">
              <a v-if="r.pr_url && r.pr_number" class="lnk mono" :href="r.pr_url" target="_blank" rel="noopener noreferrer">
                #{{ r.pr_number }}
              </a>
              <span v-else class="mono">-</span>
            </td>
            <td class="center"><span class="pill run-badge mono">{{ runBadgeValue(r.trigger_kind) }}</span></td>
            <td class="center"><span class="pill run-badge mono">{{ runBadgeValue(r.agent_key) }}</span></td>
            <td class="center"><span class="pill run-badge mono">{{ runBadgeValue(r.wait_state) }}</span></td>
            <td class="mono center">{{ formatDateTime(r.wait_since, locale) }}</td>
            <td class="mono center">{{ formatDurationSince(r.wait_since, locale) }}</td>
            <td class="mono center">{{ formatDateTime(r.last_heartbeat_at, locale) }}</td>
            <td class="center">
              <RouterLink class="lnk" :to="{ name: 'run-details', params: { runId: r.id } }">
                {{ t("pages.runs.details") }}
              </RouterLink>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="muted">{{ t("states.noWaitQueue") }}</div>
    </div>

    <div class="pane approvals">
      <div class="row">
        <h3>{{ t("pages.runs.pendingApprovals") }}</h3>
        <button class="btn" type="button" @click="runs.loadPendingApprovals()" :disabled="runs.approvalsLoading">
          {{ t("common.refresh") }}
        </button>
      </div>
      <table v-if="runs.pendingApprovals.length" class="tbl">
        <thead>
          <tr>
            <th class="center">{{ t("pages.runs.project") }}</th>
            <th class="center">{{ t("pages.runs.issue") }}</th>
            <th class="center">{{ t("pages.runs.pr") }}</th>
            <th class="center">{{ t("pages.runs.tool") }}</th>
            <th class="center">{{ t("pages.runs.action") }}</th>
            <th class="center">{{ t("pages.runs.requestedBy") }}</th>
            <th class="center">{{ t("pages.runs.created") }}</th>
            <th class="center">{{ t("pages.runs.resolve") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in runs.pendingApprovals" :key="item.id">
            <td class="center">
              <RouterLink v-if="item.project_id" class="lnk" :to="{ name: 'project-details', params: { projectId: item.project_id } }">
                {{ item.project_name || item.project_slug || item.project_id }}
              </RouterLink>
              <span v-else class="mono">-</span>
            </td>
            <td class="center">
              <span class="mono">{{ item.issue_number ? `#${item.issue_number}` : "-" }}</span>
            </td>
            <td class="center">
              <span class="mono">{{ item.pr_number ? `#${item.pr_number}` : "-" }}</span>
            </td>
            <td class="center"><span class="pill run-badge mono">{{ item.tool_name }}</span></td>
            <td class="center"><span class="pill run-badge mono">{{ item.action }}</span></td>
            <td class="center"><span class="mono">{{ item.requested_by }}</span></td>
            <td class="center"><span class="mono">{{ formatDateTime(item.created_at, locale) }}</span></td>
            <td class="center actions-inline">
              <button
                class="btn"
                type="button"
                :disabled="runs.resolvingApprovalID === item.id"
                @click="resolveApproval(item.id, 'approved')"
              >
                {{ t("pages.runs.approve") }}
              </button>
              <button
                class="btn danger"
                type="button"
                :disabled="runs.resolvingApprovalID === item.id"
                @click="resolveApproval(item.id, 'denied')"
              >
                {{ t("pages.runs.deny") }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="muted">{{ t("states.noPendingApprovals") }}</div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import { formatDateTime, formatDurationSince } from "../shared/lib/datetime";
import { useRunsStore } from "../features/runs/store";

const { t, locale } = useI18n({ useScope: "global" });
const runs = useRunsStore();
const pageSize = 20;
const currentPage = ref(1);

const totalPages = computed(() => Math.max(1, Math.ceil(runs.items.length / pageSize)));
const pageItems = computed(() => {
  const start = (currentPage.value - 1) * pageSize;
  const end = start + pageSize;
  return runs.items.slice(start, end);
});

async function loadAll() {
  await Promise.all([runs.load(), runs.loadRuntimeViews(), runs.loadPendingApprovals()]);
  if (currentPage.value > totalPages.value) {
    currentPage.value = totalPages.value;
  }
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value -= 1;
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value += 1;
  }
}

function runBadgeValue(value: string | null | undefined): string {
  const trimmed = value?.trim();
  if (!trimmed) {
    return "-";
  }
  return trimmed;
}

async function resolveApproval(id: number, decision: "approved" | "denied" | "expired" | "failed") {
  let reason = "";
  if (decision !== "approved") {
    reason = window.prompt(t("pages.runs.reasonPrompt"), "") ?? "";
  }
  await runs.resolvePendingApproval(id, decision, reason);
}

onMounted(() => void loadAll());

watch(
  () => runs.items.length,
  () => {
    if (currentPage.value > totalPages.value) {
      currentPage.value = totalPages.value;
    }
  },
);
</script>

<style scoped>
h2 {
  margin: 0;
  letter-spacing: -0.01em;
}
.pill.s-succeeded {
  background: rgba(5, 150, 105, 0.12);
  border-color: rgba(5, 150, 105, 0.3);
}
.pill.s-failed {
  background: rgba(180, 35, 24, 0.12);
  border-color: rgba(180, 35, 24, 0.3);
}
.pill.s-running {
  background: rgba(37, 99, 235, 0.12);
  border-color: rgba(37, 99, 235, 0.3);
}
.run-badge {
  min-width: 88px;
  text-align: center;
}
.pager {
  margin-top: 12px;
}
.approvals {
  margin-top: 12px;
  border: 1px solid rgba(17, 24, 39, 0.1);
  border-radius: 14px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.6);
}
.runtime-pane {
  margin-top: 12px;
  border: 1px solid rgba(17, 24, 39, 0.1);
  border-radius: 14px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.6);
}
.runtime-filters {
  margin-bottom: 10px;
  gap: 10px;
  flex-wrap: wrap;
}
.runtime-filter {
  display: grid;
  gap: 4px;
  min-width: 180px;
}
h3 {
  margin: 0;
}
.actions-inline {
  display: inline-flex;
  gap: 8px;
}
</style>
