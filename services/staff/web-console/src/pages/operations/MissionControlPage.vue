<template>
  <div class="mission-control-page">
    <PageHeader
      :title="t('pages.missionControlPrototype.title')"
      :hint="t('pages.missionControlPrototype.hint')"
    />

    <VAlert v-if="prototype.error" type="error" variant="tonal" class="mt-4">
      {{ t(prototype.error.messageKey) }}
    </VAlert>

    <div v-if="prototype.loading" class="mission-control-page__loading mt-4">
      <VSkeletonLoader type="article, article, article" />
    </div>

    <section v-else class="mission-control-page__shell mt-4">
      <div class="mission-control-page__toolbar">
        <div class="mission-control-page__toolbar-main">
          <VBtnToggle divided mandatory :model-value="activeRouteState.screen" @update:model-value="onSelectScreen">
            <VBtn v-for="option in screenOptions" :key="option.screen" :value="option.screen">
              {{ option.label }}
            </VBtn>
          </VBtnToggle>

          <VSelect
            :model-value="activeRouteState.projectId"
            :items="prototype.projectOptions"
            item-title="title"
            item-value="projectId"
            density="compact"
            variant="outlined"
            hide-details
            label="Проект"
            class="mission-control-page__select"
            @update:model-value="onSelectProject"
          />

          <VSelect
            :model-value="activeRouteState.initiativeId"
            :items="initiativeOptions"
            item-title="title"
            item-value="initiativeId"
            density="compact"
            variant="outlined"
            hide-details
            label="Инициатива"
            class="mission-control-page__select mission-control-page__select--wide"
            @update:model-value="onSelectInitiative"
          />

          <VTextField
            :model-value="activeRouteState.search"
            density="compact"
            variant="outlined"
            hide-details
            prepend-inner-icon="mdi-magnify"
            label="Поиск по инициативам, артефактам и executions"
            class="mission-control-page__search"
            @update:model-value="onUpdateSearch"
          />
        </div>

        <div class="mission-control-page__toolbar-side">
          <VChip v-if="prototype.currentProject" size="small" variant="tonal" color="primary">
            {{ prototype.currentProject.title }}
          </VChip>
          <VBtn variant="text" prepend-icon="mdi-microphone" @click="voiceDialogOpen = true">Голос</VBtn>
        </div>
      </div>

      <MissionControlPrototypeHomeView
        v-if="activeRouteState.screen === 'home'"
        :project-title="prototype.currentProject?.title || ''"
        :project-summary="prototype.currentProject?.summary || ''"
        :attention-cards="prototype.attentionCards"
        :columns="prototype.homeColumns"
        @open-voice="voiceDialogOpen = true"
        @launch-workflow="onOpenStudio"
        @select-initiative="onOpenInitiative"
      />

      <MissionControlPrototypeWorkspaceView
        v-else-if="activeRouteState.screen === 'initiative'"
        :initiative="prototype.currentInitiative"
        :workflow="prototype.currentWorkflow"
        :workspace-view="activeRouteState.workspaceView"
        :stage-views="prototype.workspaceStageViews"
        :artifact-views="prototype.workspaceArtifacts"
        :selected-artifact="selectedArtifactView"
        :activity="prototype.currentInitiativeActivity"
        :flow-nodes="prototype.workspaceFlowNodes"
        :flow-relations="prototype.workspaceFlowRelations"
        @update:view="onUpdateWorkspaceView"
        @select-artifact="onSelectArtifact"
        @open-executions="onOpenExecutions"
        @open-studio="onOpenStudio"
      />

      <MissionControlPrototypeWorkflowStudioView
        v-else-if="activeRouteState.screen === 'studio'"
        :workflow="prototype.currentWorkflow"
        :workflow-options="prototype.workflowOptions"
        :nodes="prototype.studioNodes"
        :relations="prototype.studioRelations"
        @select-workflow="onSelectWorkflow"
      />

      <MissionControlPrototypeExecutionsView
        v-else
        :groups="prototype.executionGroups"
      />
    </section>

    <MissionControlPrototypeVoiceFab @click="voiceDialogOpen = true" />

    <VDialog v-model="voiceDialogOpen" max-width="760">
      <VCard rounded="xl">
        <VCardTitle>{{ t("pages.missionControlPrototype.voice.title") }}</VCardTitle>
        <VCardText class="mission-control-page__voice-sheet">
          <p class="mission-control-page__voice-copy">
            Голосовой запуск станет центральным входом в платформу: отсюда можно создать инициативу, задачу или
            запустить workflow по вашему описанию.
          </p>

          <VTextarea
            v-model="voiceDraft"
            variant="outlined"
            auto-grow
            rows="4"
            hide-details
            label="Распознанная команда"
          />

          <div class="mission-control-page__voice-chips">
            <VChip size="small" variant="tonal" color="primary">Запуск workflow</VChip>
            <VChip size="small" variant="tonal" color="info">Создать задачу</VChip>
            <VChip size="small" variant="tonal" color="warning">Запустить агента</VChip>
          </div>
        </VCardText>
        <VCardActions>
          <VSpacer />
          <VBtn variant="text" @click="voiceDialogOpen = false">Закрыть</VBtn>
          <VBtn color="primary" prepend-icon="mdi-rocket-launch-outline" @click="onOpenStudio">Перейти к workflow</VBtn>
        </VCardActions>
      </VCard>
    </VDialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

