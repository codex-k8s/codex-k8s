import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { createAllowedUser, deleteUser, listUsers } from "./api";
import type { User } from "./types";

export const useUsersStore = defineStore("users", {
  state: () => ({
    items: [] as User[],
    loading: false,
    error: null as ApiError | null,
    creating: false,
    createError: null as ApiError | null,
    deleting: false,
    deleteError: null as ApiError | null,
  }),
  actions: {
    async load(): Promise<void> {
      this.loading = true;
      this.error = null;
      try {
        this.items = await listUsers();
      } catch (e) {
        this.error = normalizeApiError(e);
      } finally {
        this.loading = false;
      }
    },

    async create(email: string, isPlatformAdmin: boolean): Promise<void> {
      this.creating = true;
      this.createError = null;
      try {
        await createAllowedUser(email, isPlatformAdmin);
        await this.load();
      } catch (e) {
        this.createError = normalizeApiError(e);
      } finally {
        this.creating = false;
      }
    },

    async remove(userId: string): Promise<void> {
      this.deleting = true;
      this.deleteError = null;
      try {
        await deleteUser(userId);
        await this.load();
      } catch (e) {
        this.deleteError = normalizeApiError(e);
      } finally {
        this.deleting = false;
      }
    },
  },
});
