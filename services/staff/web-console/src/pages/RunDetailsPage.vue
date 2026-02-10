<template>
  <section class="card">
    <div class="row">
      <div>
        <h2>{{ t("pages.runDetails.title") }}</h2>
        <div class="muted mono">{{ t("pages.runDetails.runId") }}: {{ runId }}</div>
      </div>
      <div class="actions">
        <RouterLink class="btn" :to="{ name: 'runs' }">{{ t("common.back") }}</RouterLink>
        <button class="btn" type="button" @click="loadAll" :disabled="details.loading">{{ t("common.refresh") }}</button>
      </div>
    </div>

    <div v-if="details.error" class="err">{{ t(details.error.messageKey) }}</div>

    <div class="grid">
      <div class="pane">
        <div class="pane-h">{{ t("pages.runDetails.flowEvents") }}</div>
        <div v-if="details.events.length" class="list">
          <div v-for="e in details.events" :key="e.createdAt + ':' + e.eventType" class="item">
            <div class="topline">
              <span class="pill">{{ e.eventType }}</span>
              <span class="mono muted">{{ e.createdAt }}</span>
            </div>
            <pre class="pre">{{ e.payloadJson }}</pre>
          </div>
        </div>
        <div v-else class="muted">{{ t("states.noEvents") }}</div>
      </div>

      <div class="pane">
        <div class="pane-h">{{ t("pages.runDetails.learningFeedback") }}</div>
        <div v-if="details.feedback.length" class="list">
          <div v-for="f in details.feedback" :key="String(f.id)" class="item">
            <div class="topline">
              <span class="pill">{{ f.kind }}</span>
              <span class="mono muted">{{ f.createdAt }}</span>
            </div>
            <pre class="pre">{{ f.explanation }}</pre>
          </div>
        </div>
        <div v-else class="muted">{{ t("states.noLearningFeedback") }}</div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

import { useRunDetailsStore } from "../features/runs/store";

const props = defineProps<{ runId: string }>();

const { t } = useI18n({ useScope: "global" });
const details = useRunDetailsStore();

async function loadAll() {
  await details.load(props.runId);
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

