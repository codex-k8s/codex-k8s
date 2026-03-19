import type { LocationQuery, LocationQueryRaw } from "vue-router";

import type {
  MissionControlActivityEntry,
  MissionControlEdge,
  MissionControlNode,
  MissionControlNodeKind,
  MissionControlRealtimeEvent,
  MissionControlRouteState,
  MissionControlSelectedNodeRef,
  MissionControlWorkspaceFreshnessStatus,
  MissionControlWorkspaceSnapshot,
  MissionControlWorkspaceWatermark,
} from "./types";

export const missionControlDefaultViewMode = "graph";
export const missionControlDefaultStatePreset = "all_active";

const missionControlViewModes = ["graph", "list"] as const;
const missionControlStatePresets = ["working", "waiting", "blocked", "review", "recent_critical_updates", "all_active"] as const;
const missionControlNodeKinds = ["discussion", "work_item", "run", "pull_request"] as const;
const missionControlRealtimeKinds = ["connected", "delta", "invalidate", "stale", "degraded", "resync_required", "heartbeat", "error"] as const;

type JsonRecord = Record<string, unknown>;

export type MissionControlGraphColumn = {
  columnIndex: number;
  nodeKinds: MissionControlNodeKind[];
  nodes: MissionControlNode[];
};

function isJsonRecord(value: unknown): value is JsonRecord {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function isString(value: unknown): value is string {
  return typeof value === "string";
}

function isStringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.every((item) => typeof item === "string");
}

