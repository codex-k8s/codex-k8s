import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { listRunEvents, listRunLearningFeedback, listRuns } from "./api";
import type { FlowEvent, LearningFeedback, Run } from "./types";

export const useRunsStore = defineStore("runs", {
  state: () => ({
    items: [] as Run[],
    loading: false,
    error: null as ApiError | null,
  }),
  actions: {
    async load(): Promise<void> {
      this.loading = true;
      this.error = null;
      try {
        const dtos = await listRuns();
        this.items = dtos.map((r) => ({
          id: r.id,
          correlationId: r.correlation_id,
          projectId: r.project_id ?? null,
          status: r.status,
          createdAt: r.created_at,
          startedAt: r.started_at ?? null,
          finishedAt: r.finished_at ?? null,
        }));
      } catch (e) {
        this.error = normalizeApiError(e);
      } finally {
        this.loading = false;
      }
    },
  },
});

export const useRunDetailsStore = defineStore("runDetails", {
  state: () => ({
    runId: "" as string,
    loading: false,
    error: null as ApiError | null,
    events: [] as FlowEvent[],
    feedback: [] as LearningFeedback[],
  }),
  actions: {
    async load(runId: string): Promise<void> {
      this.runId = runId;
      this.loading = true;
      this.error = null;
      try {
        const [ev, fb] = await Promise.all([listRunEvents(runId), listRunLearningFeedback(runId)]);
        this.events = ev.map((e) => ({
          correlationId: e.correlation_id,
          eventType: e.event_type,
          createdAt: e.created_at,
          payloadJson: e.payload_json,
        }));
        this.feedback = fb.map((f) => ({
          id: f.id,
          runId: f.run_id,
          kind: f.kind,
          explanation: f.explanation,
          createdAt: f.created_at,
        }));
      } catch (e) {
        this.error = normalizeApiError(e);
      } finally {
        this.loading = false;
      }
    },
  },
});

