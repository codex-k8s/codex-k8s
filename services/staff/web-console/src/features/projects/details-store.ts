import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { getProject } from "./api";
import type { Project } from "./types";

export const useProjectDetailsStore = defineStore("projectDetails", {
  state: () => ({
    projectId: "" as string,
    item: null as Project | null,
    loading: false,
    error: null as ApiError | null,
  }),
  actions: {
    async load(projectId: string): Promise<void> {
      this.projectId = projectId;
      this.loading = true;
      this.error = null;
      try {
        this.item = await getProject(projectId);
      } catch (e) {
        this.item = null;
        this.error = normalizeApiError(e);
      } finally {
        this.loading = false;
      }
    },
  },
});
