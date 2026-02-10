import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { getProject } from "./api";

export type ProjectDetails = {
  id: string;
  slug: string;
  name: string;
};

export const useProjectDetailsStore = defineStore("projectDetails", {
  state: () => ({
    projectId: "" as string,
    item: null as ProjectDetails | null,
    loading: false,
    error: null as ApiError | null,
  }),
  actions: {
    async load(projectId: string): Promise<void> {
      this.projectId = projectId;
      this.loading = true;
      this.error = null;
      try {
        const dto = await getProject(projectId);
        this.item = { id: dto.id, slug: dto.slug, name: dto.name };
      } catch (e) {
        this.item = null;
        this.error = normalizeApiError(e);
      } finally {
        this.loading = false;
      }
    },
  },
});

