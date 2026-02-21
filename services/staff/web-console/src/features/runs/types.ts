import type {
  ApprovalRequest,
  FlowEvent,
  ResolveApprovalDecisionResponse,
  Run,
  RunLogs,
  RunNamespaceCleanupResponse,
} from "../../shared/api/generated";

export type {
  ApprovalRequest,
  FlowEvent,
  ResolveApprovalDecisionResponse,
  Run,
  RunLogs,
  RunNamespaceCleanupResponse,
} from "../../shared/api/generated";

export type RunRealtimeMessageType = "snapshot" | "run" | "events" | "logs" | "error";

export type RunRealtimeMessage = {
  type: RunRealtimeMessageType;
  run?: Run;
  events?: FlowEvent[];
  logs?: RunLogs;
  message?: string;
  sent_at: string;
};
