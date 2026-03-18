import {
  getMissionControlNode as getMissionControlNodeRequest,
  getMissionControlWorkspace as getMissionControlWorkspaceRequest,
  listMissionControlNodeActivity as listMissionControlNodeActivityRequest,
} from "../../shared/api/sdk";

import type {
  MissionControlActivityEntry as MissionControlActivityEntryDto,
  MissionControlActiveFilter,
  MissionControlAllowedAction,
  MissionControlDashboardSnapshot,
  MissionControlEntityCard,
  MissionControlEntityDetails,
  MissionControlEntityDetailsPayload,
  MissionControlEntityKind,
  MissionControlEntityRef,
  MissionControlLaunchSurface as MissionControlLaunchSurfaceDto,
  MissionControlNode as MissionControlNodeDto,
  MissionControlNodeActivityItemsResponse,
  MissionControlNodeDetails as MissionControlNodeDetailsDto,
  MissionControlNodeDetailsPayload as MissionControlNodeDetailsPayloadDto,
  MissionControlNodeRef as MissionControlNodeRefDto,
  MissionControlProviderDeepLink,
  MissionControlRelation,
  MissionControlTimelineEntry,
  MissionControlViewMode,
  MissionControlWorkspaceSnapshot as MissionControlWorkspaceSnapshotDto,
  MissionControlWorkspaceWatermark as MissionControlWorkspaceWatermarkDto,
} from "../../shared/api/generated";

const missionControlFallbackProviderReference: MissionControlEntityCard["provider_reference"] = {
  provider: "platform",
  external_id: "",
};

function entityKey(ref: MissionControlEntityRef): string {
  return `${ref.entity_kind}:${ref.entity_public_id}`;
}

function legacyEntityKindFromNodeKind(nodeKind: MissionControlNodeDto["node_kind"]): MissionControlEntityKind {
  switch (nodeKind) {
    case "run":
      return "agent";
    default:
      return nodeKind;
  }
}

function nodeKindFromLegacyEntityKind(entityKind: MissionControlEntityKind): MissionControlNodeDto["node_kind"] {
  switch (entityKind) {
    case "agent":
      return "run";
    default:
      return entityKind;
  }
}

function legacyViewMode(viewMode: MissionControlWorkspaceSnapshotDto["view_mode"]): MissionControlViewMode {
  return viewMode === "graph" ? "board" : "list";
}

function workspaceViewMode(viewMode: MissionControlViewMode): MissionControlWorkspaceSnapshotDto["view_mode"] {
  return viewMode === "board" ? "graph" : "list";
}

function legacyStateFromNode(node: MissionControlNodeDto): MissionControlEntityCard["state"] {
  switch (node.active_state) {
    case "working":
    case "waiting":
    case "blocked":
    case "review":
    case "recent_critical_updates":
      return node.active_state;
    case "archived":
    default:
      return "recent_critical_updates";
  }
}

function legacySyncStatusFromNode(node: MissionControlNodeDto): MissionControlEntityCard["sync_status"] {
  if (node.continuity_status === "stale_provider" || node.badges.includes("provider_stale")) {
    return "degraded";
  }
  if (node.has_blocking_gap) {
    return "failed";
  }
  if (node.continuity_status !== "complete" && node.continuity_status !== "out_of_scope") {
    return "pending_sync";
  }
  return "synced";
}

function legacyBadgesFromNode(node: MissionControlNodeDto): MissionControlEntityCard["badges"] {
  const badges = new Set<MissionControlEntityCard["badges"][number]>();
  for (const badge of node.badges) {
    switch (badge) {
      case "continuity_gap":
        badges.add("blocked");
        break;
      case "provider_stale":
        badges.add("realtime_stale");
        break;
      case "waiting_mcp":
        badges.add("waiting_mcp");
        break;
      case "review_required":
        badges.add("owner_review");
        break;
      default:
        break;
    }
  }
  return Array.from(badges);
}

