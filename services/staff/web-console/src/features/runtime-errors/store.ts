import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { listRuntimeErrors, markRuntimeErrorViewed } from "./api";
import type { RuntimeError } from "./types";

const defaultStackLimit = 5;

function sortNewest(items: RuntimeError[]): RuntimeError[] {
  const out = [...items];
  out.sort((a, b) => String(b.created_at || "").localeCompare(String(a.created_at || "")));
  return out;
}

export const useRuntimeErrorsStore = defineStore("runtimeErrors", {
  state: () => ({
    items: [] as RuntimeError[],
    loading: false,
    error: null as ApiError | null,
  }),
  actions: {
    async loadActive(limit = defaultStackLimit): Promise<void> {
      if (this.loading) {
        return;
      }
      this.loading = true;
      this.error = null;
      try {
        const loaded = await listRuntimeErrors({ state: "active" }, limit);
        this.items = sortNewest(loaded).slice(0, Math.max(1, limit));
      } catch (err) {
        this.error = normalizeApiError(err);
      } finally {
        this.loading = false;
      }
    },
    async dismiss(id: string): Promise<void> {
      const normalizedID = String(id || "").trim();
      if (!normalizedID) {
        return;
      }
      try {
        await markRuntimeErrorViewed(normalizedID);
        this.items = this.items.filter((item) => item.id !== normalizedID);
      } catch (err) {
        this.error = normalizeApiError(err);
      }
    },
    clear(): void {
      this.items = [];
      this.error = null;
      this.loading = false;
    },
  },
});

