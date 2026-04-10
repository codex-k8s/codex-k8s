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
        <VBtnToggle divided mandatory :model-value="activeRouteState.screen" @update:model-value="onSelectScreen">
          <VBtn v-for="option in screenOptions" :key="option.screen" :value="option.screen">
            {{ option.label }}
          </VBtn>
        </VBtnToggle>

        <VMenu v-model="createMenuOpen" location="bottom start">
          <template #activator="{ props: menuProps }">
            <VBtn
              color="primary"
              prepend-icon="mdi-plus-circle-outline"
              variant="tonal"
              v-bind="menuProps"
            >
              Создать
            </VBtn>
          </template>

          <VList min-width="320">
            <VListItem
              v-for="item in createMenuItems"
              :key="item.mode"
              :prepend-icon="item.icon"
              :title="item.title"
              :subtitle="item.subtitle"
              @click="onOpenCreateDialog(item.mode)"
            />
          </VList>
        </VMenu>

        <VBtn
          variant="outlined"
          prepend-icon="mdi-briefcase-search-outline"
          class="mission-control-page__initiative-button"
          @click="initiativePickerOpen = true"
        >
          {{ initiativePickerLabel }}
        </VBtn>

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

      <MissionControlPrototypeHomeView
        v-if="activeRouteState.screen === 'home'"
        :project-title="prototype.currentProject?.title || ''"
        :project-summary="prototype.currentProject?.summary || ''"
        :attention-cards="prototype.attentionCards"
        :columns="prototype.homeColumns"
        :selected-initiative-title="homeSelectedInitiativeTitle"
        :selected-filter-label="homeFilterLabel"
        @open-attention="onOpenAttention"
        @select-initiative="onFocusInitiative"
        @open-workspace="onOpenInitiative"
        @clear-initiative="onClearInitiative"
        @clear-filter="onClearHomeFilter"
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
        @update:view="onUpdateWorkspaceView"
        @select-artifact="onSelectArtifact"
        @open-executions="onOpenExecutions"
        @open-studio="onOpenStudio"
      />

      <MissionControlPrototypeWorkflowStudioView
        v-else-if="activeRouteState.screen === 'studio'"
        :workflow="prototype.currentWorkflow"
        :workflow-options="prototype.workflowOptions"
        @select-workflow="onSelectWorkflow"
      />

      <MissionControlPrototypeExecutionsView
        v-else
        :groups="prototype.executionGroups"
      />
    </section>

    <MissionControlPrototypeVoiceFab @click="voiceDialogOpen = true" />

    <VDialog v-model="initiativePickerOpen" max-width="760">
      <VCard rounded="xl">
        <VCardTitle>Выбор инициативы</VCardTitle>
        <VCardText class="mission-control-page__initiative-sheet">
          <VTextField
            v-model="initiativePickerSearch"
            density="compact"
            variant="outlined"
            hide-details
            prepend-inner-icon="mdi-magnify"
            label="Поиск по инициативам"
          />

          <div v-if="activeRouteState.screen === 'home'" class="mission-control-page__initiative-actions">
            <VBtn variant="text" prepend-icon="mdi-format-list-bulleted" @click="onPickAllInitiatives">
              Все инициативы проекта
            </VBtn>
          </div>

          <div class="mission-control-page__initiative-list">
            <button
              v-for="initiative in filteredInitiatives"
              :key="initiative.initiativeId"
              type="button"
              class="mission-control-page__initiative-option"
              @click="onPickInitiative(initiative.initiativeId)"
            >
              <div class="mission-control-page__initiative-option-title">{{ initiative.title }}</div>
              <div class="mission-control-page__initiative-option-summary">{{ initiative.summary }}</div>
            </button>
          </div>
        </VCardText>
        <VCardActions>
          <VSpacer />
          <VBtn variant="text" @click="initiativePickerOpen = false">Закрыть</VBtn>
        </VCardActions>
      </VCard>
    </VDialog>

    <VDialog v-model="voiceDialogOpen" max-width="760">
      <VCard rounded="xl">
        <VCardTitle>{{ t("pages.missionControlPrototype.voice.title") }}</VCardTitle>
        <VCardText class="mission-control-page__voice-sheet">
          <p class="mission-control-page__voice-copy">
            Голосовой запуск остается главным способом поставить работу. Агент затем создает нужные GitHub Issue и PR через
            <code>gh</code>, а машина приемки проверяет, что в body проставлены watermark-блок и обязательные поля workflow.
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
          <VBtn color="primary" prepend-icon="mdi-rocket-launch-outline" @click="onOpenStudio">
            Перейти к редактору workflow
          </VBtn>
        </VCardActions>
      </VCard>
    </VDialog>

    <VDialog v-model="createDialogOpen" max-width="760">
      <VCard rounded="xl">
        <VCardTitle>{{ createDialogTitle }}</VCardTitle>
        <VCardText class="mission-control-page__create-sheet">
          <p class="mission-control-page__voice-copy">{{ createDialogSummary }}</p>

          <VTextField
            v-model="createTitle"
            density="compact"
            variant="outlined"
            hide-details
            label="Короткое название"
          />

          <VSelect
            v-if="createMode === 'workflow' || createMode === 'hotfix'"
            v-model="createWorkflowId"
            :items="prototype.workflowOptions"
            item-title="title"
            item-value="workflowId"
            density="compact"
            variant="outlined"
            hide-details
            label="Шаблон workflow"
          />

          <VTextarea
            v-model="createDraft"
            variant="outlined"
            auto-grow
            rows="4"
            hide-details
            label="Что должен сделать агент"
          />

          <div class="mission-control-page__watermark-note">
            <strong>Что будет дальше</strong>
            <ul>
              <li>Агент создаст нужные GitHub Issue и PR через <code>gh</code>.</li>
              <li>Workflow добавит в prompt, какие этапы пройти и какой follow-up артефакт должен появиться дальше.</li>
              <li>При приемке машина проверит watermark-блок и обязательные поля в body, и при нехватке данных возобновит сессию.</li>
            </ul>
          </div>
        </VCardText>
        <VCardActions>
          <VBtn variant="text" prepend-icon="mdi-microphone" @click="onCreateByVoice">Продиктовать</VBtn>
          <VSpacer />
          <VBtn variant="text" @click="createDialogOpen = false">Закрыть</VBtn>
          <VBtn color="primary" prepend-icon="mdi-shape-outline" @click="onSubmitCreateDialog">
            {{ createDialogActionLabel }}
          </VBtn>
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
  MissionInitiativeWorkspaceView,
} from "../../features/mission-control/prototype/types";
import { useUiContextStore } from "../../features/ui-context/store";
import PageHeader from "../../shared/ui/PageHeader.vue";

