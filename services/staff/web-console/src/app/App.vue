<template>
  <VApp>
    <VAppBar border :elevation="0" density="comfortable">
      <VAppBarNavIcon v-if="auth.isAuthed" @click="toggleDrawer" />

      <RouterLink class="brand" :to="{ name: 'projects' }">
        <img class="logo" src="/brand/logo.png" alt="codex-k8s logo" />
        <div class="brand-text">
          <div class="brand-title">{{ t("app.title") }}</div>
          <div class="brand-subtitle">{{ t("app.subtitle") }}</div>
        </div>
      </RouterLink>

      <VSpacer />

      <VBreadcrumbs v-if="auth.isAuthed" class="breadcrumbs" :items="breadcrumbs" density="compact" />

      <VSpacer />

      <VMenu v-if="auth.isAuthed" :close-on-content-click="false">
        <template #activator="{ props: menuProps }">
          <VBtn v-bind="menuProps" variant="tonal" prepend-icon="mdi-database-outline">
            {{ contextLabel }}
          </VBtn>
        </template>
        <VCard min-width="420">
          <VCardTitle class="text-subtitle-2">{{ t("context.title") }}</VCardTitle>
          <VCardText>
            <VRow density="compact">
              <VCol cols="12" md="6">
                <VSelect
                  v-model="projectIdModel"
                  :items="projectOptions"
                  :label="t('context.project')"
                  hide-details
                />
              </VCol>
              <VCol cols="12" md="6">
                <VSelect v-model="envModel" :items="envOptions" :label="t('context.env')" hide-details />
              </VCol>
              <VCol cols="12">
                <VSelect v-model="namespaceModel" :items="namespaceOptions" :label="t('context.namespace')" hide-details />
              </VCol>
            </VRow>
          </VCardText>
        </VCard>
      </VMenu>

      <VMenu v-if="auth.isAuthed">
        <template #activator="{ props: menuProps }">
          <VBtn v-bind="menuProps" icon="mdi-bell-outline" variant="text" />
        </template>
        <VList density="compact" min-width="320">
          <VListItem :title="t('notifications.title')" disabled />
          <VDivider />
          <VListItem
            v-for="n in notifications"
            :key="n.id"
            :title="n.title"
            :subtitle="n.subtitle"
            prepend-icon="mdi-information-outline"
          />
        </VList>
      </VMenu>

      <VBtnToggle v-model="locale" class="ml-2" divided density="compact" mandatory>
        <VBtn value="en">{{ t("i18n.enFlag") }}</VBtn>
        <VBtn value="ru">{{ t("i18n.ruFlag") }}</VBtn>
      </VBtnToggle>

      <template v-if="auth.status === 'authed'">
        <VMenu>
          <template #activator="{ props: menuProps }">
            <VBtn v-bind="menuProps" class="ml-2" variant="tonal" prepend-icon="mdi-account-circle-outline">
              {{ auth.me?.email || t("common.loading") }}
            </VBtn>
          </template>
          <VList density="compact" min-width="280">
            <VListItem :title="auth.me?.email || '-'" :subtitle="auth.me?.githubLogin ? '@' + auth.me.githubLogin : ''" />
            <VListItem v-if="auth.me?.isPlatformAdmin" :title="t('roles.admin')" prepend-icon="mdi-shield-account-outline" />
            <VDivider />
            <VListItem :title="t('common.logout')" prepend-icon="mdi-logout" @click="logout" />
          </VList>
        </VMenu>
      </template>

      <template v-else-if="auth.status === 'anon'">
        <VBtn class="ml-2" variant="tonal" href="/oauth2/start" prepend-icon="mdi-github">
          {{ t("common.loginWithGitHub") }}
        </VBtn>
      </template>
    </VAppBar>

    <VNavigationDrawer
      v-if="auth.isAuthed"
      v-model="drawerOpen"
      :rail="drawerRail && !isMobile"
      :temporary="isMobile"
      :permanent="!isMobile"
      width="320"
    >
      <VList nav density="compact">
        <template v-for="g in navGroups" :key="g.id">
          <VListSubheader class="mt-2">
            {{ t(g.titleKey) }}
          </VListSubheader>

          <VListItem
            v-for="item in groupedItems[g.id]"
            :key="item.routeName"
            :prepend-icon="item.icon"
            :to="navTo(item)"
            :disabled="isNavDisabled(item)"
            rounded="lg"
          >
            <VListItemTitle>{{ t(item.titleKey) }}</VListItemTitle>
            <template #append>
              <VChip v-if="item.comingSoon" size="x-small" variant="tonal" color="warning" class="font-weight-bold">
                {{ t("common.comingSoon") }}
              </VChip>
            </template>
          </VListItem>
        </template>
      </VList>

      <template #append>
        <div class="pa-2 d-flex justify-end">
          <VBtn
            v-if="!isMobile"
            variant="text"
            :icon="drawerRail ? 'mdi-arrow-expand-right' : 'mdi-arrow-collapse-left'"
            @click="drawerRail = !drawerRail"
          />
        </div>
      </template>
    </VNavigationDrawer>

    <VMain class="main-bg">
      <VContainer class="content" fluid>
        <VProgressLinear v-if="auth.status === 'loading'" indeterminate color="primary" />

        <VCard v-else-if="auth.status === 'anon'" class="mx-auto mt-10" max-width="720" variant="outlined">
          <VCardTitle class="text-h6">{{ t("states.accessRequiredTitle") }}</VCardTitle>
          <VCardText class="text-body-2 text-medium-emphasis">
            {{ t("states.accessRequiredText") }}
          </VCardText>
          <VCardActions>
            <VBtn variant="tonal" href="/oauth2/start" prepend-icon="mdi-github">
              {{ t("common.loginWithGitHub") }}
            </VBtn>
          </VCardActions>
        </VCard>

        <RouterView v-else />
      </VContainer>
    </VMain>

    <SnackbarHost />
  </VApp>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { useDisplay } from "vuetify";

