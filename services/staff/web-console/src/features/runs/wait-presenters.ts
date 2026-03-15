import type {
  GitHubRateLimitManualAction,
  GitHubRateLimitWaitItem,
  Run,
  RunRealtimeMessage,
  RunWaitItemView,
  RunWaitNextStepView,
  RunWaitProjection,
  RunWaitProjectionView,
  RunWaitQueueRowView,
  RunWaitRealtimeEntryView,
} from "./types";

const waitRealtimeKinds = new Set<RunRealtimeMessage["type"]>([
  "wait_entered",
  "wait_updated",
  "wait_resolved",
  "wait_manual_action_required",
]);

function contourLabelKey(value: GitHubRateLimitWaitItem["contour_kind"]): string {
  switch (value) {
    case "platform_pat":
      return "runs.waits.contours.platformPat";
    case "agent_bot_token":
      return "runs.waits.contours.agentBotToken";
  }
}

function contourColor(value: GitHubRateLimitWaitItem["contour_kind"]): string {
  switch (value) {
    case "platform_pat":
      return "warning";
    case "agent_bot_token":
      return "info";
  }
}

function limitLabelKey(value: GitHubRateLimitWaitItem["limit_kind"]): string {
  switch (value) {
    case "primary":
      return "runs.waits.limitKinds.primary";
    case "secondary":
      return "runs.waits.limitKinds.secondary";
  }
}

function limitColor(value: GitHubRateLimitWaitItem["limit_kind"]): string {
  switch (value) {
    case "primary":
      return "success";
    case "secondary":
      return "warning";
  }
}

function stateLabelKey(value: GitHubRateLimitWaitItem["state"]): string {
  switch (value) {
    case "open":
      return "runs.waits.states.open";
    case "auto_resume_scheduled":
      return "runs.waits.states.autoResumeScheduled";
    case "auto_resume_in_progress":
      return "runs.waits.states.autoResumeInProgress";
    case "resolved":
      return "runs.waits.states.resolved";
    case "manual_action_required":
      return "runs.waits.states.manualActionRequired";
    case "cancelled":
      return "runs.waits.states.cancelled";
  }
}

function stateColor(value: GitHubRateLimitWaitItem["state"]): string {
  switch (value) {
    case "resolved":
      return "success";
    case "manual_action_required":
      return "error";
    case "auto_resume_in_progress":
      return "info";
    case "cancelled":
      return "secondary";
    case "open":
    case "auto_resume_scheduled":
    default:
      return "warning";
  }
}

function confidenceLabelKey(value: GitHubRateLimitWaitItem["confidence"]): string {
  switch (value) {
    case "deterministic":
      return "runs.waits.confidence.deterministic";
    case "conservative":
      return "runs.waits.confidence.conservative";
    case "provider_uncertain":
      return "runs.waits.confidence.providerUncertain";
  }
}

function confidenceColor(value: GitHubRateLimitWaitItem["confidence"]): string {
  switch (value) {
    case "deterministic":
      return "success";
    case "conservative":
      return "warning";
    case "provider_uncertain":
      return "secondary";
  }
}

function operationLabelKey(value: GitHubRateLimitWaitItem["operation_class"]): string {
  switch (value) {
    case "run_status_comment":
      return "runs.waits.operationClasses.runStatusComment";
    case "issue_label_transition":
      return "runs.waits.operationClasses.issueLabelTransition";
    case "repository_provider_call":
      return "runs.waits.operationClasses.repositoryProviderCall";
    case "agent_github_call":
      return "runs.waits.operationClasses.agentGithubCall";
  }
}

function recoveryHintLabelKey(value: GitHubRateLimitWaitItem["recovery_hint"]["hint_kind"]): string {
  switch (value) {
    case "rate_limit_reset":
      return "runs.waits.hints.rateLimitReset";
    case "retry_after":
      return "runs.waits.hints.retryAfter";
    case "exponential_backoff":
      return "runs.waits.hints.exponentialBackoff";
    case "manual_only":
      return "runs.waits.hints.manualOnly";
  }
}

function recoveryHintColor(value: GitHubRateLimitWaitItem["recovery_hint"]["hint_kind"]): string {
  switch (value) {
    case "rate_limit_reset":
      return "success";
    case "retry_after":
      return "info";
    case "exponential_backoff":
      return "warning";
    case "manual_only":
      return "error";
  }
}

function recoveryHintSourceLabelKey(value: GitHubRateLimitWaitItem["recovery_hint"]["source_headers"]): string {
  switch (value) {
    case "reset_at":
      return "runs.waits.hintSources.resetAt";
    case "retry_after":
      return "runs.waits.hintSources.retryAfter";
    case "provider_uncertain":
      return "runs.waits.hintSources.providerUncertain";
  }
}

