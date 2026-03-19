import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { getMissionControlNode, getMissionControlWorkspace, listMissionControlNodeActivity } from "./api";
import {
  missionControlDefaultStatePreset,
  missionControlDefaultViewMode,
  missionControlEdgeKey,
  missionControlNodeKey,
  missionControlWatermarkKey,
  sortMissionControlActivity,
  workspaceFreshnessStatus,
} from "./lib";
import type {
  MissionControlActivityEntry,
  MissionControlEdge,
  MissionControlNode,
  MissionControlNodeDetails,
  MissionControlRealtimeNotice,
  MissionControlRealtimeState,
  MissionControlSelectedNodeRef,
  MissionControlWorkspaceFreshnessStatus,
  MissionControlWorkspaceSnapshot,
  MissionControlWorkspaceWatermark,
} from "./types";

const missionControlDefaultLimit = 30;
const missionControlActivityLimit = 50;

function mergeMissionControlNodes(items: MissionControlNode[]): MissionControlNode[] {
  const byKey = new Map<string, MissionControlNode>();
  for (const item of items) {
    byKey.set(missionControlNodeKey(item), item);
  }
  return Array.from(byKey.values());
}

function mergeMissionControlEdges(items: MissionControlEdge[]): MissionControlEdge[] {
  const byKey = new Map<string, MissionControlEdge>();
  for (const item of items) {
    byKey.set(missionControlEdgeKey(item), item);
  }
  return Array.from(byKey.values());
}

function mergeMissionControlWatermarks(items: MissionControlWorkspaceWatermark[]): MissionControlWorkspaceWatermark[] {
  const byKey = new Map<string, MissionControlWorkspaceWatermark>();
  for (const item of items) {
    byKey.set(missionControlWatermarkKey(item), item);
  }
  return Array.from(byKey.values());
}

function mergeMissionControlActivity(items: MissionControlActivityEntry[]): MissionControlActivityEntry[] {
  const byKey = new Map<string, MissionControlActivityEntry>();
  for (const item of items) {
    byKey.set(item.entry_id, item);
  }
  return sortMissionControlActivity(Array.from(byKey.values()));
}

function mergeMissionControlSnapshot(
  current: MissionControlWorkspaceSnapshot,
  next: MissionControlWorkspaceSnapshot,
): MissionControlWorkspaceSnapshot {
  const rootGroups = [...current.root_groups];
  for (const group of next.root_groups) {
    const existingIndex = rootGroups.findIndex(
      (item) =>
        item.root_node_kind === group.root_node_kind && item.root_node_public_id === group.root_node_public_id,
    );
    if (existingIndex >= 0) {
      rootGroups.splice(existingIndex, 1, group);
      continue;
    }
    rootGroups.push(group);
  }

  return {
    ...next,
    root_groups: rootGroups,
    nodes: mergeMissionControlNodes([...current.nodes, ...next.nodes]),
    edges: mergeMissionControlEdges([...current.edges, ...next.edges]),
    workspace_watermarks: mergeMissionControlWatermarks([
      ...current.workspace_watermarks,
      ...next.workspace_watermarks,
    ]),
  };
}

function freshnessStatusFromNotice(
  snapshot: MissionControlWorkspaceSnapshot | null,
  notice: MissionControlRealtimeNotice | null,
  connectedFreshnessStatus: MissionControlWorkspaceFreshnessStatus,
): MissionControlWorkspaceFreshnessStatus {
  if (notice?.kind === "degraded") return "degraded";
  if (notice?.kind === "stale") return "stale";
  if (connectedFreshnessStatus !== "") return connectedFreshnessStatus;
  return workspaceFreshnessStatus(snapshot);
}

