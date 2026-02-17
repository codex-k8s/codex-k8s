import { defineStore } from "pinia";

import { deleteCookie, getCookie, setCookie } from "../../shared/lib/cookies";

export type UiEnv = "ai" | "production" | "all";
export type ClusterMode = "view-only" | "dry-run" | "normal";

const cookieKeyEnv = "codexk8s_env";
const cookieKeyNamespace = "codexk8s_namespace";
const cookieKeyProjectId = "codexk8s_project_id";

function readInitialEnv(): UiEnv {
  const v = (getCookie(cookieKeyEnv) || "").toLowerCase();
  if (v === "all") return "all";
  if (v === "prod" || v === "production") return "production";
  return "ai";
}

export const useUiContextStore = defineStore("uiContext", {
  state: () => ({
    env: readInitialEnv() as UiEnv,
    namespace: getCookie(cookieKeyNamespace) || "",
    projectId: getCookie(cookieKeyProjectId) || "",
  }),
  getters: {
    clusterMode: (s): ClusterMode => {
      if (s.env === "all" || s.env === "production") return "view-only";
      if (s.env === "ai") return "dry-run";
      return "dry-run";
    },
  },
  actions: {
    setEnv(next: UiEnv): void {
      this.env = next;
      setCookie(cookieKeyEnv, next, { maxAgeDays: 365, sameSite: "Lax" });
    },
    setNamespace(next: string): void {
      this.namespace = next;
      if (next) setCookie(cookieKeyNamespace, next, { maxAgeDays: 365, sameSite: "Lax" });
      else deleteCookie(cookieKeyNamespace);
    },
    setProjectId(next: string): void {
      this.projectId = next;
      if (next) setCookie(cookieKeyProjectId, next, { maxAgeDays: 365, sameSite: "Lax" });
      else deleteCookie(cookieKeyProjectId);
    },
  },
});