function asTrimmedString(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function asQueryString(value: LocationQuery[string]): string {
  if (typeof value === "string") return value.trim();
  if (Array.isArray(value)) {
    return typeof value[0] === "string" ? value[0].trim() : "";
  }
  return "";
}

function isMissionControlViewMode(value: string): value is MissionControlRouteState["viewMode"] {
  return (missionControlViewModes as readonly string[]).includes(value);
}

function isMissionControlStatePreset(value: string): value is MissionControlRouteState["statePreset"] {
  return (missionControlStatePresets as readonly string[]).includes(value);
}

export function isMissionControlNodeKind(value: string): value is MissionControlNodeKind {
  return (missionControlNodeKinds as readonly string[]).includes(value);
}

function nodeKindFromLegacyEntityKind(value: string): MissionControlNodeKind | "" {
  if (value === "agent") return "run";
  return isMissionControlNodeKind(value) ? value : "";
}

export function normalizeMissionControlRouteQuery(query: LocationQuery): MissionControlRouteState {
  const viewMode = asQueryString(query.view);
  const statePreset = asQueryString(query.filter);
  const search = asQueryString(query.q);

  const rawNodeKind = asQueryString(query.node_kind) || asQueryString(query.entity_kind);
  const rawNodePublicId = asQueryString(query.node_id) || asQueryString(query.entity_id);
  const nodeKind = nodeKindFromLegacyEntityKind(rawNodeKind);

  return {
    viewMode: isMissionControlViewMode(viewMode) ? viewMode : viewMode === "board" ? "graph" : missionControlDefaultViewMode,
    statePreset: isMissionControlStatePreset(statePreset) ? statePreset : missionControlDefaultStatePreset,
    search,
    nodeKind,
    nodePublicId: nodeKind && rawNodePublicId !== "" ? rawNodePublicId : "",
  };
}

export function buildMissionControlRouteQuery(state: MissionControlRouteState): LocationQueryRaw {
  return {
    view: state.viewMode !== missionControlDefaultViewMode ? state.viewMode : undefined,
    filter: state.statePreset !== missionControlDefaultStatePreset ? state.statePreset : undefined,
    q: state.search.trim() !== "" ? state.search.trim() : undefined,
    node_kind: state.nodeKind || undefined,
    node_id: state.nodePublicId || undefined,
  };
}

export function patchMissionControlRouteState(
  current: MissionControlRouteState,
  patch: Partial<MissionControlRouteState>,
): MissionControlRouteState {
  return {
    ...current,
    ...patch,
  };
}

export function missionControlRouteStateEquals(left: MissionControlRouteState, right: MissionControlRouteState): boolean {
  return (
    left.viewMode === right.viewMode &&
    left.statePreset === right.statePreset &&
    left.search === right.search &&
    left.nodeKind === right.nodeKind &&
    left.nodePublicId === right.nodePublicId
  );
}

export function missionControlNodeKey(node: Pick<MissionControlNode, "node_kind" | "node_public_id">): string {
  return `${node.node_kind}:${node.node_public_id}`;
}

export function missionControlSelectedNodeKey(node: MissionControlSelectedNodeRef): string {
  return missionControlNodeKey(node);
}

export function missionControlEdgeKey(edge: MissionControlEdge): string {
  return [
    edge.edge_kind,
    edge.source_node_kind,
    edge.source_node_public_id,
    edge.target_node_kind,
    edge.target_node_public_id,
    edge.source_of_truth,
    edge.visibility_tier,
    edge.is_primary_path ? "primary" : "secondary",
  ].join(":");
}

export function missionControlWatermarkKey(watermark: MissionControlWorkspaceWatermark): string {
  return watermark.watermark_kind;
}

export function workspaceFreshnessStatus(snapshot: MissionControlWorkspaceSnapshot | null): MissionControlWorkspaceFreshnessStatus {
  if (!snapshot) return "";

  let status: MissionControlWorkspaceFreshnessStatus = "fresh";
  for (const watermark of snapshot.workspace_watermarks) {
    if (watermark.status === "degraded") {
      return "degraded";
    }
    if (watermark.status === "stale") {
      status = "stale";
    }
  }
  return status;
}

export function missionControlGraphColumns(
  rootNodePublicId: string,
  nodesByKey: Map<string, MissionControlNode>,
): MissionControlGraphColumn[] {
  const columns = new Map<number, MissionControlNode[]>();

  for (const node of nodesByKey.values()) {
    if (node.root_node_public_id !== rootNodePublicId) {
      continue;
    }
    const bucket = columns.get(node.column_index);
    if (bucket) {
      bucket.push(node);
      continue;
    }
    columns.set(node.column_index, [node]);
  }

  return Array.from(columns.entries())
    .sort((left, right) => left[0] - right[0])
    .map(([columnIndex, items]) => {
      const nodes = [...items].sort((left, right) => {
        if (left.visibility_tier !== right.visibility_tier) {
          return left.visibility_tier === "primary" ? -1 : 1;
        }
        if (left.has_blocking_gap !== right.has_blocking_gap) {
          return left.has_blocking_gap ? -1 : 1;
        }
        return left.title.localeCompare(right.title);
      });

      return {
        columnIndex,
        nodeKinds: Array.from(new Set(nodes.map((node) => node.node_kind))),
        nodes,
      };
    });
}

export function missionControlColumnKindKey(column: MissionControlGraphColumn): string {
  return column.nodeKinds.join(":");
}

function parseLegacyImpactCount(payload: JsonRecord): number {
  const deltaEntities = Array.isArray(payload.delta_entities) ? payload.delta_entities.length : 0;
  const deltaRelations = Array.isArray(payload.delta_relations) ? payload.delta_relations.length : 0;
  const deltaTimelineEntries = Array.isArray(payload.delta_timeline_entries) ? payload.delta_timeline_entries.length : 0;
  return deltaEntities + deltaRelations + deltaTimelineEntries;
}

function parseWorkspaceImpactCount(payload: JsonRecord): number {
  const deltaNodes = Array.isArray(payload.delta_nodes) ? payload.delta_nodes.length : 0;
  const deltaEdges = Array.isArray(payload.delta_edges) ? payload.delta_edges.length : 0;
  const deltaGaps = Array.isArray(payload.delta_gaps) ? payload.delta_gaps.length : 0;
  const deltaWatermarks = Array.isArray(payload.delta_workspace_watermarks) ? payload.delta_workspace_watermarks.length : 0;
  return deltaNodes + deltaEdges + deltaGaps + deltaWatermarks;
}

function parseAffectedCount(payload: JsonRecord): number {
  if (Array.isArray(payload.affected_node_refs)) {
    return payload.affected_node_refs.length;
  }
  if (Array.isArray(payload.affected_entity_refs)) {
    return payload.affected_entity_refs.length;
  }
  return 0;
}

export function parseMissionControlRealtimeEnvelope(raw: string): MissionControlRealtimeEvent | null {
  const text = String(raw || "").trim();
  if (text === "") return null;

  let parsed: unknown;
  try {
    parsed = JSON.parse(text);
  } catch {
    return null;
  }
  if (!isJsonRecord(parsed)) return null;

  const eventKind = asTrimmedString(parsed.event_kind);
  if (!(missionControlRealtimeKinds as readonly string[]).includes(eventKind)) {
    return null;
  }

  const snapshotId = asTrimmedString(parsed.snapshot_id);
  const resumeToken = asTrimmedString(parsed.resume_token);
  const occurredAt = asTrimmedString(parsed.occurred_at);
  const payload = parsed.payload;
  if (snapshotId === "" || occurredAt === "" || !isJsonRecord(payload)) {
    return null;
  }

  switch (eventKind) {
    case "connected": {
      const freshnessStatus = asTrimmedString(payload.snapshot_freshness_status);
      const serverCursor = asTrimmedString(payload.server_cursor);
      if ((freshnessStatus !== "fresh" && freshnessStatus !== "stale" && freshnessStatus !== "degraded") || serverCursor === "") {
        return null;
      }
      return {
        event_kind: "connected",
        snapshot_id: snapshotId,
        resume_token: resumeToken,
        occurred_at: occurredAt,
        payload: {
          snapshot_freshness_status: freshnessStatus,
          server_cursor: serverCursor,
        },
      };
    }
    case "delta": {
      const changedCommandIds = isStringArray(payload.changed_command_ids) ? payload.changed_command_ids : null;
      if (changedCommandIds === null) {
        return null;
      }

      return {
        event_kind: "delta",
        snapshot_id: snapshotId,
        resume_token: resumeToken,
        occurred_at: occurredAt,
        payload: {
          changed_command_ids: changedCommandIds,
          impact_count: parseLegacyImpactCount(payload) + parseWorkspaceImpactCount(payload),
        },
      };
    }
    case "invalidate": {
      const reason = asTrimmedString(payload.reason);
      const refreshScope = asTrimmedString(payload.refresh_scope);
      if (reason === "" || refreshScope === "") {
        return null;
      }

      return {
        event_kind: "invalidate",
        snapshot_id: snapshotId,
        resume_token: resumeToken,
        occurred_at: occurredAt,
        payload: {
          reason,
          refresh_scope: refreshScope,
          affected_count: parseAffectedCount(payload),
        },
      };
    }
    case "stale": {
      const reason = asTrimmedString(payload.reason);
      const staleSince = asTrimmedString(payload.stale_since);
      const suggestedRefresh = asTrimmedString(payload.suggested_refresh);
      if (reason === "" || staleSince === "" || suggestedRefresh === "") {
        return null;
      }

      return {
        event_kind: "stale",
        snapshot_id: snapshotId,
        resume_token: resumeToken,
        occurred_at: occurredAt,
        payload: {
          reason,
          stale_since: staleSince,
          suggested_refresh: suggestedRefresh,
        },
      };
    }
    case "degraded": {
      const reason = asTrimmedString(payload.reason);
      const fallbackMode = asTrimmedString(payload.fallback_mode);
      const affectedCapabilities = isStringArray(payload.affected_capabilities) ? payload.affected_capabilities : [];
      if (reason === "" || fallbackMode === "") {
        return null;
      }

      return {
        event_kind: "degraded",
        snapshot_id: snapshotId,
        resume_token: resumeToken,
        occurred_at: occurredAt,
        payload: {
          reason,
          fallback_mode: fallbackMode,
          affected_capabilities: affectedCapabilities,
        },
      };
    }
    case "resync_required": {
      const reason = asTrimmedString(payload.reason);
      const requiredSnapshotId = asTrimmedString(payload.required_snapshot_id);
      const droppedEventCount = Number(payload.dropped_event_count);
      if (reason === "" || requiredSnapshotId === "" || !Number.isFinite(droppedEventCount)) {
        return null;
      }

      return {
        event_kind: "resync_required",
        snapshot_id: snapshotId,
        resume_token: resumeToken,
        occurred_at: occurredAt,
        payload: {
          reason,
          required_snapshot_id: requiredSnapshotId,
          dropped_event_count: Math.max(0, Math.trunc(droppedEventCount)),
        },
      };
    }
    case "heartbeat": {
      const serverTime = asTrimmedString(payload.server_time);
      const heartbeatSnapshotId = asTrimmedString(payload.snapshot_id);
      if (serverTime === "" || heartbeatSnapshotId === "") {
        return null;
      }

      return {
        event_kind: "heartbeat",
        snapshot_id: snapshotId,
        resume_token: resumeToken,
        occurred_at: occurredAt,
        payload: {
          server_time: serverTime,
          snapshot_id: heartbeatSnapshotId,
        },
      };
    }
    case "error": {
      const code = asTrimmedString(payload.code);
      const message = asTrimmedString(payload.message);
      const retryable = payload.retryable;
      if (code === "" || message === "" || typeof retryable !== "boolean") {
        return null;
      }

      return {
        event_kind: "error",
        snapshot_id: snapshotId,
        resume_token: resumeToken,
        occurred_at: occurredAt,
        payload: {
          code,
          message,
          retryable,
        },
      };
    }
    default:
      return null;
  }
}

export function sortMissionControlActivity(items: MissionControlActivityEntry[]): MissionControlActivityEntry[] {
  return [...items].sort((left, right) => {
    if (left.occurred_at === right.occurred_at) {
      return left.entry_id < right.entry_id ? 1 : -1;
    }
    return left.occurred_at < right.occurred_at ? 1 : -1;
  });
}
