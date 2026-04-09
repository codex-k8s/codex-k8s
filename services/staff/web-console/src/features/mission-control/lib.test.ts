import test from "node:test";
import assert from "node:assert/strict";

import {
  buildMissionControlRouteQuery,
  missionControlGraphColumns,
  missionControlNodeKey,
  normalizeMissionControlRouteQuery,
  patchMissionControlRouteState,
  parseMissionControlRealtimeEnvelope,
  workspaceFreshnessStatus,
} from "./lib.ts";

test("normalizeMissionControlRouteQuery applies graph defaults and accepts node deep-link", () => {
  const state = normalizeMissionControlRouteQuery({
    view: "list",
    filter: "blocked",
    q: "review",
    node_kind: "pull_request",
    node_id: "codex-k8s/kodex/pull/1",
  });

  assert.deepEqual(state, {
    viewMode: "list",
    statePreset: "blocked",
    search: "review",
    nodeKind: "pull_request",
    nodePublicId: "codex-k8s/kodex/pull/1",
  });
});

test("normalizeMissionControlRouteQuery upgrades legacy board and agent links", () => {
  const state = normalizeMissionControlRouteQuery({
    view: "board",
    entity_kind: "agent",
    entity_id: "run-1",
  });

  assert.deepEqual(state, {
    viewMode: "graph",
    statePreset: "all_active",
    search: "",
    nodeKind: "run",
    nodePublicId: "run-1",
  });
});

test("buildMissionControlRouteQuery omits default values", () => {
  const query = buildMissionControlRouteQuery({
    viewMode: "graph",
    statePreset: "all_active",
    search: "",
    nodeKind: "",
    nodePublicId: "",
  });

  assert.deepEqual(query, {
    view: undefined,
    filter: undefined,
    q: undefined,
    node_kind: undefined,
    node_id: undefined,
  });
});

test("patchMissionControlRouteState preserves selected node across workspace controls", () => {
  const state = {
    viewMode: "graph" as const,
    statePreset: "all_active" as const,
    search: "",
    nodeKind: "run" as const,
    nodePublicId: "run-1",
  };

  assert.deepEqual(patchMissionControlRouteState(state, { viewMode: "list" }), {
    viewMode: "list",
    statePreset: "all_active",
    search: "",
    nodeKind: "run",
    nodePublicId: "run-1",
  });
  assert.deepEqual(patchMissionControlRouteState(state, { statePreset: "blocked" }), {
    viewMode: "graph",
    statePreset: "blocked",
    search: "",
    nodeKind: "run",
    nodePublicId: "run-1",
  });
  assert.deepEqual(patchMissionControlRouteState(state, { search: "review" }), {
    viewMode: "graph",
    statePreset: "all_active",
    search: "review",
    nodeKind: "run",
    nodePublicId: "run-1",
  });
});

test("missionControlGraphColumns groups nodes by root and column order", () => {
  const columns = missionControlGraphColumns(
    "issue-1",
    new Map([
      [
        missionControlNodeKey({ node_kind: "work_item", node_public_id: "issue-1" }),
        {
          node_kind: "work_item",
          node_public_id: "issue-1",
          title: "Issue 1",
          visibility_tier: "primary",
          active_state: "working",
          continuity_status: "complete",
          coverage_class: "open_primary",
          root_node_public_id: "issue-1",
          column_index: 0,
          has_blocking_gap: false,
          badges: [],
          projection_version: 1,
        },
      ],
      [
        missionControlNodeKey({ node_kind: "run", node_public_id: "run-1" }),
        {
          node_kind: "run",
          node_public_id: "run-1",
          title: "Run 1",
          visibility_tier: "secondary_dimmed",
          active_state: "waiting",
          continuity_status: "missing_pull_request",
          coverage_class: "recent_closed_context",
          root_node_public_id: "issue-1",
          column_index: 1,
          has_blocking_gap: true,
          badges: ["continuity_gap"],
          projection_version: 2,
        },
      ],
    ]),
  );

  assert.equal(columns.length, 2);
  assert.equal(columns[0]?.columnIndex, 0);
  assert.equal(columns[1]?.columnIndex, 1);
  assert.equal(columns[1]?.nodes[0]?.node_public_id, "run-1");
});

