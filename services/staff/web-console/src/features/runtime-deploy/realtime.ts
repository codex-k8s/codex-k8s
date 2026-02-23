import { subscribeRunRealtime, type RunRealtimeState } from "../runs/realtime";
import type { RunRealtimeMessage } from "../runs/types";

export type RuntimeDeployRealtimeState = RunRealtimeState;
export type RuntimeDeployRealtimeMessage = RunRealtimeMessage;

type SubscribeRuntimeDeployRealtimeParams = {
  runId: string;
  onMessage: (message: RuntimeDeployRealtimeMessage) => void;
  onStateChange?: (state: RuntimeDeployRealtimeState) => void;
};

export function subscribeRuntimeDeployRealtime(params: SubscribeRuntimeDeployRealtimeParams): () => void {
  const runId = String(params.runId || "").trim();
  if (!runId) {
    return () => undefined;
  }

  return subscribeRunRealtime({
    runId,
    includeLogs: false,
    eventsLimit: 100,
    tailLines: 50,
    onMessage: params.onMessage,
    onStateChange: params.onStateChange,
  });
}
