import { createRealtimeClient, type RealtimeConnectionState } from "../../shared/ws/realtime-client.ts";
import type { RunRealtimeMessage, RunRealtimeMessageType } from "./types";

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
      wait_projection: payload.wait_projection,
      wait_resolution: payload.wait_resolution,
      wait_manual_action: payload.wait_manual_action,
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
