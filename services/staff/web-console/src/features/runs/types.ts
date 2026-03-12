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

export type RealtimePagination = {
  page: number;
  page_size: number;
  total_count: number;
};

export type RunsRealtimeMessageType = "snapshot" | "error";

export type RunsRealtimeMessage = {
  type: RunsRealtimeMessageType;
  items?: Run[];
  pagination?: RealtimePagination;
  wait_queue_count?: number;
  pending_approvals_count?: number;
  message?: string;
  sent_at: string;
};
