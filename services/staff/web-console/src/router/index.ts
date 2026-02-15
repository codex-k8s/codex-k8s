import { watch } from "vue";
import { createRouter, createWebHistory } from "vue-router";
import type { Pinia } from "pinia";

import { useAuthStore } from "../features/auth/store";
import { i18n } from "../i18n";
import { findNavItemByRouteName } from "../app/navigation";
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

  function updateDocumentTitle(to = router.currentRoute.value) {
    if (typeof document === "undefined") return;

    const t = i18n.global.t;
    const appTitle = t("app.title");

    const meta = to.meta as Record<string, unknown>;
    const crumbKey = typeof meta.crumbKey === "string" ? meta.crumbKey : "";
    if (crumbKey) {
      const pageTitle = t(crumbKey);
      document.title = `${pageTitle} · ${appTitle}`;
      return;
    }

    const routeName = typeof to.name === "string" ? to.name : "";
    const navItem = findNavItemByRouteName(routeName);
    if (navItem) {
      document.title = `${t(navItem.titleKey)} · ${appTitle}`;
      return;
    }

    document.title = appTitle;
  }

  router.afterEach((to) => updateDocumentTitle(to));
  watch(i18n.global.locale, () => updateDocumentTitle(), { immediate: true });

  return router;
}