const route = useRoute();
const router = useRouter();
const prototype = useMissionControlPrototypeStore();
const uiContext = useUiContextStore();
const { t } = useI18n({ useScope: "global" });

const initiativePickerOpen = ref(false);
const initiativePickerSearch = ref("");
const voiceDialogOpen = ref(false);
const createMenuOpen = ref(false);
const createDialogOpen = ref(false);
const voiceDraft = ref(
  "Собери новый workflow для инициативы Mission Control: сначала owner narrative, потом дизайн, затем фронтенд-прототип и follow-up задачу на backend.",
);
const createMode = ref<"issue" | "workflow" | "hotfix" | "follow-up">("issue");
const createTitle = ref("");
const createDraft = ref("");
const createWorkflowId = ref("");
const activeRouteState = ref<MissionControlPrototypeRouteState>({
  screen: "home",
  projectId: "",
  initiativeId: "",
  workflowId: "",
  artifactId: "",
  search: "",
  homeFilter: "all",
  workspaceView: "overview",
});

const routeState = computed(() => normalizeMissionControlPrototypeRouteQuery(route.query));
const createMenuItems = [
  {
    mode: "issue" as const,
    icon: "mdi-clipboard-plus-outline",
    title: "Новая задача",
    subtitle: "Обычный GitHub Issue без полного workflow.",
  },
  {
    mode: "workflow" as const,
    icon: "mdi-rocket-launch-outline",
    title: "Запустить workflow",
    subtitle: "Подготовить инициативу по одному из шаблонов.",
  },
  {
    mode: "hotfix" as const,
    icon: "mdi-flash-outline",
    title: "Горячее исправление",
    subtitle: "Быстрый путь triage → fix → qa → release → postdeploy.",
  },
  {
    mode: "follow-up" as const,
    icon: "mdi-source-branch-plus",
    title: "Follow-up задача",
    subtitle: "Создать следующий этап после текущего результата.",
  },
];
const screenOptions = computed(() => [
  { screen: "home" as const, label: t("pages.missionControlPrototype.screens.home") },
  { screen: "initiative" as const, label: t("pages.missionControlPrototype.screens.initiative") },
  { screen: "studio" as const, label: t("pages.missionControlPrototype.screens.studio") },
  { screen: "executions" as const, label: t("pages.missionControlPrototype.screens.executions") },
]);
const filteredInitiatives = computed(() => {
  const needle = initiativePickerSearch.value.trim().toLowerCase();
  if (needle === "") {
    return prototype.projectInitiatives;
  }

  return prototype.projectInitiatives.filter((initiative) =>
    [initiative.title, initiative.summary, ...initiative.tags].some((part) => part.toLowerCase().includes(needle)),
  );
});
const initiativePickerLabel = computed(() => {
  if (activeRouteState.value.screen === "home" && activeRouteState.value.initiativeId === "") {
    return "Все инициативы проекта";
  }
  return prototype.currentInitiative?.title || "Выбрать инициативу";
});
const homeSelectedInitiativeTitle = computed(() =>
  activeRouteState.value.screen === "home" ? prototype.currentInitiative?.title || "" : "",
);
const homeFilterLabel = computed(() => {
  switch (activeRouteState.value.homeFilter) {
    case "needs-decision":
      return "Нуждаются в решении";
    case "blocked":
      return "Есть блокеры";
    case "release-ready":
      return "Почти готовы к выпуску";
    case "all":
      return "";
  }
});
const selectedArtifactView = computed(
  () => prototype.workspaceArtifacts.find((artifact) => artifact.artifactId === activeRouteState.value.artifactId) ?? null,
);
const createDialogTitle = computed(() => {
  switch (createMode.value) {
    case "issue":
      return "Новая задача";
    case "workflow":
      return "Запуск workflow";
    case "hotfix":
      return "Горячее исправление";
    case "follow-up":
      return "Follow-up задача";
  }
});
const createDialogSummary = computed(() => {
  switch (createMode.value) {
    case "issue":
      return "Используйте этот сценарий, когда нужен одиночный GitHub Issue без отдельного шаблона процесса.";
    case "workflow":
      return "Workflow не создает артефакты сам по себе: он задает prompt, по которому агент через gh откроет нужные Issue и PR.";
    case "hotfix":
      return "Hotfix задает жесткий путь до релиза и follow-up по коренной причине. Все служебные поля будут ожидаться в body.";
    case "follow-up":
      return "Follow-up нужен, когда после текущего PR или issue должен появиться следующий этап: review, qa, release или backend handover.";
  }
});
const createDialogActionLabel = computed(() =>
  createMode.value === "workflow" || createMode.value === "hotfix" ? "Открыть шаблон" : "Подготовить голосовую команду",
);

