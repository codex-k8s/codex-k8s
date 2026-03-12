export type RealtimeConnectionState = "connecting" | "connected" | "reconnecting" | "closed";

type RealtimeClientConfig<TMessage> = {
  url: string;
  parseMessage: (raw: string) => TMessage | null;
  onMessage: (message: TMessage) => void;
  onStateChange?: (state: RealtimeConnectionState) => void;
  firstMessageTimeoutMs?: number;
  onFirstMessageTimeout?: () => void;
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
  let firstMessageTimer: number | null = null;
  let reconnectAttempt = 0;
  let active = false;
  let hasReceivedMessage = false;
  let firstMessageTimeoutNotified = false;

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

  function clearFirstMessageTimer(): void {
    if (firstMessageTimer !== null) {
      window.clearTimeout(firstMessageTimer);
      firstMessageTimer = null;
    }
  }

  function detachSocketHandlers(target: WebSocket): void {
    target.onopen = null;
    target.onmessage = null;
    target.onerror = null;
    target.onclose = null;
  }

  function armFirstMessageTimer(): void {
    const timeoutMs = config.firstMessageTimeoutMs;
    if (!active || hasReceivedMessage || firstMessageTimeoutNotified || !timeoutMs || timeoutMs <= 0) {
      return;
    }

    clearFirstMessageTimer();
    firstMessageTimer = window.setTimeout(() => {
      firstMessageTimer = null;
      if (!active || hasReceivedMessage || firstMessageTimeoutNotified) {
        return;
      }

      firstMessageTimeoutNotified = true;
      config.onFirstMessageTimeout?.();
    }, timeoutMs);
  }

  function closeSocket(): void {
    if (!socket) return;
    const currentSocket = socket;
    socket = null;
    if (currentSocket.readyState === WebSocket.CONNECTING) {
      currentSocket.onopen = () => {
        detachSocketHandlers(currentSocket);
        try {
          currentSocket.close();
        } catch {
          // Ignore close errors.
        }
      };
      currentSocket.onmessage = null;
      currentSocket.onerror = null;
      currentSocket.onclose = null;
      return;
    }
    detachSocketHandlers(currentSocket);
    try {
      currentSocket.close();
    } catch {
      // Ignore close errors.
    }
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

    const currentSocket = new WebSocket(config.url);
    socket = currentSocket;

    currentSocket.onopen = () => {
      if (socket !== currentSocket) {
        return;
      }
      reconnectAttempt = 0;
      notifyState("connected");
    };

    currentSocket.onmessage = (event: MessageEvent<string>) => {
      if (socket !== currentSocket) {
        return;
      }
      const parsed = config.parseMessage(String(event.data ?? ""));
      if (!parsed) return;
      hasReceivedMessage = true;
      clearFirstMessageTimer();
      config.onMessage(parsed);
    };

    currentSocket.onerror = () => {
      if (socket !== currentSocket) return;
      if (currentSocket.readyState === WebSocket.CONNECTING) {
        return;
      }
      try {
        currentSocket.close();
      } catch {
        // Ignore close errors.
      }
    };

    currentSocket.onclose = () => {
      if (socket === currentSocket) {
        socket = null;
      }
      if (!active) return;
      if (socket !== null) return;
      scheduleReconnect();
    };
  }

  function start(): void {
    if (active) return;
    active = true;
    reconnectAttempt = 0;
    hasReceivedMessage = false;
    firstMessageTimeoutNotified = false;
    armFirstMessageTimer();
    connect();
  }

  function stop(): void {
    active = false;
    clearRetryTimer();
    clearFirstMessageTimer();
    closeSocket();
    notifyState("closed");
  }

  return { start, stop };
}
