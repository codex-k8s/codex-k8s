import type { LocationQuery, LocationQueryRaw } from "vue-router";

import type {
  MissionControlPrototypeRouteState,
  MissionControlScreen,
  MissionInitiativeWorkspaceView,
} from "./types";

const defaultScreen: MissionControlScreen = "home";
const defaultWorkspaceView: MissionInitiativeWorkspaceView = "overview";

function asQueryString(value: LocationQuery[string]): string {
  if (typeof value === "string") {
    return value.trim();
  }
  if (Array.isArray(value) && typeof value[0] === "string") {
    return value[0].trim();
  }
  return "";
}

function isScreen(value: string): value is MissionControlScreen {
  return value === "home" || value === "initiative" || value === "studio" || value === "executions";
}

function isWorkspaceView(value: string): value is MissionInitiativeWorkspaceView {
  return value === "overview" || value === "flow" || value === "artifacts" || value === "activity";
}

export function normalizeMissionControlPrototypeRouteQuery(query: LocationQuery): MissionControlPrototypeRouteState {
  const rawScreen = asQueryString(query.screen);
  const rawView = asQueryString(query.view);

  return {
    screen: isScreen(rawScreen) ? rawScreen : defaultScreen,
    projectId: asQueryString(query.project),
    initiativeId: asQueryString(query.initiative),
    workflowId: asQueryString(query.workflow),
    artifactId: asQueryString(query.artifact),
    search: asQueryString(query.q),
    workspaceView: isWorkspaceView(rawView) ? rawView : defaultWorkspaceView,
  };
}

export function buildMissionControlPrototypeRouteQuery(
  state: MissionControlPrototypeRouteState,
  defaults: {
    projectId: string;
    initiativeId: string;
    workflowId: string;
  },
): LocationQueryRaw {
  return {
    screen: state.screen !== defaultScreen ? state.screen : undefined,
    project: state.projectId !== "" && state.projectId !== defaults.projectId ? state.projectId : undefined,
    initiative: state.initiativeId !== "" && state.initiativeId !== defaults.initiativeId ? state.initiativeId : undefined,
    workflow: state.workflowId !== "" && state.workflowId !== defaults.workflowId ? state.workflowId : undefined,
    artifact: state.artifactId || undefined,
    q: state.search || undefined,
    view: state.workspaceView !== defaultWorkspaceView ? state.workspaceView : undefined,
  };
}

export function patchMissionControlPrototypeRouteState(
  current: MissionControlPrototypeRouteState,
  patch: Partial<MissionControlPrototypeRouteState>,
): MissionControlPrototypeRouteState {
  return {
    ...current,
    ...patch,
  };
}

export function missionControlPrototypeRouteStateEquals(
  left: MissionControlPrototypeRouteState,
  right: MissionControlPrototypeRouteState,
): boolean {
  return (
    left.screen === right.screen &&
    left.projectId === right.projectId &&
    left.initiativeId === right.initiativeId &&
    left.workflowId === right.workflowId &&
    left.artifactId === right.artifactId &&
    left.search === right.search &&
    left.workspaceView === right.workspaceView
  );
}
