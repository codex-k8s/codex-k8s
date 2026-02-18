import { listRuntimeErrors as listRuntimeErrorsRequest, markRuntimeErrorViewed as markRuntimeErrorViewedRequest } from "../../shared/api/sdk";

import type { RuntimeError } from "./types";

export type RuntimeErrorsFilters = {
  state?: "active" | "viewed" | "all";
  level?: "error" | "warning" | "critical";
  source?: string;
  runId?: string;
  correlationId?: string;
};

export async function listRuntimeErrors(filters: RuntimeErrorsFilters = {}, limit = 5): Promise<RuntimeError[]> {
  const resp = await listRuntimeErrorsRequest({
    query: {
      limit,
      state: filters.state,
      level: filters.level,
      source: filters.source?.trim() || undefined,
      run_id: filters.runId?.trim() || undefined,
      correlation_id: filters.correlationId?.trim() || undefined,
    },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

export async function markRuntimeErrorViewed(runtimeErrorId: string): Promise<RuntimeError> {
  const resp = await markRuntimeErrorViewedRequest({
    path: { runtime_error_id: runtimeErrorId.trim() },
    throwOnError: true,
  });
  return resp.data;
}