import MissionControlPrototypeExecutionsView from "../../features/mission-control/prototype/MissionControlPrototypeExecutionsView.vue";
import MissionControlPrototypeHomeView from "../../features/mission-control/prototype/MissionControlPrototypeHomeView.vue";
import MissionControlPrototypeVoiceFab from "../../features/mission-control/prototype/MissionControlPrototypeVoiceFab.vue";
import MissionControlPrototypeWorkflowStudioView from "../../features/mission-control/prototype/MissionControlPrototypeWorkflowStudioView.vue";
import MissionControlPrototypeWorkspaceView from "../../features/mission-control/prototype/MissionControlPrototypeWorkspaceView.vue";
import {
  buildMissionControlPrototypeRouteQuery,
  missionControlPrototypeRouteStateEquals,
  normalizeMissionControlPrototypeRouteQuery,
  patchMissionControlPrototypeRouteState,
} from "../../features/mission-control/prototype/route";
import { useMissionControlPrototypeStore } from "../../features/mission-control/prototype/store";
import type {
  MissionControlPrototypeRouteState,
  MissionControlScreen,
  MissionInitiativeWorkspaceView,
} from "../../features/mission-control/prototype/types";
import PageHeader from "../../shared/ui/PageHeader.vue";

const route = useRoute();
const router = useRouter();
const prototype = useMissionControlPrototypeStore();
const { t } = useI18n({ useScope: "global" });

const voiceDialogOpen = ref(false);
const voiceDraft = ref("Собери новый workflow для инициативы Mission Control: сначала owner narrative, потом дизайн, затем фронтенд-прототип и follow-up задачу на backend.");
const activeRouteState = ref<MissionControlPrototypeRouteState>({
  screen: "home",
  projectId: "",
  initiativeId: "",
  workflowId: "",
  artifactId: "",
  search: "",
  workspaceView: "overview",
});

const routeState = computed(() => normalizeMissionControlPrototypeRouteQuery(route.query));
const screenOptions = computed(() => [
  { screen: "home" as const, label: t("pages.missionControlPrototype.screens.home") },
  { screen: "initiative" as const, label: t("pages.missionControlPrototype.screens.initiative") },
  { screen: "studio" as const, label: t("pages.missionControlPrototype.screens.studio") },
  { screen: "executions" as const, label: t("pages.missionControlPrototype.screens.executions") },
]);
const initiativeOptions = computed(() =>
  prototype.projectInitiatives.map((initiative) => ({
    initiativeId: initiative.initiativeId,
    title: initiative.title,
  })),
);
const selectedArtifactView = computed(
  () => prototype.workspaceArtifacts.find((artifact) => artifact.artifactId === activeRouteState.value.artifactId) ?? null,
);

watch(
  routeState,
  async (nextState) => {
    const normalizedState = await prototype.syncRouteState(nextState);
    activeRouteState.value = normalizedState;

    if (!missionControlPrototypeRouteStateEquals(nextState, normalizedState)) {
      await replaceRoute(normalizedState);
    }
  },
  { immediate: true, deep: true },
);

