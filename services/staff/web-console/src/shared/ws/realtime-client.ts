export type RealtimeConnectionState = "connecting" | "connected" | "reconnecting" | "closed";

type RealtimeClientConfig<TMessage> = {
  url: string;
  parseMessage: (raw: string) => TMessage | null;
  onMessage: (message: TMessage) => void;
  onStateChange?: (state: RealtimeConnectionState) => void;
  minReconnectDelayMs?: number;
  maxReconnectDelayMs?: number;
  reconnectFactor?: number;
  reconnectJitter?: number;
};

type RealtimeClient = {
  start: () => void;
  stop: () => void;
};

const defaultMinReconnectDelayMs = 600;
const defaultMaxReconnectDelayMs = 12000;
const defaultReconnectFactor = 1.8;
const defaultReconnectJitter = 0.2;

export function createRealtimeClient<TMessage>(config: RealtimeClientConfig<TMessage>): RealtimeClient {
  let socket: WebSocket | null = null;
  let retryTimer: number | null = null;
  let reconnectAttempt = 0;
  let active = false;

  const minDelayMs = Math.max(100, config.minReconnectDelayMs ?? defaultMinReconnectDelayMs);
  const maxDelayMs = Math.max(minDelayMs, config.maxReconnectDelayMs ?? defaultMaxReconnectDelayMs);
  const reconnectFactor = Math.max(1.1, config.reconnectFactor ?? defaultReconnectFactor);
  const reconnectJitter = Math.min(0.9, Math.max(0, config.reconnectJitter ?? defaultReconnectJitter));

  function notifyState(state: RealtimeConnectionState): void {
    config.onStateChange?.(state);
  }

  function clearRetryTimer(): void {
    if (retryTimer !== null) {
      window.clearTimeout(retryTimer);
      retryTimer = null;
    }
  }

  function closeSocket(): void {
    if (!socket) return;
    socket.onopen = null;
    socket.onmessage = null;
    socket.onerror = null;
    socket.onclose = null;
    try {
      socket.close();
    } catch {
      // Ignore close errors.
    }
    socket = null;
  }

  function scheduleReconnect(): void {
    if (!active) return;
    reconnectAttempt += 1;
    notifyState("reconnecting");

    const exponential = minDelayMs * Math.pow(reconnectFactor, reconnectAttempt - 1);
    const capped = Math.min(maxDelayMs, exponential);
    const jitterWindow = capped * reconnectJitter;
    const jitterOffset = (Math.random() * 2 - 1) * jitterWindow;
    const delayMs = Math.max(minDelayMs, Math.round(capped + jitterOffset));

    clearRetryTimer();
    retryTimer = window.setTimeout(() => {
      retryTimer = null;
      connect();
    }, delayMs);
  }

  function connect(): void {
    if (!active) return;
    closeSocket();
    notifyState(reconnectAttempt === 0 ? "connecting" : "reconnecting");

    socket = new WebSocket(config.url);

    socket.onopen = () => {
      reconnectAttempt = 0;
      notifyState("connected");
    };

    socket.onmessage = (event: MessageEvent<string>) => {
      const parsed = config.parseMessage(String(event.data ?? ""));
      if (!parsed) return;
      config.onMessage(parsed);
    };

    socket.onerror = () => {
      if (!socket) return;
      try {
        socket.close();
      } catch {
        // Ignore close errors.
      }
    };

    socket.onclose = () => {
      if (!active) return;
      scheduleReconnect();
    };
  }

  function start(): void {
    if (active) return;
    active = true;
    reconnectAttempt = 0;
    connect();
  }

  function stop(): void {
    active = false;
    clearRetryTimer();
    closeSocket();
    notifyState("closed");
  }

  return { start, stop };
}