function manualActionLabelKey(value: GitHubRateLimitManualAction["kind"]): string {
  switch (value) {
    case "requeue_platform_operation":
      return "runs.waits.manualActions.requeuePlatformOperation";
    case "resume_agent_session":
      return "runs.waits.manualActions.resumeAgentSession";
    case "retry_after_operator_review":
      return "runs.waits.manualActions.retryAfterOperatorReview";
  }
}

function commentMirrorLabelKey(value: RunWaitProjection["comment_mirror_state"]): string {
  switch (value) {
    case "synced":
      return "runs.waits.commentMirror.synced";
    case "pending_retry":
      return "runs.waits.commentMirror.pendingRetry";
    case "not_attempted":
      return "runs.waits.commentMirror.notAttempted";
  }
}

function commentMirrorColor(value: RunWaitProjection["comment_mirror_state"]): string {
  switch (value) {
    case "synced":
      return "success";
    case "pending_retry":
      return "warning";
    case "not_attempted":
      return "secondary";
  }
}

function resolutionLabelKey(value: NonNullable<RunRealtimeMessage["wait_resolution"]>["resolution_kind"]): string {
  switch (value) {
    case "auto_resumed":
      return "runs.waits.realtime.resolutions.autoResumed";
    case "manually_resolved":
      return "runs.waits.realtime.resolutions.manuallyResolved";
    case "cancelled":
      return "runs.waits.realtime.resolutions.cancelled";
  }
}

function resolutionColor(value: NonNullable<RunRealtimeMessage["wait_resolution"]>["resolution_kind"]): string {
  switch (value) {
    case "auto_resumed":
      return "success";
    case "manually_resolved":
      return "info";
    case "cancelled":
      return "secondary";
  }
}

function waitRealtimeLabelKey(kind: RunWaitRealtimeEntryView["kind"]): string {
  switch (kind) {
    case "wait_entered":
      return "runs.waits.realtime.waitEntered";
    case "wait_updated":
      return "runs.waits.realtime.waitUpdated";
    case "wait_resolved":
      return "runs.waits.realtime.waitResolved";
    case "wait_manual_action_required":
      return "runs.waits.realtime.waitManualActionRequired";
  }
}

function waitRealtimeIcon(kind: RunWaitRealtimeEntryView["kind"]): string {
  switch (kind) {
    case "wait_entered":
      return "mdi-timer-sand";
    case "wait_updated":
      return "mdi-refresh";
    case "wait_resolved":
      return "mdi-check-circle-outline";
    case "wait_manual_action_required":
      return "mdi-alert-octagon-outline";
  }
}

function waitRealtimeColor(kind: RunWaitRealtimeEntryView["kind"]): string {
  switch (kind) {
    case "wait_entered":
      return "warning";
    case "wait_updated":
      return "info";
    case "wait_resolved":
      return "success";
    case "wait_manual_action_required":
      return "error";
  }
}

function buildWaitItemView(item: GitHubRateLimitWaitItem): RunWaitItemView {
  return {
    waitId: item.wait_id,
    contourLabelKey: contourLabelKey(item.contour_kind),
    contourColor: contourColor(item.contour_kind),
    limitLabelKey: limitLabelKey(item.limit_kind),
    limitColor: limitColor(item.limit_kind),
    stateLabelKey: stateLabelKey(item.state),
    stateColor: stateColor(item.state),
    confidenceLabelKey: confidenceLabelKey(item.confidence),
    confidenceColor: confidenceColor(item.confidence),
    operationLabelKey: operationLabelKey(item.operation_class),
    enteredAt: item.entered_at,
    resumeNotBefore: item.resume_not_before ?? item.recovery_hint.resume_not_before ?? null,
    attemptsUsed: item.attempts_used,
    maxAttempts: item.max_attempts,
    recoveryHint: {
      kind: item.recovery_hint.hint_kind,
      kindLabelKey: recoveryHintLabelKey(item.recovery_hint.hint_kind),
      kindColor: recoveryHintColor(item.recovery_hint.hint_kind),
      sourceLabelKey: recoveryHintSourceLabelKey(item.recovery_hint.source_headers),
      detailsMarkdown: item.recovery_hint.details_markdown,
      resumeNotBefore: item.recovery_hint.resume_not_before ?? item.resume_not_before ?? null,
    },
    manualAction: item.manual_action
      ? {
          kind: item.manual_action.kind,
          kindLabelKey: manualActionLabelKey(item.manual_action.kind),
          summary: item.manual_action.summary,
          detailsMarkdown: item.manual_action.details_markdown,
          suggestedNotBefore: item.manual_action.suggested_not_before ?? null,
        }
      : null,
  };
}