test("workspaceFreshnessStatus prefers degraded over stale", () => {
  assert.equal(
    workspaceFreshnessStatus({
      snapshot_id: "snapshot-1",
      view_mode: "graph",
      generated_at: "2026-03-19T12:00:00Z",
      resume_token: "resume-1",
      effective_filters: {
        open_scope: "open_only",
        assignment_scope: "assigned_to_me_or_unassigned",
        state_preset: "all_active",
      },
      summary: {
        root_count: 1,
        node_count: 1,
        blocking_gap_count: 0,
        warning_gap_count: 0,
        recent_closed_context_count: 0,
        working_count: 1,
        waiting_count: 0,
        blocked_count: 0,
        review_count: 0,
        recent_critical_updates_count: 0,
      },
      workspace_watermarks: [
        {
          watermark_kind: "provider_freshness",
          status: "stale",
          summary: "provider lag",
          observed_at: "2026-03-19T11:59:00Z",
        },
        {
          watermark_kind: "graph_projection",
          status: "degraded",
          summary: "projection rebuild lag",
          observed_at: "2026-03-19T12:00:00Z",
        },
      ],
      root_groups: [],
      nodes: [],
      edges: [],
    }),
    "degraded",
  );
});

test("parseMissionControlRealtimeEnvelope accepts legacy delta payload", () => {
  const parsed = parseMissionControlRealtimeEnvelope(
    JSON.stringify({
      event_kind: "delta",
      snapshot_id: "snapshot-2",
      resume_token: "resume-2",
      occurred_at: "2026-03-19T11:00:00Z",
      payload: {
        delta_entities: [{ entity_public_id: "issue-2" }],
        delta_relations: [{ relation_kind: "linked_to" }],
        delta_timeline_entries: [{ entry_id: "timeline-2" }],
        changed_command_ids: ["command-2"],
      },
    }),
  );

  assert.deepEqual(parsed, {
    event_kind: "delta",
    snapshot_id: "snapshot-2",
    resume_token: "resume-2",
    occurred_at: "2026-03-19T11:00:00Z",
    payload: {
      changed_command_ids: ["command-2"],
      impact_count: 3,
    },
  });
});

test("parseMissionControlRealtimeEnvelope accepts workspace delta payload", () => {
  const parsed = parseMissionControlRealtimeEnvelope(
    JSON.stringify({
      event_kind: "delta",
      snapshot_id: "snapshot-3",
      resume_token: "resume-3",
      occurred_at: "2026-03-19T12:00:00Z",
      payload: {
        delta_nodes: [{ node_public_id: "run-3" }],
        delta_edges: [{ edge_kind: "spawned_run" }],
        delta_gaps: [{ gap_id: 10 }],
        delta_workspace_watermarks: [{ watermark_kind: "provider_freshness" }],
        changed_command_ids: [],
      },
    }),
  );

  assert.deepEqual(parsed, {
    event_kind: "delta",
    snapshot_id: "snapshot-3",
    resume_token: "resume-3",
    occurred_at: "2026-03-19T12:00:00Z",
    payload: {
      changed_command_ids: [],
      impact_count: 4,
    },
  });
});

test("parseMissionControlRealtimeEnvelope rejects invalid payload", () => {
  assert.equal(
    parseMissionControlRealtimeEnvelope(
      JSON.stringify({
        event_kind: "connected",
        snapshot_id: "snapshot-1",
        resume_token: "resume-1",
        occurred_at: "2026-03-19T10:00:00Z",
        payload: {
          snapshot_freshness_status: "broken",
          server_cursor: "cursor-1",
        },
      }),
    ),
    null,
  );
});