import { persistLocale, type Locale } from "../i18n/locale";
import { useAuthStore } from "../features/auth/store";
import { useProjectsStore } from "../features/projects/projects-store";
import { useUiContextStore } from "../features/ui-context/store";
import { navGroups, navItems, findNavItemByRouteName, type NavItem } from "./navigation";
import SnackbarHost from "../shared/ui/feedback/SnackbarHost.vue";

const auth = useAuthStore();
const projects = useProjectsStore();
const uiContext = useUiContextStore();
const { t, locale } = useI18n({ useScope: "global" });
const route = useRoute();
const display = useDisplay();

const drawerOpen = ref(true);
const drawerRail = ref(false);
const isMobile = computed(() => display.mobile.value);

const envModel = computed({
  get: () => uiContext.env,
  set: (v) => uiContext.setEnv(v),
});
const namespaceModel = computed({
  get: () => uiContext.namespace,
  set: (v) => uiContext.setNamespace(v),
});
const projectIdModel = computed({
  get: () => uiContext.projectId,
  set: (v) => uiContext.setProjectId(v),
});

watch(
  locale,
  (next) => {
    persistLocale(next as Locale);
  },
  { immediate: true },
);

async function logout() {
  await auth.logout();
}

onMounted(() => {
  void auth.ensureLoaded().then(() => {
    if (auth.status === "authed") {
      void projects.load();
    }
  });
});

watch(
  () => route.params.projectId,
  (v) => {
    if (typeof v === "string" && v) {
      uiContext.setProjectId(v);
    }
  },
  { immediate: true },
);

const projectOptions = computed(() =>
  projects.items.map((p) => ({
    title: p.name || p.slug || p.id,
    value: p.id,
  })),
);

