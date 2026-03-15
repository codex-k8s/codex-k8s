import type {
  ApprovalRequest,
  FlowEvent,
  GitHubRateLimitManualAction,
  GitHubRateLimitRecoveryHint,
  GitHubRateLimitWaitItem,
  ResolveApprovalDecisionResponse,
  Run,
  RunLogs,
  RunNamespaceCleanupResponse,
  RunRealtimeMessage as GeneratedRunRealtimeMessage,
  RunWaitManualActionEvent,
  RunWaitProjection,
  RunWaitResolution,
} from "../../shared/api/generated";

export type {
  ApprovalRequest,
  FlowEvent,
  GitHubRateLimitManualAction,
  GitHubRateLimitRecoveryHint,
  GitHubRateLimitWaitItem,
  ResolveApprovalDecisionResponse,
  Run,
  RunLogs,
  RunNamespaceCleanupResponse,
  RunWaitManualActionEvent,
  RunWaitProjection,
  RunWaitResolution,
} from "../../shared/api/generated";

export type RunRealtimeMessageType = GeneratedRunRealtimeMessage["type"];

export type RunRealtimeMessage = GeneratedRunRealtimeMessage;

export type RunWaitRealtimeMessageType = Extract<
  RunRealtimeMessageType,
  "wait_entered" | "wait_updated" | "wait_resolved" | "wait_manual_action_required"
>;

export type RunWaitManualActionView = {
  kind: GitHubRateLimitManualAction["kind"];
  kindLabelKey: string;
  summary: string;
  detailsMarkdown: string;
  suggestedNotBefore: string | null;
};

export type RunWaitRecoveryHintView = {
  kind: GitHubRateLimitRecoveryHint["hint_kind"];
  kindLabelKey: string;
  kindColor: string;
  sourceLabelKey: string;
  detailsMarkdown: string;
  resumeNotBefore: string | null;
};

export type RunWaitItemView = {
  waitId: string;
  contourLabelKey: string;
  contourColor: string;
  limitLabelKey: string;
  limitColor: string;
  stateLabelKey: string;
  stateColor: string;
  confidenceLabelKey: string;
  confidenceColor: string;
  operationLabelKey: string;
  enteredAt: string;
  resumeNotBefore: string | null;
  attemptsUsed: number;
  maxAttempts: number;
  recoveryHint: RunWaitRecoveryHintView;
  manualAction: RunWaitManualActionView | null;
};

export type RunWaitCommentMirrorView = {
  labelKey: string;
  color: string;
};

export type RunWaitProjectionView = {
  waitStateLabelKey: string;
  waitReasonLabelKey: string;
  commentMirror: RunWaitCommentMirrorView;
  dominantWait: RunWaitItemView;
  relatedWaits: RunWaitItemView[];
};

export type RunWaitQueueRowView = {
  runId: string;
  projectId?: string | null;
  projectLabel: string;
  status: string;
  triggerKind: string;
  agentKey: string;
  waitState: string | null;
  waitReason: string | null;
  waitSince: string | null;
  projection: RunWaitProjectionView | null;
};

export type RunWaitNextStepView = {
  labelKey: string;
  color: string;
  summary: string;
  detailsMarkdown: string;
  scheduledAt: string | null;
  scheduledAtLabelKey: string | null;
};

export type RunWaitRealtimeEntryView = {
  id: string;
  kind: RunWaitRealtimeMessageType;
  labelKey: string;
  color: string;
  icon: string;
  occurredAt: string;
  waitId: string;
  contourLabelKey?: string;
  limitLabelKey?: string;
  resolutionLabelKey?: string;
  manualActionLabelKey?: string;
  manualActionSummary?: string;
  detailsMarkdown?: string;
};

export type RealtimePagination = {
  page: number;
  page_size: number;
  total_count: number;
};

export type RunsRealtimeMessageType = "snapshot" | "error";

export type RunsRealtimeMessage = {
  type: RunsRealtimeMessageType;
  items?: Run[];
  pagination?: RealtimePagination;
  wait_queue_count?: number;
  pending_approvals_count?: number;
  message?: string;
  sent_at: string;
};
