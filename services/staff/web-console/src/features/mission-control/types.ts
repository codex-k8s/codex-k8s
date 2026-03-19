import type {
  MissionControlActivityEntry,
  MissionControlContinuityGap,
  MissionControlEdge,
  MissionControlLaunchPreview,
  MissionControlLaunchSurface,
  MissionControlNode,
  MissionControlNodeActivityItemsResponse,
  MissionControlNodeDetails,
  MissionControlNodeKind,
  MissionControlNodeRef,
  MissionControlProviderDeepLink,
  MissionControlRootGroup,
  MissionControlStageNextStepTemplate,
  MissionControlWorkspaceSnapshot,
  MissionControlWorkspaceStatePreset,
  MissionControlWorkspaceViewMode,
  MissionControlWorkspaceWatermark,
} from "../../shared/api/generated";
import type { RealtimeConnectionState } from "../../shared/ws/realtime-client";

export type {
  MissionControlActivityEntry,
  MissionControlContinuityGap,
  MissionControlEdge,
  MissionControlLaunchPreview,
  MissionControlLaunchSurface,
  MissionControlNode,
  MissionControlNodeActivityItemsResponse,
  MissionControlNodeDetails,
  MissionControlNodeKind,
  MissionControlNodeRef,
  MissionControlProviderDeepLink,
  MissionControlRootGroup,
  MissionControlStageNextStepTemplate,
  MissionControlWorkspaceSnapshot,
  MissionControlWorkspaceStatePreset,
  MissionControlWorkspaceViewMode,
  MissionControlWorkspaceWatermark,
} from "../../shared/api/generated";

export type MissionControlRouteState = {
  viewMode: MissionControlWorkspaceViewMode;
  statePreset: MissionControlWorkspaceStatePreset;
  search: string;
  nodeKind: MissionControlNodeKind | "";
  nodePublicId: string;
};

export type MissionControlSelectedNodeRef = {
  node_kind: MissionControlNodeKind;
  node_public_id: string;
};

export type MissionControlWorkspaceFreshnessStatus = "fresh" | "stale" | "degraded" | "";

export type MissionControlRealtimeState = Exclude<RealtimeConnectionState, "closed">;

export type MissionControlRealtimeEvent =
  | {
      event_kind: "connected";
      snapshot_id: string;
      resume_token: string;
      occurred_at: string;
      payload: {
        snapshot_freshness_status: Exclude<MissionControlWorkspaceFreshnessStatus, "">;
        server_cursor: string;
      };
    }
  | {
      event_kind: "delta";
      snapshot_id: string;
      resume_token: string;
      occurred_at: string;
      payload: {
        changed_command_ids: string[];
        impact_count: number;
      };
    }
  | {
      event_kind: "invalidate";
      snapshot_id: string;
      resume_token: string;
      occurred_at: string;
      payload: {
        reason: string;
        refresh_scope: string;
        affected_count: number;
      };
    }
  | {
      event_kind: "stale";
      snapshot_id: string;
      resume_token: string;
      occurred_at: string;
      payload: {
        reason: string;
        stale_since: string;
        suggested_refresh: string;
      };
    }
  | {
      event_kind: "degraded";
      snapshot_id: string;
      resume_token: string;
      occurred_at: string;
      payload: {
        reason: string;
        fallback_mode: string;
        affected_capabilities: string[];
      };
    }
  | {
      event_kind: "resync_required";
      snapshot_id: string;
      resume_token: string;
      occurred_at: string;
      payload: {
        reason: string;
        required_snapshot_id: string;
        dropped_event_count: number;
      };
    }
  | {
      event_kind: "heartbeat";
      snapshot_id: string;
      resume_token: string;
      occurred_at: string;
      payload: {
        server_time: string;
        snapshot_id: string;
      };
    }
  | {
      event_kind: "error";
      snapshot_id: string;
      resume_token: string;
      occurred_at: string;
      payload: {
        code: string;
        message: string;
        retryable: boolean;
      };
    };

export type MissionControlRealtimeNotice =
  | {
      kind: "invalidate";
      reason: string;
      refreshScope: string;
      affectedCount: number;
      occurredAt: string;
    }
  | {
      kind: "stale";
      reason: string;
      staleSince: string;
      suggestedRefresh: string;
      occurredAt: string;
    }
  | {
      kind: "degraded";
      reason: string;
      fallbackMode: string;
      affectedCapabilities: string[];
      occurredAt: string;
    }
  | {
      kind: "resync_required";
      reason: string;
      requiredSnapshotId: string;
      droppedEventCount: number;
      occurredAt: string;
    }
  | {
      kind: "error";
      code: string;
      message: string;
      retryable: boolean;
      occurredAt: string;
    };

export type MissionControlInfoRow = {
  labelKey: string;
  value: string;
  mono?: boolean;
  href?: string;
};

export type MissionControlRelatedNodesSection = {
  titleKey: string;
  refs: MissionControlNodeRef[];
};
