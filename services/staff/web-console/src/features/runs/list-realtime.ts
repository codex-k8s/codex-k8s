import { createRealtimeClient, type RealtimeConnectionState } from "../../shared/ws/realtime-client";
import type { RealtimePagination, RunsRealtimeMessage, RunsRealtimeMessageType } from "./types";

export type RunsListRealtimeState = Exclude<RealtimeConnectionState, "closed">;

type SubscribeRunsRealtimeParams = {
  page: number;
  pageSize: number;
  onMessage: (message: RunsRealtimeMessage) => void;
  onStateChange?: (state: RunsListRealtimeState) => void;
  onInitialMessageTimeout?: () => void;
};

const runsRealtimeMessageTypes = new Set<RunsRealtimeMessageType>(["snapshot", "error"]);
const initialRunsRealtimeMessageTimeoutMs = 10000;

function buildRunsRealtimeURL(params: { page: number; pageSize: number }): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const url = new URL(`${protocol}//${window.location.host}/api/v1/staff/runs/realtime`);
  url.searchParams.set("page", String(Math.max(1, Math.trunc(params.page))));
  url.searchParams.set("page_size", String(Math.max(1, Math.trunc(params.pageSize))));
  return url.toString();
}

function parsePagination(raw: unknown): RealtimePagination | undefined {
  if (!raw || typeof raw !== "object") return undefined;
  const payload = raw as Partial<RealtimePagination>;
  const page = Number(payload.page);
  const pageSize = Number(payload.page_size);
  const totalCount = Number(payload.total_count);
  if (!Number.isFinite(page) || !Number.isFinite(pageSize) || !Number.isFinite(totalCount)) {
    return undefined;
  }
  return {
    page: Math.max(1, Math.trunc(page)),
    page_size: Math.max(1, Math.trunc(pageSize)),
    total_count: Math.max(0, Math.trunc(totalCount)),
  };
}

function parseRunsRealtimeMessage(raw: string): RunsRealtimeMessage | null {
  const text = String(raw || "").trim();
  if (!text) return null;
  try {
    const payload = JSON.parse(text) as Partial<RunsRealtimeMessage>;
    if (!payload || typeof payload !== "object") return null;
    const type = String(payload.type || "") as RunsRealtimeMessageType;
    if (!runsRealtimeMessageTypes.has(type)) return null;
    return {
      type,
      items: Array.isArray(payload.items) ? payload.items : undefined,
      pagination: parsePagination(payload.pagination),
      wait_queue_count: Number.isFinite(Number(payload.wait_queue_count))
        ? Math.max(0, Math.trunc(Number(payload.wait_queue_count)))
        : undefined,
      pending_approvals_count: Number.isFinite(Number(payload.pending_approvals_count))
        ? Math.max(0, Math.trunc(Number(payload.pending_approvals_count)))
        : undefined,
      message: typeof payload.message === "string" ? payload.message : undefined,
      sent_at: String(payload.sent_at || new Date().toISOString()),
    };
  } catch {
    return null;
  }
}

export function subscribeRunsRealtime(params: SubscribeRunsRealtimeParams): () => void {
  const url = buildRunsRealtimeURL({
    page: params.page,
    pageSize: params.pageSize,
  });

  const client = createRealtimeClient<RunsRealtimeMessage>({
    url,
    parseMessage: parseRunsRealtimeMessage,
    onMessage: params.onMessage,
    firstMessageTimeoutMs: initialRunsRealtimeMessageTimeoutMs,
    onFirstMessageTimeout: params.onInitialMessageTimeout,
    onStateChange: (state) => {
      if (state === "closed") return;
      params.onStateChange?.(state);
    },
  });

  client.start();
  return () => client.stop();
}
