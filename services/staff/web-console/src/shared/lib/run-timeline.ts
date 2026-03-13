import { formatCompactDateTime } from "./datetime.ts";
import type { FlowEvent, Run } from "../../features/runs/types";

export type TimelinePhaseStepKey =
  | "created"
  | "buildDeploy"
  | "started"
  | "authResolved"
  | "agentReady"
  | "waiting"
  | "finished";

export type TimelineSubtitleKind = "waitState" | "status" | "buildFailed";

export type TimelinePhaseStep = {
  key: TimelinePhaseStepKey;
  at: string | null;
  atLabel: string | null;
  subtitleKind?: TimelineSubtitleKind;
  subtitleValue?: string;
  color: string;
  icon: string;
  showSpinner?: boolean;
};

export type TimelineStatusEntry = {
  key: string;
  at: string | null;
  text: string;
  timeLabel: string;
  repeatCount: number;
};

type EventPayload = Record<string, unknown>;

const runtimeModeFullEnv = "full-env";
const runtimeModeCodeOnly = "code-only";

function parsePayload(raw: string): EventPayload | null {
  const value = String(raw || "").trim();
  if (!value) return null;
  try {
    const parsed = JSON.parse(value) as unknown;
    if (parsed && typeof parsed === "object") return parsed as EventPayload;
    if (typeof parsed === "string") {
      const nested = JSON.parse(parsed) as unknown;
      if (nested && typeof nested === "object") return nested as EventPayload;
    }
  } catch {
    return null;
  }
  return null;
}

function findEventAt(events: FlowEvent[], eventType: string): string | null {
  for (const eventItem of events) {
    if (eventItem.event_type === eventType) {
      return eventItem.created_at || null;
    }
  }
  return null;
}

function findAuthResolvedAt(events: FlowEvent[]): string | null {
  for (const eventItem of events) {
    if (eventItem.event_type !== "run.codex.auth.synchronized") continue;
    const payload = parsePayload(eventItem.payload_json || "");
    if (String(payload?.source || "").trim() === "device_auth") {
      return eventItem.created_at || null;
    }
  }
  return null;
}

function recoveryStatusText(eventType: string, payload: EventPayload | null): string | null {
  const workerId = typeof payload?.worker_id === "string" ? payload.worker_id.trim() : "";
  const ownerId = typeof payload?.previous_lease_owner === "string" ? payload.previous_lease_owner.trim() : "";
  const actorId = workerId || ownerId;

  switch (eventType) {
    case "worker.instance.heartbeat.missed":
    case "run.lease.detected_stale":
    case "run.lease.released":
      return actorId ? `${eventType} · ${actorId}` : eventType;
    case "run.reclaimed_after_stale_lease":
      return actorId ? `${eventType} · ${actorId}` : eventType;
    default:
      return null;
  }
}

function resolveRuntimeMode(events: FlowEvent[], run: Run | null): string {
  for (const eventItem of events) {
    const payload = parsePayload(eventItem.payload_json || "");
    const runtimeMode = typeof payload?.runtime_mode === "string" ? payload.runtime_mode.trim() : "";
    if (runtimeMode === runtimeModeFullEnv || runtimeMode === runtimeModeCodeOnly) {
      return runtimeMode;
    }
  }

  if (run?.trigger_label === "mode:discussion" || run?.trigger_kind === "ai_repair") {
    return runtimeModeCodeOnly;
  }
  return runtimeModeFullEnv;
}

