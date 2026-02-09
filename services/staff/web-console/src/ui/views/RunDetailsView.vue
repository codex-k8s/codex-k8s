<template>
  <section class="card">
    <div class="row">
      <div>
        <h2>Run Details</h2>
        <div class="muted mono">run_id: {{ runId }}</div>
      </div>
      <div class="actions">
        <RouterLink class="btn" to="/runs">Back</RouterLink>
        <button class="btn" @click="loadAll" :disabled="loading">Refresh</button>
      </div>
    </div>

    <div v-if="error" class="err">{{ error }}</div>

    <div class="grid">
      <div class="pane">
        <div class="pane-h">Flow Events</div>
        <div v-if="events.length" class="list">
          <div v-for="e in events" :key="e.created_at + ':' + e.event_type" class="item">
            <div class="topline">
              <span class="pill">{{ e.event_type }}</span>
              <span class="mono muted">{{ e.created_at }}</span>
            </div>
            <pre class="pre">{{ e.payload_json }}</pre>
          </div>
        </div>
        <div v-else class="muted">No events.</div>
      </div>

      <div class="pane">
        <div class="pane-h">Learning Feedback</div>
        <div v-if="feedback.length" class="list">
          <div v-for="f in feedback" :key="String(f.id)" class="item">
            <div class="topline">
              <span class="pill">{{ f.kind }}</span>
              <span class="mono muted">{{ f.created_at }}</span>
            </div>
            <pre class="pre">{{ f.explanation }}</pre>
          </div>
        </div>
        <div v-else class="muted">No learning feedback.</div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { api } from "../api";

type FlowEvent = {
  correlation_id: string;
  event_type: string;
  created_at: string;
  payload_json: string;
};

type Feedback = {
  id: number;
  run_id: string;
  kind: string;
  explanation: string;
  created_at: string;
};

const route = useRoute();
const runId = String(route.params.run_id || "");

const loading = ref(false);
const error = ref<string | null>(null);
const events = ref<FlowEvent[]>([]);
const feedback = ref<Feedback[]>([]);

async function loadAll() {
  loading.value = true;
  error.value = null;
  try {
    const [ev, fb] = await Promise.all([
      api.get(`/api/v1/staff/runs/${runId}/events?limit=500`),
      api.get(`/api/v1/staff/runs/${runId}/learning-feedback?limit=200`),
    ]);
    events.value = (ev.data.items ?? []) as FlowEvent[];
    feedback.value = (fb.data.items ?? []) as Feedback[];
  } catch (e: any) {
    error.value = e?.message ?? "Failed to load";
  } finally {
    loading.value = false;
  }
}

onMounted(() => void loadAll());
</script>

<style scoped>
.card {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(17, 24, 39, 0.1);
  border-radius: 14px;
  padding: 16px;
}
.row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}
.actions {
  display: flex;
  gap: 10px;
  align-items: center;
}
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
.pill {
  display: inline-block;
  font-size: 11px;
  font-weight: 900;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px solid rgba(17, 24, 39, 0.14);
  background: rgba(255, 255, 255, 0.8);
}
.pre {
  margin: 0;
  white-space: pre-wrap;
  font-size: 12px;
  opacity: 0.9;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}
.muted {
  opacity: 0.7;
}
.btn {
  border: 1px solid rgba(17, 24, 39, 0.14);
  background: rgba(255, 255, 255, 0.75);
  padding: 8px 10px;
  border-radius: 10px;
  font-weight: 800;
  cursor: pointer;
  text-decoration: none;
  color: inherit;
}
.err {
  margin-top: 10px;
  color: #b42318;
  font-weight: 700;
}
@media (max-width: 960px) {
  .grid {
    grid-template-columns: 1fr;
  }
}
</style>

