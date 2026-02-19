import {
  deleteRunNamespace as deleteRunNamespaceRequest,
  getRun as getRunRequest,
  getRunAccessKeyStatus as getRunAccessKeyStatusRequest,
  getRunLogs as getRunLogsRequest,
  listPendingApprovals as listPendingApprovalsRequest,
  listRunJobs as listRunJobsRequest,
  listRunEvents as listRunEventsRequest,
  listRunWaits as listRunWaitsRequest,
  regenerateRunAccessKey as regenerateRunAccessKeyRequest,
  listRuns as listRunsRequest,
  revokeRunAccessKey as revokeRunAccessKeyRequest,
  resolveApprovalDecision as resolveApprovalDecisionRequest,
} from "../../shared/api/sdk";

import type {
  ApprovalRequest,
  FlowEvent,
  RunAccessKeyIssueResponse,
  RunAccessKeyStatus,
  ResolveApprovalDecisionResponse,
  Run,
  RunLogs,
  RunNamespaceCleanupResponse,
} from "./types";

export type RunListFilters = {
  triggerKind?: string;
  status?: string;
  agentKey?: string;
};

export type RunWaitFilters = RunListFilters & {
  waitState?: string;
};

export async function listRuns(limit = 1000): Promise<Run[]> {
  const resp = await listRunsRequest({ query: { limit }, throwOnError: true });
  return resp.data.items ?? [];
}

export async function listRunJobs(filters: RunListFilters = {}, limit = 200): Promise<Run[]> {
  const resp = await listRunJobsRequest({
    query: {
      limit,
      trigger_kind: filters.triggerKind?.trim() || undefined,
      status: filters.status?.trim() || undefined,
      agent_key: filters.agentKey?.trim() || undefined,
    },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

export async function listRunWaits(filters: RunWaitFilters = {}, limit = 200): Promise<Run[]> {
  const resp = await listRunWaitsRequest({
    query: {
      limit,
      trigger_kind: filters.triggerKind?.trim() || undefined,
      status: filters.status?.trim() || undefined,
      agent_key: filters.agentKey?.trim() || undefined,
      wait_state: filters.waitState?.trim() || undefined,
    },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

export async function getRun(runId: string): Promise<Run> {
  const resp = await getRunRequest({ path: { run_id: runId }, throwOnError: true });
  return resp.data;
}

export async function deleteRunNamespace(runId: string): Promise<RunNamespaceCleanupResponse> {
  const resp = await deleteRunNamespaceRequest({ path: { run_id: runId }, throwOnError: true });
  return resp.data;
}

export async function listRunEvents(runId: string, limit = 500): Promise<FlowEvent[]> {
  const resp = await listRunEventsRequest({
    path: { run_id: runId },
    query: { limit },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

export async function getRunLogs(runId: string, tailLines = 200): Promise<RunLogs> {
  const resp = await getRunLogsRequest({
    path: { run_id: runId },
    query: { tail_lines: tailLines },
    throwOnError: true,
  });
  return resp.data;
}

export async function getRunAccessKeyStatus(runId: string): Promise<RunAccessKeyStatus> {
  const resp = await getRunAccessKeyStatusRequest({
    path: { run_id: runId },
    throwOnError: true,
  });
  return resp.data;
}

export async function regenerateRunAccessKey(runId: string, ttlSeconds?: number): Promise<RunAccessKeyIssueResponse> {
  const body = ttlSeconds && ttlSeconds > 0 ? { ttl_seconds: ttlSeconds } : undefined;
  const resp = await regenerateRunAccessKeyRequest({
    path: { run_id: runId },
    body,
    throwOnError: true,
  });
  return resp.data;
}

export async function revokeRunAccessKey(runId: string): Promise<RunAccessKeyStatus> {
  const resp = await revokeRunAccessKeyRequest({
    path: { run_id: runId },
    throwOnError: true,
  });
  return resp.data;
}

export async function listPendingApprovals(limit = 200): Promise<ApprovalRequest[]> {
  const resp = await listPendingApprovalsRequest({ query: { limit }, throwOnError: true });
  return resp.data.items ?? [];
}

export async function resolveApprovalDecision(
  approvalRequestId: number,
  decision: "approved" | "denied" | "expired" | "failed",
  reason: string,
): Promise<ResolveApprovalDecisionResponse> {
  const resp = await resolveApprovalDecisionRequest({
    path: { approval_request_id: approvalRequestId },
    body: {
      decision,
      reason: reason.trim() === "" ? undefined : reason,
    },
    throwOnError: true,
  });
  return resp.data;
}