function relationKindFromEdge(edgeKind: string): MissionControlRelation["relation_kind"] {
  switch (edgeKind) {
    case "formalized_from":
      return "formalized_from";
    case "blocked_by":
      return "blocked_by";
    case "tracked_by_command":
      return "tracked_by_command";
    default:
      return "linked_to";
  }
}

function relationSourceKind(sourceOfTruth: string): MissionControlRelation["source_kind"] {
  switch (sourceOfTruth) {
    case "provider":
    case "command":
      return sourceOfTruth;
    default:
      return "platform";
  }
}

function legacyRelationFromEdge(edge: MissionControlWorkspaceSnapshotDto["edges"][number]): MissionControlRelation {
  return {
    relation_kind: relationKindFromEdge(edge.edge_kind),
    source_kind: relationSourceKind(edge.source_of_truth),
    source_entity_kind: legacyEntityKindFromNodeKind(edge.source_node_kind),
    source_entity_public_id: edge.source_node_public_id,
    target_entity_kind: legacyEntityKindFromNodeKind(edge.target_node_kind),
    target_entity_public_id: edge.target_node_public_id,
    direction: "outbound",
  };
}

function dedupeRelations(relations: MissionControlRelation[]): MissionControlRelation[] {
  const byKey = new Map<string, MissionControlRelation>();
  for (const relation of relations) {
    byKey.set(
      [
        relation.relation_kind,
        relation.source_kind,
        relation.source_entity_kind,
        relation.source_entity_public_id,
        relation.target_entity_kind,
        relation.target_entity_public_id,
        relation.direction,
      ].join(":"),
      relation,
    );
  }
  return Array.from(byKey.values());
}

function relationCountIndex(relations: MissionControlRelation[]): Map<string, number> {
  const index = new Map<string, number>();
  for (const relation of relations) {
    const sourceKey = entityKey({
      entity_kind: relation.source_entity_kind,
      entity_public_id: relation.source_entity_public_id,
    });
    const targetKey = entityKey({
      entity_kind: relation.target_entity_kind,
      entity_public_id: relation.target_entity_public_id,
    });
    index.set(sourceKey, (index.get(sourceKey) ?? 0) + 1);
    index.set(targetKey, (index.get(targetKey) ?? 0) + 1);
  }
  return index;
}

function legacyEntityCardFromNode(
  node: MissionControlNodeDto,
  relationCounts: Map<string, number>,
): MissionControlEntityCard {
  const entityKind = legacyEntityKindFromNodeKind(node.node_kind);
  const providerReference =
    node.provider_reference ?? {
      ...missionControlFallbackProviderReference,
      external_id: node.node_public_id,
    };

  return {
    entity_kind: entityKind,
    entity_public_id: node.node_public_id,
    title: node.title,
    state: legacyStateFromNode(node),
    sync_status: legacySyncStatusFromNode(node),
    provider_reference: providerReference,
    relation_count: relationCounts.get(entityKey({ entity_kind: entityKind, entity_public_id: node.node_public_id })) ?? 0,
    last_timeline_at: node.last_activity_at ?? undefined,
    badges: legacyBadgesFromNode(node),
    projection_version: node.projection_version,
  };
}

function legacyTimelineEntryFromActivity(item: MissionControlActivityEntryDto): MissionControlTimelineEntry {
  return {
    entry_id: item.entry_id,
    entity_kind: legacyEntityKindFromNodeKind(item.node_kind),
    entity_public_id: item.node_public_id,
    source_kind: item.source_kind,
    source_ref: item.source_ref,
    occurred_at: item.occurred_at,
    summary: item.summary,
    body_markdown: item.body_markdown ?? undefined,
    provider_url: item.provider_url ?? undefined,
    is_read_only: item.is_read_only,
  };
}

function firstNodePublicID(items: NodeRefDtoList | undefined): string | undefined {
  return items && items.length > 0 ? items[0].node_public_id : undefined;
}

type NodeRefDtoList = Array<MissionControlNodeRefDto>;

