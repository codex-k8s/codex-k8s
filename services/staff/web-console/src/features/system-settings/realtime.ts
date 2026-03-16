import { createRealtimeClient, type RealtimeConnectionState } from "../../shared/ws/realtime-client";
import type { SystemSettingsRealtimeMessage } from "./types";

export type SystemSettingsRealtimeConnectionState = Exclude<RealtimeConnectionState, "closed">;

type SubscribeSystemSettingsRealtimeParams = {
  onMessage: (message: SystemSettingsRealtimeMessage) => void;
  onStateChange?: (state: SystemSettingsRealtimeConnectionState) => void;
  onInitialMessageTimeout?: () => void;
};

const realtimeMessageTypes = new Set<SystemSettingsRealtimeMessage["type"]>(["snapshot", "error"]);
const initialSystemSettingsRealtimeTimeoutMs = 10000;

function buildSystemSettingsRealtimeURL(): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/api/v1/staff/system-settings/realtime`;
}

function parseSystemSettingsRealtimeMessage(raw: string): SystemSettingsRealtimeMessage | null {
  const text = String(raw || "").trim();
  if (!text) return null;

  try {
    const payload = JSON.parse(text) as Partial<SystemSettingsRealtimeMessage>;
    const type = String(payload.type || "") as SystemSettingsRealtimeMessage["type"];
    if (!realtimeMessageTypes.has(type)) return null;
    const sentAt = String(payload.sent_at || "");
    if (!sentAt) return null;

    return {
      type,
      items: Array.isArray(payload.items) ? payload.items : undefined,
      message: typeof payload.message === "string" ? payload.message : undefined,
      sent_at: sentAt,
    };
  } catch {
    return null;
  }
}

export function subscribeSystemSettingsRealtime(params: SubscribeSystemSettingsRealtimeParams): () => void {
  const client = createRealtimeClient<SystemSettingsRealtimeMessage>({
    url: buildSystemSettingsRealtimeURL(),
    parseMessage: parseSystemSettingsRealtimeMessage,
    onMessage: params.onMessage,
    firstMessageTimeoutMs: initialSystemSettingsRealtimeTimeoutMs,
    onFirstMessageTimeout: params.onInitialMessageTimeout,
    onStateChange: (state) => {
      if (state === "closed") return;
      params.onStateChange?.(state);
    },
  });

  client.start();
  return () => client.stop();
}
