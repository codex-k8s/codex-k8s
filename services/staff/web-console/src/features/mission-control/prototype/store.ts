import { computed, ref } from "vue";
import { defineStore } from "pinia";

import {
  buildAttentionCards,
  buildExecutionGroups,
  buildHomeColumns,
  buildProjectOptions,
  buildWorkflowOptions,
  buildWorkflowStudioNodes,
  buildWorkflowStudioRelations,
  buildWorkspaceArtifactViews,
  buildWorkspaceFlowNodes,
  buildWorkspaceFlowRelations,
  buildWorkspaceStageViews,
} from "./presenters";
import { MissionControlPrototypeSourceError, missionControlPrototypeSource } from "./source";
import type {
  MissionControlPrototypeError,
  MissionControlPrototypeModel,
  MissionControlPrototypeRouteState,
  MissionControlScreen,
} from "./types";

const missionControlPrototypeDefaultRouteState: MissionControlPrototypeRouteState = {
  screen: "home",
  projectId: "",
  initiativeId: "",
  workflowId: "",
  artifactId: "",
  search: "",
  homeFilter: "all",
  workspaceView: "overview",
};

function asUiError(error: unknown): MissionControlPrototypeError {
  if (error instanceof MissionControlPrototypeSourceError) {
    return error.uiError;
  }
  return {
    messageKey: "pages.missionControlPrototype.errors.unknown",
    debugMessage: error instanceof Error ? error.message : String(error),
  };
}

