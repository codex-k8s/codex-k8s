import test from "node:test";
import assert from "node:assert/strict";

import { buildRunTimelinePhases, buildRunTimelineStatuses } from "./run-timeline.ts";

const timelineTranslations = {
  "runs.timeline.events.workerHeartbeatMissed": "Worker heartbeat missed",
  "runs.timeline.events.workerHeartbeatMissedWithActor": "Worker heartbeat missed · {actorId}",
  "runs.timeline.events.staleLeaseDetected": "Stale lease detected",
  "runs.timeline.events.staleLeaseDetectedWithActor": "Stale lease detected · {actorId}",
  "runs.timeline.events.staleLeaseReleased": "Stale lease released",
  "runs.timeline.events.staleLeaseReleasedWithActor": "Stale lease released · {actorId}",
  "runs.timeline.events.leaseReclaimedAfterStale": "Lease reclaimed after stale lease",
  "runs.timeline.events.leaseReclaimedAfterStaleWithActor": "Lease reclaimed after stale lease · {actorId}",
} as const;

function translateTimelineStatus(key: string, params?: Record<string, string>): string {
  const template = timelineTranslations[key as keyof typeof timelineTranslations] || key;
  return template.replaceAll("{actorId}", params?.actorId ?? "");
}

test("buildRunTimelinePhases adds build-deploy, auth and ready steps for full-env runs", () => {
  const run = {
    created_at: "2026-03-12T10:00:00Z",
    started_at: null,
    finished_at: null,
    wait_state: null,
    wait_since: null,
    status: "running",
    trigger_kind: "dev",
    trigger_label: "run:dev",
  } as const;

  const events = [
    {
      event_type: "run.agent.ready",
      created_at: "2026-03-12T10:04:00Z",
      payload_json: "{\"runtime_mode\":\"full-env\"}",
    },
    {
      event_type: "run.codex.auth.synchronized",
      created_at: "2026-03-12T10:03:30Z",
      payload_json: "{\"source\":\"device_auth\"}",
    },
    {
      event_type: "run.codex.auth.required",
      created_at: "2026-03-12T10:03:00Z",
      payload_json: "{}",
    },
    {
      event_type: "run.started",
      created_at: "2026-03-12T10:02:00Z",
      payload_json: "{\"runtime_mode\":\"full-env\"}",
    },
  ];

  const steps = buildRunTimelinePhases(run as never, events as never, "ru", new Date("2026-03-12T12:00:00Z"));
  assert.deepEqual(steps.map((item) => item.key), ["created", "buildDeploy", "started", "authResolved", "agentReady"]);
  assert.equal(steps[1]?.showSpinner, false);
  assert.equal(steps[4]?.atLabel, "10:04");
});

test("buildRunTimelinePhases omits build-deploy for code-only runs", () => {
  const run = {
    created_at: "2026-03-12T10:00:00Z",
    started_at: null,
    finished_at: null,
    wait_state: null,
    wait_since: null,
    status: "running",
    trigger_kind: "ai_repair",
    trigger_label: "run:ai-repair",
  } as const;

  const events = [
    {
      event_type: "run.started",
      created_at: "2026-03-12T10:01:00Z",
      payload_json: "{\"runtime_mode\":\"code-only\"}",
    },
  ];

  const steps = buildRunTimelinePhases(run as never, events as never, "en", new Date("2026-03-12T12:00:00Z"));
  assert.deepEqual(steps.map((item) => item.key), ["created", "started", "agentReady"]);
});

test("buildRunTimelineStatuses collapses adjacent duplicates and formats compact timestamps", () => {
  const events = [
    {
      event_type: "run.agent.status_reported",
      created_at: "2026-03-12T10:06:00Z",
      payload_json: "{\"status_text\":\"Проверяю тесты\",\"agent_key\":\"dev\"}",
    },
    {
      event_type: "run.agent.status_reported",
      created_at: "2026-03-12T10:05:00Z",
      payload_json: "{\"status_text\":\"Проверяю тесты\",\"agent_key\":\"dev\"}",
    },
    {
      event_type: "run.agent.status_reported",
      created_at: "2026-03-11T16:00:00Z",
      payload_json: "{\"status_text\":\"Обновляю API\",\"agent_key\":\"dev\"}",
    },
  ];

  const entries = buildRunTimelineStatuses(
    events as never,
    "ru",
    translateTimelineStatus,
    new Date("2026-03-12T12:00:00Z"),
  );
  assert.equal(entries.length, 2);
  assert.equal(entries[0]?.at, "2026-03-12T10:06:00Z");
  assert.equal(entries[0]?.repeatCount, 2);
  assert.equal(entries[0]?.timeLabel, "10:06");
  assert.equal(entries[1]?.timeLabel, "11 мар 16:00");
});

test("buildRunTimelineStatuses includes stale lease recovery events in timeline", () => {
  const events = [
    {
      event_type: "run.reclaimed_after_stale_lease",
      created_at: "2026-03-12T10:08:00Z",
      payload_json: "{\"worker_id\":\"worker-2\"}",
    },
    {
      event_type: "run.lease.released",
      created_at: "2026-03-12T10:07:30Z",
      payload_json: "{\"previous_lease_owner\":\"worker-old\"}",
    },
    {
      event_type: "run.lease.detected_stale",
      created_at: "2026-03-12T10:07:00Z",
      payload_json: "{\"previous_lease_owner\":\"worker-old\"}",
    },
    {
      event_type: "worker.instance.heartbeat.missed",
      created_at: "2026-03-12T10:06:30Z",
      payload_json: "{\"worker_id\":\"worker-old\"}",
    },
  ];

  const entries = buildRunTimelineStatuses(
    events as never,
    "en",
    translateTimelineStatus,
    new Date("2026-03-12T12:00:00Z"),
  );
  assert.deepEqual(
    entries.map((item) => item.text),
    [
      "Lease reclaimed after stale lease · worker-2",
      "Stale lease released · worker-old",
      "Stale lease detected · worker-old",
      "Worker heartbeat missed · worker-old",
    ],
  );
});

test("buildRunTimelineStatuses uses event translation without actor placeholder leakage", () => {
  const events = [
    {
      event_type: "run.lease.detected_stale",
      created_at: "2026-03-12T10:07:00Z",
      payload_json: "{}",
    },
  ];

  const entries = buildRunTimelineStatuses(
    events as never,
    "en",
    translateTimelineStatus,
    new Date("2026-03-12T12:00:00Z"),
  );
  assert.deepEqual(entries.map((item) => item.text), ["Stale lease detected"]);
});
