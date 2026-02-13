import {
  deleteRunNamespace as deleteRunNamespaceRequest,
  getRun as getRunRequest,
  listPendingApprovals as listPendingApprovalsRequest,
  listRunEvents as listRunEventsRequest,
  listRuns as listRunsRequest,
  resolveApprovalDecision as resolveApprovalDecisionRequest,
} from "../../shared/api/sdk";

import type {
  ApprovalRequest,
  FlowEvent,
  ResolveApprovalDecisionResponse,
  Run,
  RunNamespaceCleanupResponse,
} from "./types";

export async function listRuns(limit = 1000): Promise<Run[]> {
  const resp = await listRunsRequest({ query: { limit }, throwOnError: true });
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
