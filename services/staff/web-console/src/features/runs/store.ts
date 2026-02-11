import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { getRun, listRunEvents, listRunLearningFeedback, listRuns } from "./api";
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
        this.items = await listRuns();
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
    run: null as Run | null,
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
        const [run, events, feedback] = await Promise.all([
          getRun(runId),
          listRunEvents(runId),
          listRunLearningFeedback(runId),
        ]);
        this.run = run;
        this.events = events;
        this.feedback = feedback;
      } catch (e) {
        this.run = null;
        this.error = normalizeApiError(e);
      } finally {
        this.loading = false;
      }
    },
  },
});
