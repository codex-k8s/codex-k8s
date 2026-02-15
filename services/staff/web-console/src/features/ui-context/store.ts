import { defineStore } from "pinia";

import { getCookie, setCookie } from "../../shared/lib/cookies";

export type UiEnv = "ai" | "ai-staging" | "prod";
export type ClusterMode = "view-only" | "dry-run" | "normal";

const cookieKeyEnv = "codexk8s_env";
const cookieKeyNamespace = "codexk8s_namespace";
const cookieKeyProjectId = "codexk8s_project_id";

function readInitialEnv(): UiEnv {
  const v = (getCookie(cookieKeyEnv) || "").toLowerCase();
  if (v === "prod") return "prod";
  if (v === "ai-staging") return "ai-staging";
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
      if (s.env === "ai-staging" || s.env === "prod") return "view-only";
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
      if (next) {
        setCookie(cookieKeyNamespace, next, { maxAgeDays: 365, sameSite: "Lax" });
      }
    },
    setProjectId(next: string): void {
      this.projectId = next;
      if (next) {
        setCookie(cookieKeyProjectId, next, { maxAgeDays: 365, sameSite: "Lax" });
      }
    },
  },
});

