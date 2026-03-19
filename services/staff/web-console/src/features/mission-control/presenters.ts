import type {
  MissionControlContinuityGap,
  MissionControlInfoRow,
  MissionControlLaunchSurface,
  MissionControlNode,
  MissionControlNodeDetails,
  MissionControlNodeKind,
  MissionControlNodeRef,
  MissionControlProviderDeepLink,
  MissionControlRelatedNodesSection,
  MissionControlWorkspaceWatermark,
} from "./types";

type DiscussionPayload = Extract<MissionControlNodeDetails["detail_payload"], { discussion_kind: string }>;
type WorkItemPayload = Extract<MissionControlNodeDetails["detail_payload"], { issue_number: number }>;
type RunPayload = Extract<MissionControlNodeDetails["detail_payload"], { run_id: string }>;
type PullRequestPayload = Extract<MissionControlNodeDetails["detail_payload"], { pull_request_number: number }>;

export function isDiscussionPayload(payload: MissionControlNodeDetails["detail_payload"]): payload is DiscussionPayload {
  return "discussion_kind" in payload;
}

export function isWorkItemPayload(payload: MissionControlNodeDetails["detail_payload"]): payload is WorkItemPayload {
  return "issue_number" in payload;
}

export function isRunPayload(payload: MissionControlNodeDetails["detail_payload"]): payload is RunPayload {
  return "run_id" in payload;
}

export function isPullRequestPayload(payload: MissionControlNodeDetails["detail_payload"]): payload is PullRequestPayload {
  return "pull_request_number" in payload;
}

export function missionControlNodeKindLabelKey(kind: MissionControlNodeKind): string {
  return `pages.missionControl.nodeKinds.${kind}`;
}

export function missionControlStateLabelKey(state: MissionControlNode["active_state"]): string {
  return `pages.missionControl.states.${state}`;
}

export function missionControlStateColor(state: MissionControlNode["active_state"]): string {
  switch (state) {
    case "working":
      return "info";
    case "waiting":
      return "warning";
    case "blocked":
      return "error";
    case "review":
      return "success";
    case "recent_critical_updates":
      return "secondary";
    case "archived":
      return "default";
  }
}

export function missionControlVisibilityLabelKey(visibility: MissionControlNode["visibility_tier"]): string {
  return `pages.missionControl.visibility.${visibility}`;
}

export function missionControlVisibilityColor(visibility: MissionControlNode["visibility_tier"]): string {
  return visibility === "primary" ? "primary" : "secondary";
}

export function missionControlContinuityStatusLabelKey(status: MissionControlNode["continuity_status"]): string {
  return `pages.missionControl.continuity.${status}`;
}

export function missionControlContinuityColor(status: MissionControlNode["continuity_status"]): string {
  switch (status) {
    case "complete":
      return "success";
    case "stale_provider":
      return "warning";
    case "out_of_scope":
      return "secondary";
    default:
      return "error";
  }
}

export function missionControlCoverageClassLabelKey(coverageClass: MissionControlNode["coverage_class"]): string {
  return `pages.missionControl.coverage.${coverageClass}`;
}

export function missionControlBadgeLabelKey(badge: MissionControlNode["badges"][number]): string {
  return `pages.missionControl.badges.${badge}`;
}

export function missionControlWatermarkLabelKey(kind: MissionControlWorkspaceWatermark["watermark_kind"]): string {
  return `pages.missionControl.watermarks.${kind}`;
}

export function missionControlWatermarkColor(status: MissionControlWorkspaceWatermark["status"]): string {
  switch (status) {
    case "fresh":
      return "success";
    case "stale":
      return "warning";
    case "degraded":
      return "error";
    case "out_of_scope":
      return "secondary";
  }
}

export function missionControlGapSeverityLabelKey(severity: MissionControlContinuityGap["severity"]): string {
  return `pages.missionControl.gapSeverity.${severity}`;
}

export function missionControlGapSeverityColor(severity: MissionControlContinuityGap["severity"]): string {
  switch (severity) {
    case "blocking":
      return "error";
    case "warning":
      return "warning";
    case "info":
      return "info";
  }
}

export function missionControlGapKindLabelKey(kind: MissionControlContinuityGap["gap_kind"]): string {
  return `pages.missionControl.gapKinds.${kind}`;
}

export function missionControlLaunchSurfaceLabelKey(actionKind: MissionControlLaunchSurface["action_kind"]): string {
  return `pages.missionControl.launchSurfaces.${actionKind}`;
}

export function missionControlLaunchSurfaceIcon(actionKind: MissionControlLaunchSurface["action_kind"]): string {
  switch (actionKind) {
    case "preview_next_stage":
      return "mdi-rocket-launch-outline";
    case "open_linked_pull_request":
      return "mdi-source-pull";
    case "open_linked_follow_up_issue":
      return "mdi-source-branch";
    case "open_provider_context":
      return "mdi-open-in-new";
    case "inspect_run_context":
      return "mdi-radar";
  }
}

export function missionControlProviderDeepLinkLabelKey(actionKind: MissionControlProviderDeepLink["action_kind"]): string {
  return `pages.missionControl.providerLinks.${actionKind}`;
}

