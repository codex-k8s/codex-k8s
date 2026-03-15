import { createRealtimeClient, type RealtimeConnectionState } from "../../shared/ws/realtime-client.ts";
import type {
  GitHubRateLimitManualAction,
  GitHubRateLimitRecoveryHint,
  GitHubRateLimitWaitItem,
  RunRealtimeMessage,
  RunRealtimeMessageType,
  RunWaitManualActionEvent,
  RunWaitProjection,
  RunWaitResolution,
} from "./types";

export type RunRealtimeState = Exclude<RealtimeConnectionState, "closed">;

type SubscribeRunRealtimeParams = {
  runId: string;
  onMessage: (message: RunRealtimeMessage) => void;
  onStateChange?: (state: RunRealtimeState) => void;
  includeLogs?: boolean;
  eventsLimit?: number;
  tailLines?: number;
};

const realtimeMessageTypes = new Set<RunRealtimeMessageType>([
  "snapshot",
  "run",
  "events",
  "logs",
  "wait_entered",
  "wait_updated",
  "wait_resolved",
  "wait_manual_action_required",
  "error",
]);

const contourKinds = new Set<GitHubRateLimitWaitItem["contour_kind"]>(["platform_pat", "agent_bot_token"]);
const limitKinds = new Set<GitHubRateLimitWaitItem["limit_kind"]>(["primary", "secondary"]);
const operationClasses = new Set<GitHubRateLimitWaitItem["operation_class"]>([
  "run_status_comment",
  "issue_label_transition",
  "repository_provider_call",
  "agent_github_call",
]);
const waitStates = new Set<GitHubRateLimitWaitItem["state"]>([
  "open",
  "auto_resume_scheduled",
  "auto_resume_in_progress",
  "resolved",
  "manual_action_required",
  "cancelled",
]);
const confidenceKinds = new Set<GitHubRateLimitWaitItem["confidence"]>([
  "deterministic",
  "conservative",
  "provider_uncertain",
]);
const recoveryHintKinds = new Set<GitHubRateLimitRecoveryHint["hint_kind"]>([
  "rate_limit_reset",
  "retry_after",
  "exponential_backoff",
  "manual_only",
]);
const recoveryHintSources = new Set<GitHubRateLimitRecoveryHint["source_headers"]>([
  "reset_at",
  "retry_after",
  "provider_uncertain",
]);
const manualActionKinds = new Set<GitHubRateLimitManualAction["kind"]>([
  "requeue_platform_operation",
  "resume_agent_session",
  "retry_after_operator_review",
]);
const commentMirrorStates = new Set<RunWaitProjection["comment_mirror_state"]>([
  "synced",
  "pending_retry",
  "not_attempted",
]);
const resolutionKinds = new Set<RunWaitResolution["resolution_kind"]>([
  "auto_resumed",
  "manually_resolved",
  "cancelled",
]);

function isObjectRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isOptionalString(value: unknown): value is string | null | undefined {
  return value === undefined || value === null || typeof value === "string";
}

function isGitHubRateLimitManualAction(value: unknown): value is GitHubRateLimitManualAction {
  if (!isObjectRecord(value)) {
    return false;
  }

  return (
    manualActionKinds.has(value.kind as GitHubRateLimitManualAction["kind"]) &&
    typeof value.summary === "string" &&
    typeof value.details_markdown === "string" &&
    isOptionalString(value.suggested_not_before)
  );
}

function isGitHubRateLimitRecoveryHint(value: unknown): value is GitHubRateLimitRecoveryHint {
  if (!isObjectRecord(value)) {
    return false;
  }

  return (
    recoveryHintKinds.has(value.hint_kind as GitHubRateLimitRecoveryHint["hint_kind"]) &&
    recoveryHintSources.has(value.source_headers as GitHubRateLimitRecoveryHint["source_headers"]) &&
    typeof value.details_markdown === "string" &&
    isOptionalString(value.resume_not_before)
  );
}