function findProviderLink(
  links: MissionControlProviderDeepLink[],
  actionKind: MissionControlProviderDeepLink["action_kind"],
): string | undefined {
  const link = links.find((item) => item.action_kind === actionKind && item.url.trim() !== "");
  return link?.url;
}

function legacyDetailPayloadFromNodeDetails(details: MissionControlNodeDetailsDto): MissionControlEntityDetailsPayload {
  const payload = details.detail_payload as MissionControlNodeDetailsPayloadDto;

  if ("discussion_kind" in payload) {
    return {
      discussion_kind: payload.discussion_kind,
      status: payload.status || undefined,
      author: payload.author || undefined,
      participant_count: payload.participant_count,
      latest_comment_excerpt: payload.latest_comment_excerpt || undefined,
      formalization_target: payload.formalization_target_refs[0]?.node_kind || undefined,
    };
  }

  if ("issue_number" in payload) {
    return {
      repository_full_name: payload.repository_full_name || undefined,
      issue_number: payload.issue_number,
      issue_url: findProviderLink(details.provider_deep_links, "provider.open_issue"),
      last_run_id: firstNodePublicID(payload.linked_run_refs),
      stage_label: payload.stage_label || undefined,
      labels: payload.labels,
      assignees: payload.assignees,
      last_provider_sync_at: payload.last_provider_sync_at ?? undefined,
    };
  }

  if ("pull_request_number" in payload) {
    return {
      repository_full_name: payload.repository_full_name || undefined,
      pull_request_number: payload.pull_request_number,
      pull_request_url: findProviderLink(details.provider_deep_links, "provider.open_pr"),
      last_run_id: payload.linked_run_ref?.node_public_id,
      branch_head: payload.branch_head || undefined,
      branch_base: payload.branch_base || undefined,
      merge_state: payload.merge_state || undefined,
      review_decision: payload.review_decision || undefined,
      checks_summary: payload.checks_summary || undefined,
      linked_issue_refs: payload.linked_issue_refs.map((item) => item.node_public_id),
    };
  }

  return {
    agent_key: payload.agent_key,
    run_status: payload.run_status || undefined,
    runtime_mode: payload.runtime_mode || undefined,
    active_run_id: payload.run_id || undefined,
    last_heartbeat_at: payload.finished_at ?? payload.started_at ?? undefined,
    last_run_repository: payload.build_ref || undefined,
  };
}

function legacyAllowedActionsFromLaunchSurfaces(
  items: MissionControlLaunchSurfaceDto[],
): MissionControlAllowedAction[] {
  return items.flatMap((item) => {
    if (item.action_kind !== "preview_next_stage") {
      return [];
    }
    return [
      {
        action_kind: "stage.next_step.execute",
        presentation:
          item.presentation === "primary" || item.presentation === "secondary" || item.presentation === "link"
            ? item.presentation
            : "secondary",
        allowed_when_degraded: false,
        approval_requirement: item.approval_requirement,
        blocked_reason: item.blocked_reason ?? undefined,
      },
    ];
  });
}

function workspaceFreshnessStatus(
  watermarks: MissionControlWorkspaceWatermarkDto[],
): MissionControlDashboardSnapshot["freshness_status"] {
  let status: MissionControlDashboardSnapshot["freshness_status"] = "fresh";
  for (const watermark of watermarks) {
    if (watermark.status === "degraded") {
      return "degraded";
    }
    if (watermark.status === "stale") {
      status = "stale";
    }
  }
  return status;
}

function workspaceStaleAfter(snapshot: MissionControlWorkspaceSnapshotDto): string {
  let latest = snapshot.generated_at;
  for (const watermark of snapshot.workspace_watermarks) {
    const candidate = watermark.window_ended_at || watermark.observed_at;
    if (candidate && candidate > latest) {
      latest = candidate;
    }
  }
  return latest;
}

