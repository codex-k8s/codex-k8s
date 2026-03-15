import test from "node:test";
import assert from "node:assert/strict";

import { parseRunRealtimeMessage } from "../../features/runs/realtime.ts";
import {
  buildRunWaitNextStepView,
  buildRunWaitProjectionView,
  buildRunWaitRealtimeEntryView,
  buildRunWaitQueueRow,
} from "../../features/runs/wait-presenters.ts";
import type { Run, RunWaitProjection } from "../../features/runs/types";

function createProjection(): RunWaitProjection {
  return {
    wait_state: "waiting_backpressure",
    wait_reason: "github_rate_limit",
    comment_mirror_state: "pending_retry",
    dominant_wait: {
      wait_id: "wait-dominant",
      contour_kind: "platform_pat",
      limit_kind: "secondary",
      operation_class: "run_status_comment",
      state: "manual_action_required",
      confidence: "conservative",
      entered_at: "2026-03-14T19:00:00Z",
      resume_not_before: "2026-03-14T19:10:00Z",
      attempts_used: 2,
      max_attempts: 3,
      recovery_hint: {
        hint_kind: "retry_after",
        resume_not_before: "2026-03-14T19:10:00Z",
        source_headers: "retry_after",
        details_markdown: "Retry after the provider retry window expires.",
      },
      manual_action: {
        kind: "resume_agent_session",
        summary: "Resume the persisted agent session",
        details_markdown: "Check the latest contour evidence, then resume the session once.",
        suggested_not_before: "2026-03-14T19:15:00Z",
      },
    },
    related_waits: [
      {
        wait_id: "wait-related",
        contour_kind: "agent_bot_token",
        limit_kind: "primary",
        operation_class: "agent_github_call",
        state: "auto_resume_scheduled",
        confidence: "deterministic",
        entered_at: "2026-03-14T19:01:00Z",
        resume_not_before: "2026-03-14T19:06:00Z",
        attempts_used: 1,
        max_attempts: 2,
        recovery_hint: {
          hint_kind: "rate_limit_reset",
          resume_not_before: "2026-03-14T19:06:00Z",
          source_headers: "reset_at",
          details_markdown: "Wait for GitHub primary reset.",
        },
      },
    ],
  };
}

test("buildRunWaitProjectionView maps dominant and related waits into UI view models", () => {
  const view = buildRunWaitProjectionView(createProjection());

  assert.ok(view);
  assert.equal(view.waitStateLabelKey, "runs.waits.waitStates.waitingBackpressure");
  assert.equal(view.commentMirror.labelKey, "runs.waits.commentMirror.pendingRetry");
  assert.equal(view.dominantWait.contourLabelKey, "runs.waits.contours.platformPat");
  assert.equal(view.dominantWait.manualAction?.kindLabelKey, "runs.waits.manualActions.resumeAgentSession");
  assert.equal(view.relatedWaits[0]?.limitLabelKey, "runs.waits.limitKinds.primary");
});

test("buildRunWaitQueueRow keeps typed projection for wait queue surfaces", () => {
  const row = buildRunWaitQueueRow({
    id: "run-1",
    correlation_id: "corr-1",
    project_id: "proj-1",
    project_slug: "proj",
    project_name: "Project",
    status: "waiting_backpressure",
    wait_state: "waiting_backpressure",
    wait_reason: "github_rate_limit",
    wait_since: "2026-03-14T19:00:00Z",
    trigger_kind: "dev",
    agent_key: "dev",
    created_at: "2026-03-14T18:50:00Z",
    wait_projection: createProjection(),
  } as Run);

  assert.equal(row.projectLabel, "Project");
  assert.equal(row.projection?.dominantWait.operationLabelKey, "runs.waits.operationClasses.runStatusComment");
});

test("buildRunWaitNextStepView prioritizes manual action guidance", () => {
  const projection = buildRunWaitProjectionView(createProjection());
  assert.ok(projection);

  const nextStep = buildRunWaitNextStepView(projection.dominantWait);

  assert.equal(nextStep.labelKey, "runs.waits.manualActions.resumeAgentSession");
  assert.equal(nextStep.scheduledAtLabelKey, "pages.runDetails.suggestedNotBefore");
  assert.equal(nextStep.scheduledAt, "2026-03-14T19:15:00Z");
});

test("parseRunRealtimeMessage accepts wait update envelopes and presenters build realtime entry", () => {
  const message = parseRunRealtimeMessage(
    JSON.stringify({
      type: "wait_updated",
      sent_at: "2026-03-14T19:05:00Z",
      wait_projection: createProjection(),
    }),
  );

  assert.ok(message);
  assert.equal(message.type, "wait_updated");

  const entry = buildRunWaitRealtimeEntryView(message);
  assert.ok(entry);
  assert.equal(entry.labelKey, "runs.waits.realtime.waitUpdated");
  assert.equal(entry.waitId, "wait-dominant");
  assert.equal(entry.manualActionLabelKey, "runs.waits.manualActions.resumeAgentSession");
});

test("buildRunWaitRealtimeEntryView maps wait resolution envelopes", () => {
  const message = parseRunRealtimeMessage(
    JSON.stringify({
      type: "wait_resolved",
      sent_at: "2026-03-14T19:20:00Z",
      wait_resolution: {
        wait_id: "wait-dominant",
        contour_kind: "platform_pat",
        resolution_kind: "auto_resumed",
        resolved_at: "2026-03-14T19:19:30Z",
      },
    }),
  );

  assert.ok(message);

  const entry = buildRunWaitRealtimeEntryView(message);
  assert.ok(entry);
  assert.equal(entry.resolutionLabelKey, "runs.waits.realtime.resolutions.autoResumed");
  assert.equal(entry.occurredAt, "2026-03-14T19:19:30Z");
});