function isGitHubRateLimitWaitItem(value: unknown): value is GitHubRateLimitWaitItem {
  if (!isObjectRecord(value)) {
    return false;
  }

  return (
    typeof value.wait_id === "string" &&
    contourKinds.has(value.contour_kind as GitHubRateLimitWaitItem["contour_kind"]) &&
    limitKinds.has(value.limit_kind as GitHubRateLimitWaitItem["limit_kind"]) &&
    operationClasses.has(value.operation_class as GitHubRateLimitWaitItem["operation_class"]) &&
    waitStates.has(value.state as GitHubRateLimitWaitItem["state"]) &&
    confidenceKinds.has(value.confidence as GitHubRateLimitWaitItem["confidence"]) &&
    typeof value.entered_at === "string" &&
    isOptionalString(value.resume_not_before) &&
    typeof value.attempts_used === "number" &&
    Number.isFinite(value.attempts_used) &&
    typeof value.max_attempts === "number" &&
    Number.isFinite(value.max_attempts) &&
    isGitHubRateLimitRecoveryHint(value.recovery_hint) &&
    (value.manual_action === undefined || isGitHubRateLimitManualAction(value.manual_action))
  );
}

function isRunWaitProjection(value: unknown): value is RunWaitProjection {
  if (!isObjectRecord(value)) {
    return false;
  }

  return (
    typeof value.wait_state === "string" &&
    typeof value.wait_reason === "string" &&
    commentMirrorStates.has(value.comment_mirror_state as RunWaitProjection["comment_mirror_state"]) &&
    isGitHubRateLimitWaitItem(value.dominant_wait) &&
    Array.isArray(value.related_waits) &&
    value.related_waits.every(isGitHubRateLimitWaitItem)
  );
}

function isRunWaitResolution(value: unknown): value is RunWaitResolution {
  if (!isObjectRecord(value)) {
    return false;
  }

  return (
    typeof value.wait_id === "string" &&
    contourKinds.has(value.contour_kind as RunWaitResolution["contour_kind"]) &&
    resolutionKinds.has(value.resolution_kind as RunWaitResolution["resolution_kind"]) &&
    typeof value.resolved_at === "string"
  );
}

function isRunWaitManualActionEvent(value: unknown): value is RunWaitManualActionEvent {
  if (!isObjectRecord(value)) {
    return false;
  }

  return (
    typeof value.wait_id === "string" &&
    typeof value.updated_at === "string" &&
    isGitHubRateLimitManualAction(value.manual_action)
  );
}

function buildRunRealtimeURL(params: { runId: string; includeLogs: boolean; eventsLimit: number; tailLines: number }): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const url = new URL(`${protocol}//${window.location.host}/api/v1/staff/runs/${encodeURIComponent(params.runId)}/realtime`);
  url.searchParams.set("limit", String(params.eventsLimit));
  url.searchParams.set("tail_lines", String(params.tailLines));
  if (params.includeLogs) {
    url.searchParams.set("include_logs", "true");
  }
  return url.toString();
}

export function parseRunRealtimeMessage(raw: string): RunRealtimeMessage | null {
  const text = String(raw || "").trim();
  if (!text) return null;
  try {
    const payload = JSON.parse(text) as Partial<RunRealtimeMessage>;
    if (!payload || typeof payload !== "object") return null;
    const type = String(payload.type || "") as RunRealtimeMessageType;
    if (!realtimeMessageTypes.has(type)) return null;
    return {
      type,
      run: payload.run,
      sent_at: String(payload.sent_at || new Date().toISOString()),
      events: Array.isArray(payload.events) ? payload.events : undefined,
      logs: payload.logs,
      wait_projection: isRunWaitProjection(payload.wait_projection) ? payload.wait_projection : undefined,
      wait_resolution: isRunWaitResolution(payload.wait_resolution) ? payload.wait_resolution : undefined,
      wait_manual_action: isRunWaitManualActionEvent(payload.wait_manual_action) ? payload.wait_manual_action : undefined,
      message: typeof payload.message === "string" ? payload.message : undefined,
    } as RunRealtimeMessage;
  } catch {
    return null;
  }
}

export function subscribeRunRealtime(params: SubscribeRunRealtimeParams): () => void {
  const runId = String(params.runId || "").trim();
  if (!runId) {
    return () => undefined;
  }

  const url = buildRunRealtimeURL({
    runId,
    includeLogs: Boolean(params.includeLogs),
    eventsLimit: Number(params.eventsLimit || 200),
    tailLines: Number(params.tailLines || 200),
  });

  const client = createRealtimeClient<RunRealtimeMessage>({
    url,
    parseMessage: parseRunRealtimeMessage,
    onMessage: params.onMessage,
    onStateChange: (state) => {
      if (state === "closed") return;
      params.onStateChange?.(state);
    },
  });

  client.start();
  return () => client.stop();
}