export const useMissionControlPrototypeStore = defineStore("missionControlPrototype", () => {
  const model = ref<MissionControlPrototypeModel | null>(null);
  const loading = ref(false);
  const error = ref<MissionControlPrototypeError | null>(null);
  const screen = ref<MissionControlScreen>("home");
  const projectId = ref("");
  const initiativeId = ref("");
  const workflowId = ref("");
  const artifactId = ref("");
  const search = ref("");
  const homeFilter = ref(missionControlPrototypeDefaultRouteState.homeFilter);
  const workspaceView = ref(missionControlPrototypeDefaultRouteState.workspaceView);

  const defaultProjectId = computed(() => model.value?.projects[0]?.projectId ?? "");
  const defaultInitiativeId = computed(
    () => model.value?.initiatives.find((initiative) => initiative.projectId === projectId.value)?.initiativeId ?? "",
  );
  const defaultWorkflowId = computed(
    () =>
      model.value?.workflows.find(
        (workflow) => workflow.kind === "project" ? workflow.projectId === projectId.value : true,
      )?.workflowId ?? "",
  );

  const projectOptions = computed(() => buildProjectOptions(model.value));
  const currentProject = computed(() => model.value?.projects.find((project) => project.projectId === projectId.value) ?? null);
  const projectInitiatives = computed(() =>
    model.value?.initiatives.filter((initiative) => initiative.projectId === projectId.value) ?? [],
  );
  const currentInitiative = computed(
    () => model.value?.initiatives.find((initiative) => initiative.initiativeId === initiativeId.value) ?? null,
  );
  const workflowOptions = computed(() => buildWorkflowOptions(model.value, projectId.value));
  const currentWorkflow = computed(
    () => model.value?.workflows.find((workflow) => workflow.workflowId === workflowId.value) ?? null,
  );
  const currentInitiativeArtifacts = computed(() =>
    model.value?.artifacts.filter((artifact) => artifact.initiativeId === initiativeId.value) ?? [],
  );
  const selectedArtifact = computed(
    () => model.value?.artifacts.find((artifact) => artifact.artifactId === artifactId.value) ?? null,
  );
  const currentInitiativeActivity = computed(() =>
    model.value?.activity.filter((item) => item.initiativeId === initiativeId.value) ?? [],
  );
  const attentionCards = computed(() => buildAttentionCards(model.value, projectId.value));
  const homeColumns = computed(() =>
    buildHomeColumns(
      model.value,
      projectId.value,
      search.value,
      screen.value === "home" ? initiativeId.value : "",
      homeFilter.value,
    ),
  );
  const workspaceStageViews = computed(() => buildWorkspaceStageViews(currentInitiative.value, currentWorkflow.value));
  const workspaceArtifacts = computed(() =>
    buildWorkspaceArtifactViews(currentInitiativeArtifacts.value, artifactId.value, search.value),
  );
  const workspaceFlowNodes = computed(() => buildWorkspaceFlowNodes(workspaceStageViews.value));
  const workspaceFlowRelations = computed(() => buildWorkspaceFlowRelations(workspaceStageViews.value));
  const studioNodes = computed(() => buildWorkflowStudioNodes(currentWorkflow.value));
  const studioRelations = computed(() => buildWorkflowStudioRelations(currentWorkflow.value));
  const executionGroups = computed(() => buildExecutionGroups(model.value, projectId.value, search.value));

  async function ensureLoaded(): Promise<void> {
    if (model.value) {
      return;
    }

    loading.value = true;
    error.value = null;

    try {
      model.value = await missionControlPrototypeSource.loadModel();
    } catch (loadError) {
      error.value = asUiError(loadError);
    } finally {
      loading.value = false;
    }
  }

  async function syncRouteState(nextState: MissionControlPrototypeRouteState): Promise<MissionControlPrototypeRouteState> {
    await ensureLoaded();

    if (!model.value) {
      return missionControlPrototypeDefaultRouteState;
    }

    const normalizedProjectId =
      model.value.projects.some((project) => project.projectId === nextState.projectId) ? nextState.projectId : defaultProjectId.value;
    const initiativesForProject = model.value.initiatives.filter((initiative) => initiative.projectId === normalizedProjectId);
    const initiativeIsOptional = nextState.screen !== "initiative";
    const requestedInitiativeId = initiativesForProject.some((initiative) => initiative.initiativeId === nextState.initiativeId)
      ? nextState.initiativeId
      : "";
    const normalizedInitiativeId =
      requestedInitiativeId !== ""
        ? requestedInitiativeId
        : initiativeIsOptional
          ? ""
          : initiativesForProject[0]?.initiativeId ?? "";
    const workflowsForProject = model.value.workflows.filter((workflow) =>
      workflow.kind === "system" ? true : workflow.projectId === normalizedProjectId,
    );
    const initiativeWorkflowId =
      initiativesForProject.find((initiative) => initiative.initiativeId === normalizedInitiativeId)?.workflowId ?? "";
    const normalizedWorkflowId = workflowsForProject.some((workflow) => workflow.workflowId === nextState.workflowId)
      ? nextState.workflowId
      : initiativeWorkflowId || workflowsForProject[0]?.workflowId || "";
    const initiativeArtifactIds =
      initiativesForProject.find((initiative) => initiative.initiativeId === normalizedInitiativeId)?.artifactIds ?? [];
    const normalizedArtifactId = initiativeArtifactIds.includes(nextState.artifactId) ? nextState.artifactId : "";

    projectId.value = normalizedProjectId;
    initiativeId.value = normalizedInitiativeId;
    workflowId.value = normalizedWorkflowId;
    artifactId.value = normalizedArtifactId;
    screen.value = nextState.screen;
    search.value = nextState.search.trim();
    homeFilter.value = nextState.homeFilter;
    workspaceView.value = nextState.workspaceView;

    return {
      screen: nextState.screen,
      projectId: normalizedProjectId,
      initiativeId: normalizedInitiativeId,
      workflowId: normalizedWorkflowId,
      artifactId: normalizedArtifactId,
      search: nextState.search.trim(),
      homeFilter: nextState.homeFilter,
      workspaceView: nextState.workspaceView,
    };
  }

  return {
    loading,
    error,
    screen,
    projectId,
    initiativeId,
    workflowId,
    artifactId,
    search,
    homeFilter,
    workspaceView,
    defaultProjectId,
    defaultInitiativeId,
    defaultWorkflowId,
    currentProject,
    currentInitiative,
    currentWorkflow,
    currentInitiativeArtifacts,
    selectedArtifact,
    currentInitiativeActivity,
    projectOptions,
    projectInitiatives,
    workflowOptions,
    attentionCards,
    homeColumns,
    workspaceStageViews,
    workspaceArtifacts,
    workspaceFlowNodes,
    workspaceFlowRelations,
    studioNodes,
    studioRelations,
    executionGroups,
    ensureLoaded,
    syncRouteState,
  };
});
