import { defineStore } from "pinia";

import { type ApiError, normalizeApiError } from "../../shared/api/errors";
import { fetchMe, logout as apiLogout } from "./api";
import type { UserIdentity } from "./types";

export type AuthStatus = "loading" | "authed" | "anon";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    status: "loading" as AuthStatus,
    me: null as UserIdentity | null,
    error: null as ApiError | null,
    didLoadOnce: false,
  }),
  getters: {
    isAuthed: (s) => s.status === "authed",
    isPlatformAdmin: (s) => Boolean(s.me?.isPlatformAdmin),
  },
  actions: {
    async refresh(): Promise<void> {
      this.status = "loading";
      this.error = null;
      try {
        const dto = await fetchMe();
        this.me = {
          id: dto.user.id,
          email: dto.user.email,
          githubLogin: dto.user.github_login,
          isPlatformAdmin: dto.user.is_platform_admin,
        };
        this.status = "authed";
      } catch (e) {
        this.me = null;
        this.status = "anon";
        this.error = normalizeApiError(e);
      } finally {
        this.didLoadOnce = true;
      }
    },

    async ensureLoaded(): Promise<void> {
      if (this.didLoadOnce) return;
      await this.refresh();
    },

    async logout(): Promise<void> {
      try {
        await apiLogout();
      } finally {
        await this.refresh();
      }
    },
  },
});

