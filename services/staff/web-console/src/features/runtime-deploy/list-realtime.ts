import { createRealtimeClient, type RealtimeConnectionState } from "../../shared/ws/realtime-client";
import type { RuntimeDeployTasksRealtimeMessage, RuntimeDeployTasksRealtimeMessageType } from "./types";
import type { RealtimePagination } from "../runs/types";

export type RuntimeDeployTasksListRealtimeState = Exclude<RealtimeConnectionState, "closed">;

type SubscribeRuntimeDeployTasksRealtimeParams = {
  page: number;
  pageSize: number;
  status?: string;
  targetEnv?: string;
  onMessage: (message: RuntimeDeployTasksRealtimeMessage) => void;
  onStateChange?: (state: RuntimeDeployTasksListRealtimeState) => void;
  onInitialMessageTimeout?: () => void;
};

const runtimeDeployTasksRealtimeMessageTypes = new Set<RuntimeDeployTasksRealtimeMessageType>(["snapshot", "error"]);
const initialRuntimeDeployTasksRealtimeMessageTimeoutMs = 10000;

function buildRuntimeDeployTasksRealtimeURL(params: { page: number; pageSize: number; status?: string; targetEnv?: string }): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const url = new URL(`${protocol}//${window.location.host}/api/v1/staff/runtime-deploy/tasks/realtime`);
  url.searchParams.set("page", String(Math.max(1, Math.trunc(params.page))));
  url.searchParams.set("page_size", String(Math.max(1, Math.trunc(params.pageSize))));
  if (params.status) {
    url.searchParams.set("status", params.status);
  }
  if (params.targetEnv) {
    url.searchParams.set("target_env", params.targetEnv);
  }
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

function parseRuntimeDeployTasksRealtimeMessage(raw: string): RuntimeDeployTasksRealtimeMessage | null {
  const text = String(raw || "").trim();
  if (!text) return null;
  try {
    const payload = JSON.parse(text) as Partial<RuntimeDeployTasksRealtimeMessage>;
    if (!payload || typeof payload !== "object") return null;
    const type = String(payload.type || "") as RuntimeDeployTasksRealtimeMessageType;
    if (!runtimeDeployTasksRealtimeMessageTypes.has(type)) return null;
    return {
      type,
      items: Array.isArray(payload.items) ? payload.items : undefined,
      pagination: parsePagination(payload.pagination),
      message: typeof payload.message === "string" ? payload.message : undefined,
      sent_at: String(payload.sent_at || new Date().toISOString()),
    };
  } catch {
    return null;
  }
}

export function subscribeRuntimeDeployTasksRealtime(params: SubscribeRuntimeDeployTasksRealtimeParams): () => void {
  const url = buildRuntimeDeployTasksRealtimeURL({
    page: params.page,
    pageSize: params.pageSize,
    status: params.status?.trim() || undefined,
    targetEnv: params.targetEnv?.trim() || undefined,
  });

  const client = createRealtimeClient<RuntimeDeployTasksRealtimeMessage>({
    url,
    parseMessage: parseRuntimeDeployTasksRealtimeMessage,
    onMessage: params.onMessage,
    firstMessageTimeoutMs: initialRuntimeDeployTasksRealtimeMessageTimeoutMs,
    onFirstMessageTimeout: params.onInitialMessageTimeout,
    onStateChange: (state) => {
      if (state === "closed") return;
      params.onStateChange?.(state);
    },
  });

  client.start();
  return () => client.stop();
}