function dashboardSummaryFromEntities(
  entities: MissionControlEntityCard[],
): MissionControlDashboardSnapshot["summary"] {
  return {
    total_entities: entities.length,
    working_count: entities.filter((item) => item.state === "working").length,
    waiting_count: entities.filter((item) => item.state === "waiting").length,
    blocked_count: entities.filter((item) => item.state === "blocked").length,
    review_count: entities.filter((item) => item.state === "review").length,
    recent_critical_updates_count: entities.filter((item) => item.state === "recent_critical_updates").length,
  };
}

function legacyDashboardSnapshotFromWorkspace(
  snapshot: MissionControlWorkspaceSnapshotDto,
): MissionControlDashboardSnapshot {
  const relations = dedupeRelations(snapshot.edges.map(legacyRelationFromEdge));
  const relationCounts = relationCountIndex(relations);
  const entities = snapshot.nodes.map((node) => legacyEntityCardFromNode(node, relationCounts));

  return {
    snapshot_id: snapshot.snapshot_id,
    view_mode: legacyViewMode(snapshot.view_mode),
    freshness_status: workspaceFreshnessStatus(snapshot.workspace_watermarks),
    generated_at: snapshot.generated_at,
    stale_after: workspaceStaleAfter(snapshot),
    realtime_resume_token: snapshot.resume_token,
    summary: dashboardSummaryFromEntities(entities),
    entities,
    relations,
    next_page_cursor: snapshot.next_root_cursor ?? undefined,
  };
}

function legacyDetailsFromNode(details: MissionControlNodeDetailsDto): MissionControlEntityDetails {
  const relations = dedupeRelations(details.adjacent_edges.map(legacyRelationFromEdge));
  const relationCounts = relationCountIndex(relations);

  return {
    entity: legacyEntityCardFromNode(details.node, relationCounts),
    detail_payload: legacyDetailPayloadFromNodeDetails(details),
    relations,
    timeline_preview: details.activity_preview.map(legacyTimelineEntryFromActivity),
    allowed_actions: legacyAllowedActionsFromLaunchSurfaces(details.launch_surfaces),
    provider_deep_links: details.provider_deep_links,
  };
}

function legacyTimelineFromResponse(
  response: MissionControlNodeActivityItemsResponse,
): { items: MissionControlTimelineEntry[]; nextCursor: string } {
  return {
    items: response.items.map(legacyTimelineEntryFromActivity),
    nextCursor: String(response.next_cursor || ""),
  };
}

export async function getMissionControlDashboard(params: {
  viewMode: MissionControlViewMode;
  activeFilter: MissionControlActiveFilter;
  search?: string;
  cursor?: string;
  limit?: number;
}): Promise<MissionControlDashboardSnapshot> {
  const resp = await getMissionControlWorkspaceRequest({
    query: {
      view_mode: workspaceViewMode(params.viewMode),
      state_preset: params.activeFilter,
      search: params.search?.trim() || undefined,
      cursor: params.cursor?.trim() || undefined,
      root_limit: params.limit,
    },
    throwOnError: true,
  });
  return legacyDashboardSnapshotFromWorkspace(resp.data);
}

export async function getMissionControlEntity(params: {
  entityKind: MissionControlEntityKind;
  entityPublicId: string;
  timelineLimit?: number;
}): Promise<MissionControlEntityDetails> {
  const resp = await getMissionControlNodeRequest({
    path: {
      node_kind: nodeKindFromLegacyEntityKind(params.entityKind),
      node_public_id: params.entityPublicId,
    },
    throwOnError: true,
  });
  return legacyDetailsFromNode(resp.data);
}

export async function listMissionControlTimeline(params: {
  entityKind: MissionControlEntityKind;
  entityPublicId: string;
  cursor?: string;
  limit?: number;
}): Promise<{ items: MissionControlTimelineEntry[]; nextCursor: string }> {
  const resp = await listMissionControlNodeActivityRequest({
    path: {
      node_kind: nodeKindFromLegacyEntityKind(params.entityKind),
      node_public_id: params.entityPublicId,
    },
    query: {
      cursor: params.cursor?.trim() || undefined,
      limit: params.limit,
    },
    throwOnError: true,
  });
  return legacyTimelineFromResponse(resp.data);
}
