import { defineStore } from "pinia";

import { deleteCookie, getCookie, setCookie } from "../../shared/lib/cookies";

const cookieKeyProjectId = "kodex_project_id";

export const useUiContextStore = defineStore("uiContext", {
  state: () => ({
    projectId: getCookie(cookieKeyProjectId) || "",
  }),
  actions: {
    setProjectId(next: string): void {
      this.projectId = next;
      if (next) setCookie(cookieKeyProjectId, next, { maxAgeDays: 365, sameSite: "Lax" });
      else deleteCookie(cookieKeyProjectId);
    },
  },
});
