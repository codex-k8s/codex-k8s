import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { createAllowedUser, listUsers } from "./api";
import type { User } from "./types";

export const useUsersStore = defineStore("users", {
  state: () => ({
    items: [] as User[],
    loading: false,
    error: null as ApiError | null,
    creating: false,
    createError: null as ApiError | null,
  }),
  actions: {
    async load(): Promise<void> {
      this.loading = true;
      this.error = null;
      try {
        const dtos = await listUsers();
        this.items = dtos.map((u) => ({
          id: u.id,
          email: u.email,
          githubLogin: u.github_login ?? null,
          isPlatformAdmin: u.is_platform_admin,
        }));
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
  },
});