export function buildRunWaitProjectionView(projection: RunWaitProjection | null | undefined): RunWaitProjectionView | null {
  if (!projection) {
    return null;
  }

  return {
    waitStateLabelKey: "runs.waits.waitStates.waitingBackpressure",
    waitReasonLabelKey: "runs.waits.waitReasons.githubRateLimit",
    commentMirror: {
      labelKey: commentMirrorLabelKey(projection.comment_mirror_state),
      color: commentMirrorColor(projection.comment_mirror_state),
    },
    dominantWait: buildWaitItemView(projection.dominant_wait),
    relatedWaits: projection.related_waits.map(buildWaitItemView),
  };
}

export function buildRunWaitQueueRow(run: Run): RunWaitQueueRowView {
  const projectLabel = String(run.project_name || run.project_slug || run.project_id || run.id);

  return {
    runId: run.id,
    projectId: run.project_id,
    projectLabel,
    status: run.status,
    triggerKind: String(run.trigger_kind || ""),
    agentKey: String(run.agent_key || ""),
    waitState: run.wait_state ?? null,
    waitReason: run.wait_reason ?? null,
    waitSince: run.wait_since ?? null,
    projection: buildRunWaitProjectionView(run.wait_projection),
  };
}

export function buildRunWaitNextStepView(item: RunWaitItemView): RunWaitNextStepView {
  if (item.manualAction) {
    return {
      labelKey: item.manualAction.kindLabelKey,
      color: "error",
      summary: item.manualAction.summary,
      detailsMarkdown: item.manualAction.detailsMarkdown,
      scheduledAt: item.manualAction.suggestedNotBefore,
      scheduledAtLabelKey: "pages.runDetails.suggestedNotBefore",
    };
  }

  return {
    labelKey: item.recoveryHint.kindLabelKey,
    color: item.recoveryHint.kindColor,
    summary: item.recoveryHint.detailsMarkdown,
    detailsMarkdown: item.recoveryHint.detailsMarkdown,
    scheduledAt: item.recoveryHint.resumeNotBefore,
    scheduledAtLabelKey: "pages.runDetails.resumeNotBefore",
  };
}

export function isRunWaitRealtimeMessage(message: RunRealtimeMessage): boolean {
  return waitRealtimeKinds.has(message.type);
}

export function buildRunWaitRealtimeEntryView(message: RunRealtimeMessage): RunWaitRealtimeEntryView | null {
  if (!isRunWaitRealtimeMessage(message)) {
    return null;
  }

  if ((message.type === "wait_entered" || message.type === "wait_updated") && message.wait_projection) {
    const dominantWait = buildWaitItemView(message.wait_projection.dominant_wait);
    return {
      id: `${message.type}:${message.sent_at}:${dominantWait.waitId}`,
      kind: message.type,
      labelKey: waitRealtimeLabelKey(message.type),
      color: waitRealtimeColor(message.type),
      icon: waitRealtimeIcon(message.type),
      occurredAt: message.sent_at,
      waitId: dominantWait.waitId,
      contourLabelKey: dominantWait.contourLabelKey,
      limitLabelKey: dominantWait.limitLabelKey,
      manualActionLabelKey: dominantWait.manualAction?.kindLabelKey,
      manualActionSummary: dominantWait.manualAction?.summary,
      detailsMarkdown: dominantWait.manualAction?.detailsMarkdown || dominantWait.recoveryHint.detailsMarkdown,
    };
  }

  if (message.type === "wait_resolved" && message.wait_resolution) {
    return {
      id: `${message.type}:${message.sent_at}:${message.wait_resolution.wait_id}`,
      kind: message.type,
      labelKey: waitRealtimeLabelKey(message.type),
      color: resolutionColor(message.wait_resolution.resolution_kind),
      icon: waitRealtimeIcon(message.type),
      occurredAt: message.wait_resolution.resolved_at || message.sent_at,
      waitId: message.wait_resolution.wait_id,
      contourLabelKey: contourLabelKey(message.wait_resolution.contour_kind),
      resolutionLabelKey: resolutionLabelKey(message.wait_resolution.resolution_kind),
    };
  }

  if (message.type === "wait_manual_action_required" && message.wait_manual_action) {
    return {
      id: `${message.type}:${message.sent_at}:${message.wait_manual_action.wait_id}`,
      kind: message.type,
      labelKey: waitRealtimeLabelKey(message.type),
      color: waitRealtimeColor(message.type),
      icon: waitRealtimeIcon(message.type),
      occurredAt: message.wait_manual_action.updated_at || message.sent_at,
      waitId: message.wait_manual_action.wait_id,
      manualActionLabelKey: manualActionLabelKey(message.wait_manual_action.manual_action.kind),
      manualActionSummary: message.wait_manual_action.manual_action.summary,
      detailsMarkdown: message.wait_manual_action.manual_action.details_markdown,
    };
  }

  return null;
}
