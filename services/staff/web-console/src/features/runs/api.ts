import {
  deleteRunNamespace as deleteRunNamespaceRequest,
  getRun as getRunRequest,
  listRunEvents as listRunEventsRequest,
  listRunLearningFeedback as listRunLearningFeedbackRequest,
  listRuns as listRunsRequest,
} from "../../shared/api/sdk";

import type { FlowEvent, LearningFeedback, Run, RunNamespaceCleanupResponse } from "./types";

export async function listRuns(limit = 200): Promise<Run[]> {
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

export async function listRunLearningFeedback(runId: string, limit = 200): Promise<LearningFeedback[]> {
  const resp = await listRunLearningFeedbackRequest({
    path: { run_id: runId },
    query: { limit },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}
