import { defineStore } from "pinia";

import { useUiContextStore } from "../ui-context/store";

export type RealtimeStatus = "disconnected" | "connecting" | "connected";

export type RealtimeEvent = {
  id: number;
  topic: string;
  scope: unknown;
  payload: unknown;
  correlation_id: string;
  project_id: string;
  run_id: string;
  task_id: string;
  created_at: string;
};

type RealtimeEventListener = (event: RealtimeEvent) => void;

const reconnectMinDelayMs = 1000;
const reconnectMaxDelayMs = 15000;

let socket: WebSocket | null = null;
let reconnectTimer: number | null = null;
let listeners = new Map<number, RealtimeEventListener>();
let listenerSeq = 0;

export const useRealtimeStore = defineStore("realtime", {
  state: () => ({
    status: "disconnected" as RealtimeStatus,
    lastEventId: 0,
    reconnectAttempt: 0,
    started: false,
  }),
  getters: {
    isConnected: (s) => s.status === "connected",
  },
  actions: {
    start(): void {
      if (typeof window === "undefined") return;
      this.started = true;
      this.connect();
    },
    stop(): void {
      this.started = false;
      this.status = "disconnected";
      this.reconnectAttempt = 0;
      if (reconnectTimer !== null) {
        window.clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
      if (socket) {
        try {
          socket.close();
        } catch {
          // no-op
        }
      }
      socket = null;
    },
    reconnect(): void {
      this.stop();
      this.started = true;
      this.connect();
    },
    connect(): void {
      if (!this.started || typeof window === "undefined") return;
      if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
        return;
      }
      this.status = "connecting";
      const url = this.buildURL();
      socket = new WebSocket(url);

      socket.onopen = () => {
        this.status = "connected";
        this.reconnectAttempt = 0;
      };

      socket.onmessage = (raw) => {
        const parsed = parseRealtimeMessage(raw.data);
        if (!parsed || parsed.type !== "event" || !parsed.event) {
          return;
        }
        const event = parsed.event;
        if (!event.id || event.id <= this.lastEventId) {
          return;
        }
        this.lastEventId = event.id;
        for (const listener of listeners.values()) {
          try {
            listener(event);
          } catch {
            // no-op
          }
        }
        this.sendAck(this.lastEventId);
      };

      socket.onerror = () => {
        this.status = "disconnected";
      };

      socket.onclose = () => {
        socket = null;
        this.status = "disconnected";
        if (!this.started) {
          return;
        }
        this.scheduleReconnect();
      };
    },
    scheduleReconnect(): void {
      if (typeof window === "undefined" || !this.started) return;
      if (reconnectTimer !== null) {
        window.clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
      this.reconnectAttempt += 1;
      const delay = Math.min(reconnectMinDelayMs * 2 ** Math.max(this.reconnectAttempt - 1, 0), reconnectMaxDelayMs);
      reconnectTimer = window.setTimeout(() => {
        reconnectTimer = null;
        this.connect();
      }, delay);
    },
    sendAck(lastEventId: number): void {
      if (!socket || socket.readyState !== WebSocket.OPEN || lastEventId <= 0) return;
      try {
        socket.send(JSON.stringify({ type: "ack", last_event_id: lastEventId }));
      } catch {
        // no-op
      }
    },
    buildURL(): string {
      const protocol = window.location.protocol === "https:" ? "wss" : "ws";
      const uiContext = useUiContextStore();
      const params = new URLSearchParams();
      params.set("topics", "run.events,run.status,run.logs,deploy.events,deploy.logs,system.errors");
      if (this.lastEventId > 0) {
        params.set("last_event_id", String(this.lastEventId));
      }
      if (uiContext.projectId) {
        params.set("project_id", uiContext.projectId);
      }
      return `${protocol}://${window.location.host}/api/v1/staff/realtime/ws?${params.toString()}`;
    },
    subscribe(listener: RealtimeEventListener): () => void {
      listenerSeq += 1;
      const id = listenerSeq;
      listeners.set(id, listener);
      return () => {
        listeners.delete(id);
      };
    },
  },
});

type WSInboundMessage =
  | { type: "event"; event: RealtimeEvent }
  | { type: "hello"; meta?: { status?: string; last_event_id?: number } }
  | { type: "subscribed"; meta?: { status?: string; last_event_id?: number } }
  | { type: "error"; error?: { code?: string; message?: string } };

function parseRealtimeMessage(raw: unknown): WSInboundMessage | null {
  if (typeof raw !== "string" || raw.trim() === "") return null;
  try {
    return JSON.parse(raw) as WSInboundMessage;
  } catch {
    return null;
  }
}

