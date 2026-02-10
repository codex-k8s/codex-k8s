<template>
  <section class="card">
    <div class="row">
      <h2>{{ t("pages.runs.title") }}</h2>
      <button class="btn" type="button" @click="load" :disabled="runs.loading">
        {{ t("common.refresh") }}
      </button>
    </div>

    <div v-if="runs.error" class="err">{{ t(runs.error.messageKey) }}</div>

    <table v-if="runs.items.length" class="tbl">
      <thead>
        <tr>
          <th>{{ t("pages.runs.status") }}</th>
          <th>{{ t("pages.runs.project") }}</th>
          <th>{{ t("pages.runs.correlation") }}</th>
          <th>{{ t("pages.runs.created") }}</th>
          <th>{{ t("pages.runs.started") }}</th>
          <th>{{ t("pages.runs.finished") }}</th>
          <th></th>
          <th>{{ t("pages.runs.id") }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in runs.items" :key="r.id">
          <td>
            <span class="pill" :class="'s-' + r.status">{{ r.status }}</span>
          </td>
          <td class="mono">{{ r.projectId || "-" }}</td>
          <td class="mono">{{ r.correlationId }}</td>
          <td class="mono">{{ r.createdAt }}</td>
          <td class="mono">{{ r.startedAt || "-" }}</td>
          <td class="mono">{{ r.finishedAt || "-" }}</td>
          <td>
            <RouterLink class="lnk" :to="{ name: 'run-details', params: { runId: r.id } }">
              {{ t("pages.runs.details") }}
            </RouterLink>
          </td>
          <td class="mono">{{ r.id }}</td>
        </tr>
      </tbody>
    </table>
    <div v-else class="muted">{{ t("states.noRuns") }}</div>
  </section>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import { useRunsStore } from "../features/runs/store";

const { t } = useI18n({ useScope: "global" });
const runs = useRunsStore();

async function load() {
  await runs.load();
}

onMounted(() => void load());
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
</style>