export function missionControlProviderDeepLinkIcon(actionKind: MissionControlProviderDeepLink["action_kind"]): string {
  switch (actionKind) {
    case "provider.open_issue":
      return "mdi-link-variant";
    case "provider.open_pr":
      return "mdi-source-pull";
    case "provider.review":
      return "mdi-comment-check-outline";
    case "provider.merge":
      return "mdi-source-merge";
    case "provider.reply_comment":
      return "mdi-comment-edit-outline";
  }
}

function pushInfoRow(
  rows: MissionControlInfoRow[],
  labelKey: string,
  value: string | number | null | undefined,
  options: { mono?: boolean; href?: string } = {},
): void {
  const text = String(value ?? "").trim();
  if (text === "") return;

  rows.push({
    labelKey,
    value: text,
    mono: options.mono,
    href: options.href,
  });
}

export function buildMissionControlInfoRows(details: MissionControlNodeDetails): MissionControlInfoRow[] {
  const rows: MissionControlInfoRow[] = [];

  pushInfoRow(rows, "pages.missionControl.sidePanel.nodeId", details.node.node_public_id, { mono: true });
  pushInfoRow(rows, "pages.missionControl.sidePanel.rootId", details.node.root_node_public_id, { mono: true });
  pushInfoRow(rows, "pages.missionControl.sidePanel.projectionVersion", details.node.projection_version, { mono: true });
  pushInfoRow(rows, "pages.missionControl.sidePanel.providerRef", details.node.provider_reference?.external_id, { mono: true });
  pushInfoRow(rows, "pages.missionControl.sidePanel.lastActivity", details.node.last_activity_at, { mono: true });

  const payload = details.detail_payload;

  if (isDiscussionPayload(payload)) {
    pushInfoRow(rows, "pages.missionControl.sidePanel.discussionKind", payload.discussion_kind);
    pushInfoRow(rows, "pages.missionControl.sidePanel.status", payload.status);
    pushInfoRow(rows, "pages.missionControl.sidePanel.author", payload.author);
    pushInfoRow(rows, "pages.missionControl.sidePanel.participants", payload.participant_count);
  }

  if (isWorkItemPayload(payload)) {
    pushInfoRow(rows, "pages.missionControl.sidePanel.repository", payload.repository_full_name, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.issueNumber", `#${payload.issue_number}`, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.stageLabel", payload.stage_label, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.lastProviderSync", payload.last_provider_sync_at, { mono: true });
  }

  if (isRunPayload(payload)) {
    pushInfoRow(rows, "pages.missionControl.sidePanel.runId", payload.run_id, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.agentKey", payload.agent_key, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.runStatus", payload.run_status);
    pushInfoRow(rows, "pages.missionControl.sidePanel.runtimeMode", payload.runtime_mode, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.triggerLabel", payload.trigger_label, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.buildRef", payload.build_ref, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.candidateNamespace", payload.candidate_namespace, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.startedAt", payload.started_at, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.finishedAt", payload.finished_at, { mono: true });
  }

  if (isPullRequestPayload(payload)) {
    pushInfoRow(rows, "pages.missionControl.sidePanel.repository", payload.repository_full_name, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.pullRequestNumber", `#${payload.pull_request_number}`, {
      mono: true,
    });
    pushInfoRow(rows, "pages.missionControl.sidePanel.branchHead", payload.branch_head, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.branchBase", payload.branch_base, { mono: true });
    pushInfoRow(rows, "pages.missionControl.sidePanel.mergeState", payload.merge_state);
    pushInfoRow(rows, "pages.missionControl.sidePanel.reviewDecision", payload.review_decision);
    pushInfoRow(rows, "pages.missionControl.sidePanel.checksSummary", payload.checks_summary);
  }

  return rows;
}

export function missionControlRelatedNodeSections(details: MissionControlNodeDetails): MissionControlRelatedNodesSection[] {
  const payload = details.detail_payload;
  const sections: MissionControlRelatedNodesSection[] = [];

  if (isDiscussionPayload(payload)) {
    sections.push({
      titleKey: "pages.missionControl.relatedSections.formalizations",
      refs: payload.formalization_target_refs,
    });
  }

  if (isWorkItemPayload(payload)) {
    sections.push({
      titleKey: "pages.missionControl.relatedSections.linkedRuns",
      refs: payload.linked_run_refs,
    });
    sections.push({
      titleKey: "pages.missionControl.relatedSections.followUpIssues",
      refs: payload.linked_follow_up_refs,
    });
  }

  if (isRunPayload(payload)) {
    sections.push({
      titleKey: "pages.missionControl.relatedSections.linkedPullRequests",
      refs: payload.linked_pull_request_refs,
    });
    sections.push({
      titleKey: "pages.missionControl.relatedSections.producedIssues",
      refs: payload.produced_issue_refs,
    });
  }

  if (isPullRequestPayload(payload)) {
    sections.push({
      titleKey: "pages.missionControl.relatedSections.linkedIssues",
      refs: payload.linked_issue_refs,
    });
    sections.push({
      titleKey: "pages.missionControl.relatedSections.linkedRuns",
      refs: payload.linked_run_ref ? [payload.linked_run_ref] : [],
    });
  }

  return sections.filter((section) => section.refs.length > 0);
}

export function missionControlAdjacentNodeRefs(details: MissionControlNodeDetails): MissionControlNodeRef[] {
  return details.adjacent_nodes.map((node) => ({
    node_kind: node.node_kind,
    node_public_id: node.node_public_id,
  }));
}