watch(
  () => projects.items,
  (items) => {
    if (uiContext.projectId) return;
    if (items.length) {
      uiContext.setProjectId(items[0].id);
    }
  },
  { immediate: true },
);

const envOptions = [
  { title: "ai", value: "ai" },
  { title: "ai-staging", value: "ai-staging" },
  { title: "prod", value: "prod" },
] as const;

const namespaceOptions = computed(() => {
  const project = "codex-k8s";
  if (uiContext.env === "ai") return [`${project}-dev-1`, `${project}-dev-2`, `${project}-dev-3`];
  if (uiContext.env === "ai-staging") return [`${project}-ai-staging`];
  return [`${project}-prod`];
});

watch(
  () => namespaceOptions.value,
  (items) => {
    if (!items.length) return;
    if (items.includes(uiContext.namespace)) return;
    uiContext.setNamespace(items[0]);
  },
  { immediate: true },
);

const contextLabel = computed(() => {
  const project = projectOptions.value.find((p) => p.value === uiContext.projectId)?.title || "-";
  return `${project} / ${uiContext.env} / ${uiContext.namespace || "-"}`;
});

const notifications = [
  { id: "n1", title: t("notifications.items.sample1.title"), subtitle: t("notifications.items.sample1.subtitle") },
  { id: "n2", title: t("notifications.items.sample2.title"), subtitle: t("notifications.items.sample2.subtitle") },
] as const;

const groupedItems = computed(() => {
  const byGroup: Record<string, NavItem[]> = Object.fromEntries(navGroups.map((g) => [g.id, []]));

  for (const item of navItems) {
    if (item.adminOnly && !auth.isPlatformAdmin) continue;
    byGroup[item.groupId].push(item);
  }
  return byGroup as Record<(typeof navGroups)[number]["id"], NavItem[]>;
});

function navTo(item: NavItem) {
  if (!item.requiresProject) return { name: item.routeName };
  if (!uiContext.projectId) return { name: "projects" };

  return {
    name: item.routeName,
    params: { projectId: uiContext.projectId },
  };
}

function isNavDisabled(item: NavItem): boolean {
  if (item.requiresProject && !uiContext.projectId) return true;
  return false;
}

function toggleDrawer(): void {
  if (isMobile.value) {
    drawerOpen.value = !drawerOpen.value;
    return;
  }
  drawerRail.value = !drawerRail.value;
}

const breadcrumbs = computed(() => {
  const rName = typeof route.name === "string" ? route.name : "";
  const meta = route.meta as Record<string, unknown>;

  const items: { title: string }[] = [];

  // For detail pages, keep breadcrumb root pointing to list pages.
  const baseRouteName = rName === "run-details" ? "runs" : rName;
  const navItem = findNavItemByRouteName(baseRouteName);
  if (navItem) {
    const g = navGroups.find((x) => x.id === navItem.groupId);
    if (g) items.push({ title: t(g.titleKey) });
    items.push({ title: t(navItem.titleKey) });
  }

  if (rName === "run-details" && typeof meta.crumbKey === "string" && meta.crumbKey) {
    items.push({ title: t(meta.crumbKey) });
  }

  return items;
});
</script>

<style scoped>
.brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: inherit;
  text-decoration: none;
}
.logo {
  width: 30px;
  height: 30px;
  border-radius: 8px;
}
.brand-text {
  line-height: 1.1;
}
.brand-title {
  font-weight: 900;
  letter-spacing: -0.01em;
}
.brand-subtitle {
  font-size: 12px;
  opacity: 0.75;
}
.breadcrumbs {
  max-width: 520px;
  overflow: hidden;
}
.content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 16px;
}
.main-bg {
  background: radial-gradient(1200px 600px at 15% 0%, rgba(255, 246, 229, 0.8), transparent 55%),
    radial-gradient(900px 450px at 95% 10%, rgba(232, 246, 255, 0.9), transparent 55%),
    linear-gradient(180deg, #fbfbfc, #f4f5f7);
}
</style>