watch(
  [routeState, () => uiContext.projectId],
  async ([nextState, selectedProjectId]) => {
    const requestedState = {
      ...nextState,
      projectId: typeof selectedProjectId === "string" && selectedProjectId !== "" ? selectedProjectId : nextState.projectId,
    };
    const normalizedState = await prototype.syncRouteState(requestedState);
    activeRouteState.value = normalizedState;

    if (!missionControlPrototypeRouteStateEquals(requestedState, normalizedState)) {
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

function onFocusInitiative(nextInitiativeId: string): void {
  updateRoute({
    initiativeId: nextInitiativeId,
    artifactId: "",
    screen: "home",
  });
}

function onClearInitiative(): void {
  updateRoute({
    initiativeId: "",
    artifactId: "",
    screen: "home",
  });
}

function onClearHomeFilter(): void {
  updateRoute({
    homeFilter: "all",
    screen: "home",
  });
}

function onPickInitiative(nextInitiativeId: string): void {
  initiativePickerOpen.value = false;
  initiativePickerSearch.value = "";

  if (activeRouteState.value.screen === "home") {
    onFocusInitiative(nextInitiativeId);
    return;
  }

  onSelectInitiative(nextInitiativeId);
}

function onPickAllInitiatives(): void {
  initiativePickerOpen.value = false;
  initiativePickerSearch.value = "";
  onClearInitiative();
}

function onSelectWorkflow(nextWorkflowId: string): void {
  updateRoute({
    workflowId: nextWorkflowId,
    screen: "studio",
  });
}

function onOpenStudio(): void {
  voiceDialogOpen.value = false;
  createDialogOpen.value = false;
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

function onOpenAttention(cardId: string): void {
  if (cardId === "active-runs") {
    updateRoute({
      screen: "executions",
    });
    return;
  }

  const filterMap = {
    "needs-decision": "needs-decision",
    blocked: "blocked",
    "release-ready": "release-ready",
  } as const;
  const nextFilter = filterMap[cardId as keyof typeof filterMap];
  if (!nextFilter) {
    return;
  }

  updateRoute({
    screen: "home",
    homeFilter: nextFilter,
  });
}

function onOpenCreateDialog(mode: "issue" | "workflow" | "hotfix" | "follow-up"): void {
  createMenuOpen.value = false;
  createMode.value = mode;
  createTitle.value =
    mode === "hotfix"
      ? "Срочное исправление"
      : mode === "follow-up"
        ? "Следующий этап после текущего результата"
        : mode === "workflow"
          ? "Новая инициатива"
          : "Новая задача";
  createDraft.value =
    mode === "hotfix"
      ? "Опишите симптом, влияние на пользователей, желаемую срочность и что надо проверить после выкладки."
      : mode === "follow-up"
        ? "Опишите, какой следующий Issue должен появиться после текущего PR или этапа, и что в нем должно быть зафиксировано."
        : mode === "workflow"
          ? "Опишите цель инициативы, ограничения, желаемые артефакты и какой этап должен идти первым."
          : "Опишите, что должен сделать агент и какой результат должен появиться в GitHub.";
  createWorkflowId.value =
    mode === "hotfix"
      ? "workflow-hotfix"
      : prototype.currentWorkflow?.workflowId || prototype.workflowOptions[0]?.workflowId || "";
  createDialogOpen.value = true;
}

function buildCreatePrompt(): string {
  const title = createTitle.value.trim();
  const draft = createDraft.value.trim();
  const workflowTitle =
    prototype.workflowOptions.find((workflow) => workflow.workflowId === createWorkflowId.value)?.title || "выбранный workflow";

  switch (createMode.value) {
    case "workflow":
      return `Запусти workflow "${workflowTitle}" для проекта "${prototype.currentProject?.title || "текущий проект"}". Название: ${title || "без названия"}. Описание: ${draft}. Агент должен создать GitHub Issue и нужные follow-up артефакты через gh, а в body добавить watermark-блок workflow, stage и expected-next-artifact.`;
    case "hotfix":
      return `Подготовь hotfix по шаблону "${workflowTitle}" для проекта "${prototype.currentProject?.title || "текущий проект"}". Название: ${title || "без названия"}. Описание: ${draft}. Агент должен открыть triage issue, рабочий PR и follow-up issue через gh и заполнить служебный блок в body.`;
    case "follow-up":
      return `Создай follow-up GitHub Issue для проекта "${prototype.currentProject?.title || "текущий проект"}". Название: ${title || "без названия"}. Описание: ${draft}. В body должны быть stage, связь с предыдущим артефактом и expected-next-artifact.`;
    case "issue":
      return `Создай GitHub Issue для проекта "${prototype.currentProject?.title || "текущий проект"}". Название: ${title || "без названия"}. Описание: ${draft}. В body нужно оставить служебный watermark-блок и заполнить обязательные поля приемки.`;
  }
}

function onCreateByVoice(): void {
  voiceDraft.value = buildCreatePrompt();
  createDialogOpen.value = false;
  voiceDialogOpen.value = true;
}

function onSubmitCreateDialog(): void {
  voiceDraft.value = buildCreatePrompt();
  if (createMode.value === "workflow" || createMode.value === "hotfix") {
    createDialogOpen.value = false;
    updateRoute({
      screen: "studio",
      workflowId: createWorkflowId.value || activeRouteState.value.workflowId,
    });
    return;
  }

  createDialogOpen.value = false;
  voiceDialogOpen.value = true;
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
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.mission-control-page__initiative-button {
  min-width: 300px;
  justify-content: flex-start;
}

.mission-control-page__search {
  min-width: 320px;
  flex: 1;
}

.mission-control-page__loading {
  padding: 18px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.88);
}

.mission-control-page__initiative-sheet,
.mission-control-page__voice-sheet,
.mission-control-page__create-sheet {
  display: grid;
  gap: 14px;
}

.mission-control-page__initiative-actions {
  display: flex;
  justify-content: flex-start;
}

.mission-control-page__initiative-list {
  display: grid;
  gap: 10px;
  max-height: 440px;
  overflow: auto;
}

.mission-control-page__initiative-option {
  display: grid;
  gap: 6px;
  padding: 14px;
  border-radius: 18px;
  border: 1px solid rgba(223, 228, 235, 0.9);
  background: rgba(248, 250, 252, 0.94);
  text-align: left;
}

.mission-control-page__initiative-option-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: rgb(33, 38, 46);
}

.mission-control-page__initiative-option-summary,
.mission-control-page__voice-copy {
  font-size: 0.86rem;
  line-height: 1.5;
  color: rgb(96, 104, 118);
}

.mission-control-page__voice-copy {
  margin: 0;
  font-size: 0.95rem;
}

.mission-control-page__voice-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.mission-control-page__watermark-note {
  padding: 14px 16px;
  border-radius: 18px;
  background: rgba(246, 248, 252, 0.92);
  border: 1px solid rgba(223, 228, 235, 0.92);
}

.mission-control-page__watermark-note strong {
  display: block;
  margin-bottom: 8px;
  font-size: 0.92rem;
  color: rgb(31, 36, 43);
}

.mission-control-page__watermark-note ul {
  margin: 0;
  padding-left: 18px;
  display: grid;
  gap: 6px;
  font-size: 0.86rem;
  line-height: 1.5;
  color: rgb(96, 104, 118);
}

@media (max-width: 900px) {
  .mission-control-page__shell {
    padding: 14px;
  }

  .mission-control-page__initiative-button,
  .mission-control-page__search {
    width: 100%;
    min-width: 0;
    max-width: none;
  }
}
</style>
