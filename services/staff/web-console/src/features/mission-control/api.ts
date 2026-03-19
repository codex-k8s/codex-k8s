import {
  getMissionControlNode as getMissionControlNodeRequest,
  getMissionControlWorkspace as getMissionControlWorkspaceRequest,
  listMissionControlNodeActivity as listMissionControlNodeActivityRequest,
  previewMissionControlLaunch as previewMissionControlLaunchRequest,
} from "../../shared/api/sdk";

import type {
  MissionControlLaunchPreview,
  MissionControlNodeActivityItemsResponse,
  MissionControlNodeDetails,
  MissionControlNodeKind,
  MissionControlWorkspaceSnapshot,
} from "./types";

export async function getMissionControlWorkspace(params: {
  viewMode: "graph" | "list";
  statePreset: "working" | "waiting" | "blocked" | "review" | "recent_critical_updates" | "all_active";
  search?: string;
  cursor?: string;
  limit?: number;
}): Promise<MissionControlWorkspaceSnapshot> {
  const response = await getMissionControlWorkspaceRequest({
    query: {
      view_mode: params.viewMode,
      state_preset: params.statePreset,
      search: params.search?.trim() || undefined,
      cursor: params.cursor?.trim() || undefined,
      root_limit: params.limit,
    },
    throwOnError: true,
  });

  return response.data;
}

export async function getMissionControlNode(params: {
  nodeKind: MissionControlNodeKind;
  nodePublicId: string;
}): Promise<MissionControlNodeDetails> {
  const response = await getMissionControlNodeRequest({
    path: {
      node_kind: params.nodeKind,
      node_public_id: params.nodePublicId,
    },
    throwOnError: true,
  });

  return response.data;
}

export async function listMissionControlNodeActivity(params: {
  nodeKind: MissionControlNodeKind;
  nodePublicId: string;
  cursor?: string;
  limit?: number;
}): Promise<MissionControlNodeActivityItemsResponse> {
  const response = await listMissionControlNodeActivityRequest({
    path: {
      node_kind: params.nodeKind,
      node_public_id: params.nodePublicId,
    },
    query: {
      cursor: params.cursor?.trim() || undefined,
      limit: params.limit,
    },
    throwOnError: true,
  });

  return response.data;
}

export async function previewMissionControlLaunch(params: {
  nodeKind: MissionControlNodeKind;
  nodePublicId: string;
  threadKind: "issue" | "pull_request";
  threadNumber: number;
  targetLabel: string;
  removedLabels?: string[];
  expectedProjectionVersion?: number;
}): Promise<MissionControlLaunchPreview> {
  const response = await previewMissionControlLaunchRequest({
    body: {
      node_kind: params.nodeKind,
      node_public_id: params.nodePublicId,
      thread_kind: params.threadKind,
      thread_number: params.threadNumber,
      target_label: params.targetLabel,
      removed_labels: params.removedLabels,
      expected_projection_version: params.expectedProjectionVersion,
    },
    throwOnError: true,
  });

  return response.data;
}
