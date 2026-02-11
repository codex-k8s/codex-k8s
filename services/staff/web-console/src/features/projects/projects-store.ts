import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { deleteProject, listProjects, upsertProject } from "./api";
import type { Project } from "./types";

export const useProjectsStore = defineStore("projects", {
  state: () => ({
    items: [] as Project[],
    loading: false,
    error: null as ApiError | null,
    saving: false,
    saveError: null as ApiError | null,
    deleting: false,
    deleteError: null as ApiError | null,
  }),
  actions: {
    async load(): Promise<void> {
      this.loading = true;
      this.error = null;
      try {
        this.items = await listProjects();
      } catch (e) {
        this.error = normalizeApiError(e);
      } finally {
        this.loading = false;
      }
    },

    async createOrUpdate(slug: string, name: string): Promise<void> {
      this.saving = true;
      this.saveError = null;
      try {
        await upsertProject(slug, name);
        await this.load();
      } catch (e) {
        this.saveError = normalizeApiError(e);
      } finally {
        this.saving = false;
      }
    },

    async remove(projectId: string): Promise<void> {
      this.deleting = true;
      this.deleteError = null;
      try {
        await deleteProject(projectId);
        await this.load();
      } catch (e) {
        this.deleteError = normalizeApiError(e);
      } finally {
        this.deleting = false;
      }
    },
  },
});
