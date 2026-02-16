import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { deleteConfigEntry, listConfigEntries, upsertConfigEntry } from "./api";
import type { ConfigEntry } from "./types";

export const useConfigEntriesStore = defineStore("configEntries", {
  state: () => ({
    items: [] as ConfigEntry[],
    loading: false,
    error: null as ApiError | null,
    saving: false,
    saveError: null as ApiError | null,
    deleting: false,
    deleteError: null as ApiError | null,
  }),
  actions: {
    async load(params: { scope: "platform" | "project" | "repository"; projectId?: string; repositoryId?: string; limit?: number }): Promise<void> {
      this.loading = true;
      this.error = null;
      try {
        this.items = await listConfigEntries(params);
      } catch (e) {
        this.error = normalizeApiError(e);
      } finally {
        this.loading = false;
      }
    },

    async upsert(params: {
      scope: "platform" | "project" | "repository";
      kind: "variable" | "secret";
      projectId: string | null;
      repositoryId: string | null;
      key: string;
      valuePlain: string | null;
      valueSecret: string | null;
      syncTargets: string[];
      mutability: "startup_required" | "runtime_mutable";
      isDangerous: boolean;
      dangerousConfirmed: boolean;
    }): Promise<ConfigEntry | null> {
      this.saving = true;
      this.saveError = null;
      try {
        const item = await upsertConfigEntry(params);
        return item;
      } catch (e) {
        this.saveError = normalizeApiError(e);
        return null;
      } finally {
        this.saving = false;
      }
    },

    async remove(id: string): Promise<void> {
      this.deleting = true;
      this.deleteError = null;
      try {
        await deleteConfigEntry(id);
      } catch (e) {
        this.deleteError = normalizeApiError(e);
      } finally {
        this.deleting = false;
      }
    },
  },
});
