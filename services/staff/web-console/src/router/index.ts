import { createRouter, createWebHistory } from "vue-router";
import type { Pinia } from "pinia";

import { useAuthStore } from "../features/auth/store";
import { routes } from "./routes";

export function createAppRouter(pinia: Pinia) {
  const router = createRouter({
    history: createWebHistory(),
    routes,
  });

  router.beforeEach(async (to) => {
    const auth = useAuthStore(pinia);
    await auth.ensureLoaded();

    if (to.meta.adminOnly && !auth.isPlatformAdmin) {
      return { name: "projects" };
    }
    return true;
  });

  return router;
}