export function buildRunTimelinePhases(run: Run | null, events: FlowEvent[], locale: string, referenceDate: Date = new Date()): TimelinePhaseStep[] {
  if (!run) return [];

  const runtimeMode = resolveRuntimeMode(events, run);
  const startedAt = findEventAt(events, "run.started") || run.started_at || null;
  const authRequestedAt = findEventAt(events, "run.codex.auth.required");
  const readyAt = findEventAt(events, "run.agent.ready");
  const authResolvedAt = findAuthResolvedAt(events) || (authRequestedAt && readyAt ? readyAt : null);

  const steps: TimelinePhaseStep[] = [
    {
      key: "created",
      at: run.created_at,
      atLabel: formatCompactDateTime(run.created_at, locale, referenceDate),
      color: "info",
      icon: "mdi-calendar-plus",
    },
  ];

  if (runtimeMode === runtimeModeFullEnv) {
    const buildCompleted = Boolean(startedAt);
    const buildFailed = Boolean(run.finished_at) && !buildCompleted;
    steps.push({
      key: "buildDeploy",
      at: buildCompleted ? startedAt : buildFailed ? run.finished_at ?? null : null,
      atLabel: buildCompleted || buildFailed
        ? formatCompactDateTime(buildCompleted ? startedAt : run.finished_at, locale, referenceDate)
        : null,
      subtitleKind: buildFailed ? "buildFailed" : undefined,
      color: buildFailed ? "error" : buildCompleted ? "success" : "warning",
      icon: buildFailed ? "mdi-alert-octagon-outline" : "mdi-hammer-wrench",
      showSpinner: !buildCompleted && !buildFailed,
    });
  }

  if (startedAt) {
    steps.push({
      key: "started",
      at: startedAt,
      atLabel: formatCompactDateTime(startedAt, locale, referenceDate),
      color: "primary",
      icon: "mdi-play-circle-outline",
    });
  }

  if (authRequestedAt) {
    const authResolved = Boolean(authResolvedAt);
    steps.push({
      key: "authResolved",
      at: authResolved ? authResolvedAt : authRequestedAt,
      atLabel: formatCompactDateTime(authResolved ? authResolvedAt : authRequestedAt, locale, referenceDate),
      color: authResolved ? "success" : run.finished_at ? "error" : "warning",
      icon: authResolved ? "mdi-shield-check-outline" : "mdi-lock-clock",
      showSpinner: !authResolved && !run.finished_at,
    });
  }

  if (startedAt || readyAt) {
    const readyReached = Boolean(readyAt);
    steps.push({
      key: "agentReady",
      at: readyAt,
      atLabel: readyReached ? formatCompactDateTime(readyAt, locale, referenceDate) : null,
      color: readyReached ? "success" : run.finished_at ? "error" : "secondary",
      icon: readyReached ? "mdi-robot-happy-outline" : "mdi-robot-outline",
      showSpinner: !readyReached && !run.finished_at,
    });
  }

  if (run.wait_state) {
    steps.push({
      key: "waiting",
      at: run.wait_since ?? null,
      atLabel: formatCompactDateTime(run.wait_since, locale, referenceDate),
      subtitleKind: "waitState",
      subtitleValue: run.wait_state,
      color: "warning",
      icon: "mdi-timer-sand",
    });
  }

  if (run.finished_at) {
    steps.push({
      key: "finished",
      at: run.finished_at,
      atLabel: formatCompactDateTime(run.finished_at, locale, referenceDate),
      subtitleKind: "status",
      subtitleValue: run.status,
      color: run.status === "succeeded" ? "success" : run.status === "failed" ? "error" : "secondary",
      icon: run.status === "succeeded" ? "mdi-check" : run.status === "failed" ? "mdi-alert-octagon-outline" : "mdi-flag-outline",
    });
  }

  return steps;
}

export function buildRunTimelineStatuses(events: FlowEvent[], locale: string, referenceDate: Date = new Date()): TimelineStatusEntry[] {
  const entries: TimelineStatusEntry[] = [];
  for (const eventItem of events) {
    const payload = parsePayload(eventItem.payload_json || "");
    let statusText = "";

    if (eventItem.event_type === "run.agent.status_reported") {
      statusText = typeof payload?.status_text === "string" ? payload.status_text.trim() : "";
    } else {
      statusText = recoveryStatusText(eventItem.event_type, payload) || "";
    }
    if (!statusText) continue;

    const previous = entries[entries.length - 1];
    if (previous && previous.text === statusText) {
      previous.repeatCount += 1;
      continue;
    }

    entries.push({
      key: `status:${eventItem.created_at}:${statusText}`,
      at: eventItem.created_at || null,
      text: statusText,
      timeLabel: formatCompactDateTime(eventItem.created_at, locale, referenceDate),
      repeatCount: 1,
    });
  }
  return entries;
}