async function replaceRoute(nextState: MissionControlPrototypeRouteState): Promise<void> {
  await router.replace({
    name: "mission-control",
    query: buildMissionControlPrototypeRouteQuery(nextState, {
      projectId: prototype.defaultProjectId || nextState.projectId,
      initiativeId: prototype.defaultInitiativeId || nextState.initiativeId,
      workflowId: prototype.defaultWorkflowId || nextState.workflowId,
    }),
  });
}

function updateRoute(patch: Partial<MissionControlPrototypeRouteState>): void {
  const nextState = patchMissionControlPrototypeRouteState(activeRouteState.value, patch);
  if (missionControlPrototypeRouteStateEquals(activeRouteState.value, nextState)) {
    return;
  }
  void replaceRoute(nextState);
}

function onSelectScreen(nextScreen: string): void {
  if (nextScreen === "home" || nextScreen === "initiative" || nextScreen === "studio" || nextScreen === "executions") {
    updateRoute({
      screen: nextScreen,
      artifactId: nextScreen === "executions" ? activeRouteState.value.artifactId : "",
    });
  }
}

function onSelectProject(nextProjectId: string | null): void {
  if (!nextProjectId) {
    return;
  }
  updateRoute({
    projectId: nextProjectId,
    initiativeId: "",
    workflowId: "",
    artifactId: "",
    screen: activeRouteState.value.screen === "studio" ? "studio" : "home",
  });
}

function onSelectInitiative(nextInitiativeId: string | null): void {
  if (!nextInitiativeId) {
    return;
  }
  updateRoute({
    initiativeId: nextInitiativeId,
    workflowId: "",
    artifactId: "",
    screen: "initiative",
  });
}

function onOpenInitiative(nextInitiativeId: string): void {
  updateRoute({
    initiativeId: nextInitiativeId,
    workflowId: "",
    screen: "initiative",
    workspaceView: "overview",
  });
}

function onSelectWorkflow(nextWorkflowId: string): void {
  updateRoute({
    workflowId: nextWorkflowId,
    screen: "studio",
  });
}

function onOpenStudio(): void {
  voiceDialogOpen.value = false;
  updateRoute({
    screen: "studio",
  });
}

function onOpenExecutions(): void {
  updateRoute({
    screen: "executions",
  });
}

function onUpdateSearch(nextSearch: string | null): void {
  updateRoute({
    search: nextSearch || "",
  });
}

function onUpdateWorkspaceView(nextView: MissionInitiativeWorkspaceView): void {
  updateRoute({
    workspaceView: nextView,
  });
}

function onSelectArtifact(nextArtifactId: string): void {
  updateRoute({
    artifactId: nextArtifactId,
  });
}
</script>

<style scoped>
.mission-control-page {
  display: grid;
}

.mission-control-page__shell {
  display: grid;
  gap: 18px;
  padding: 20px;
  border-radius: 32px;
  background:
    radial-gradient(circle at top left, rgba(255, 232, 186, 0.18), transparent 28%),
    radial-gradient(circle at top right, rgba(201, 229, 255, 0.2), transparent 26%),
    linear-gradient(180deg, rgba(249, 248, 245, 0.96), rgba(245, 246, 249, 0.96));
  border: 1px solid rgba(223, 228, 235, 0.94);
}

.mission-control-page__toolbar {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  flex-wrap: wrap;
}

.mission-control-page__toolbar-main {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
  flex: 1;
}

.mission-control-page__toolbar-side {
  display: flex;
  gap: 10px;
  align-items: center;
}

.mission-control-page__select {
  width: 220px;
}

.mission-control-page__select--wide {
  width: 340px;
}

.mission-control-page__search {
  min-width: 320px;
  max-width: 480px;
}

.mission-control-page__loading {
  padding: 18px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.88);
}

.mission-control-page__voice-sheet {
  display: grid;
  gap: 14px;
}

.mission-control-page__voice-copy {
  margin: 0;
  font-size: 0.95rem;
  line-height: 1.55;
  color: rgb(89, 98, 112);
}

.mission-control-page__voice-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

@media (max-width: 900px) {
  .mission-control-page__shell {
    padding: 14px;
  }

  .mission-control-page__select,
  .mission-control-page__select--wide,
  .mission-control-page__search {
    width: 100%;
    min-width: 0;
    max-width: none;
  }
}
</style>