export const useMissionControlStore = defineStore("missionControl", {
  state: () => ({
    viewMode: missionControlDefaultViewMode as "graph" | "list",
    statePreset: missionControlDefaultStatePreset as
      | "working"
      | "waiting"
      | "blocked"
      | "review"
      | "recent_critical_updates"
      | "all_active",
    search: "",
    snapshot: null as MissionControlWorkspaceSnapshot | null,
    loading: false,
    loadingMore: false,
    refreshing: false,
    error: null as ApiError | null,
    selectedRef: null as MissionControlSelectedNodeRef | null,
    selectedDetails: null as MissionControlNodeDetails | null,
    selectedActivity: [] as MissionControlActivityEntry[],
    selectedActivityCursor: "",
    selectedLoading: false,
    selectedActivityLoading: false,
    selectedDetailsError: null as ApiError | null,
    selectedActivityError: null as ApiError | null,
    realtimeState: "closed" as MissionControlRealtimeState | "closed",
    realtimeNotice: null as MissionControlRealtimeNotice | null,
    connectedFreshnessStatus: "" as MissionControlWorkspaceFreshnessStatus,
    snapshotRequestId: 0,
    detailsRequestId: 0,
    activityRequestId: 0,
  }),
  getters: {
    effectiveFreshnessStatus(state): MissionControlWorkspaceFreshnessStatus {
      return freshnessStatusFromNotice(state.snapshot, state.realtimeNotice, state.connectedFreshnessStatus);
    },
    hasMore(state): boolean {
      return String(state.snapshot?.next_root_cursor || "").trim() !== "";
    },
    hasSelectedActivityMore(state): boolean {
      return state.selectedActivityCursor.trim() !== "";
    },
    selectedNode(state): MissionControlNode | null {
      if (state.selectedDetails) {
        return state.selectedDetails.node;
      }
      if (!state.selectedRef || !state.snapshot) {
        return null;
      }
      return (
        state.snapshot.nodes.find(
          (node) =>
            node.node_kind === state.selectedRef?.node_kind &&
            node.node_public_id === state.selectedRef?.node_public_id,
        ) ?? null
      );
    },
  },
  actions: {
    configureQuery(params: {
      viewMode: "graph" | "list";
      statePreset: "working" | "waiting" | "blocked" | "review" | "recent_critical_updates" | "all_active";
      search: string;
    }): void {
      this.viewMode = params.viewMode;
      this.statePreset = params.statePreset;
      this.search = params.search.trim();
    },

    setRealtimeState(state: MissionControlRealtimeState | "closed"): void {
      this.realtimeState = state;
    },

    clearRealtimeNotice(): void {
      this.realtimeNotice = null;
    },

    applyRealtimeNotice(notice: MissionControlRealtimeNotice): void {
      this.realtimeNotice = notice;
    },

    setConnectedFreshnessStatus(status: MissionControlWorkspaceFreshnessStatus): void {
      this.connectedFreshnessStatus = status;
    },

    async loadSnapshot(options: { append?: boolean; refresh?: boolean } = {}): Promise<void> {
      const requestId = this.snapshotRequestId + 1;
      this.snapshotRequestId = requestId;

      const append = options.append === true;
      const refresh = options.refresh === true;
      if (append) {
        this.loadingMore = true;
      } else if (refresh) {
        this.refreshing = true;
      } else {
        this.loading = true;
      }
      if (!append) {
        this.error = null;
      }

      const cursor = append ? String(this.snapshot?.next_root_cursor || "") : "";

      try {
        const snapshot = await getMissionControlWorkspace({
          viewMode: this.viewMode,
          statePreset: this.statePreset,
          search: this.search,
          cursor,
          limit: missionControlDefaultLimit,
        });
        if (this.snapshotRequestId !== requestId) {
          return;
        }

        this.snapshot = append && this.snapshot ? mergeMissionControlSnapshot(this.snapshot, snapshot) : snapshot;
        this.connectedFreshnessStatus = workspaceFreshnessStatus(snapshot);

        if (
          (this.realtimeNotice?.kind === "stale" || this.realtimeNotice?.kind === "degraded") &&
          workspaceFreshnessStatus(snapshot) === "fresh"
        ) {
          this.realtimeNotice = null;
        }
      } catch (error) {
        if (this.snapshotRequestId !== requestId) {
          return;
        }
        this.error = normalizeApiError(error);
      } finally {
        if (this.snapshotRequestId !== requestId) {
          return;
        }
        this.loading = false;
        this.loadingMore = false;
        this.refreshing = false;
      }
    },

    async refreshSnapshot(): Promise<void> {
      await this.loadSnapshot({ refresh: true });
      if (this.selectedRef) {
        await this.loadSelectedNode(this.selectedRef);
      }
    },

    async loadSelectedNode(ref: MissionControlSelectedNodeRef): Promise<void> {
      this.selectedRef = ref;
      const requestId = this.detailsRequestId + 1;
      this.detailsRequestId = requestId;
      this.selectedLoading = true;
      this.selectedDetailsError = null;
      this.selectedActivityError = null;
      this.selectedActivityCursor = "";

      try {
        const details = await getMissionControlNode({
          nodeKind: ref.node_kind,
          nodePublicId: ref.node_public_id,
        });
        if (this.detailsRequestId !== requestId) {
          return;
        }

        this.selectedDetails = details;
        this.selectedActivity = mergeMissionControlActivity(details.activity_preview);
        this.selectedActivityCursor = "";
        await this.loadSelectedActivity({ append: false });
      } catch (error) {
        if (this.detailsRequestId !== requestId) {
          return;
        }
        this.selectedDetails = null;
        this.selectedActivity = [];
        this.selectedActivityCursor = "";
        this.selectedDetailsError = normalizeApiError(error);
        this.selectedActivityError = null;
      } finally {
        if (this.detailsRequestId !== requestId) {
          return;
        }
        this.selectedLoading = false;
      }
    },

    clearSelectedNode(): void {
      this.selectedRef = null;
      this.selectedDetails = null;
      this.selectedActivity = [];
      this.selectedActivityCursor = "";
      this.selectedLoading = false;
      this.selectedActivityLoading = false;
      this.selectedDetailsError = null;
      this.selectedActivityError = null;
    },

    async loadSelectedActivity(options: { append: boolean }): Promise<void> {
      if (!this.selectedRef) {
        return;
      }

      const requestId = this.activityRequestId + 1;
      this.activityRequestId = requestId;
      this.selectedActivityLoading = true;
      this.selectedActivityError = null;
      const cursor = options.append ? this.selectedActivityCursor : "";

      try {
        const response = await listMissionControlNodeActivity({
          nodeKind: this.selectedRef.node_kind,
          nodePublicId: this.selectedRef.node_public_id,
          cursor,
          limit: missionControlActivityLimit,
        });
        if (this.activityRequestId !== requestId) {
          return;
        }

        this.selectedActivity = mergeMissionControlActivity(
          options.append ? [...this.selectedActivity, ...response.items] : response.items,
        );
        this.selectedActivityCursor = String(response.next_cursor || "");
      } catch (error) {
        if (this.activityRequestId !== requestId) {
          return;
        }
        this.selectedActivityError = normalizeApiError(error);
      } finally {
        if (this.activityRequestId !== requestId) {
          return;
        }
        this.selectedActivityLoading = false;
      }
    },
  },
});
