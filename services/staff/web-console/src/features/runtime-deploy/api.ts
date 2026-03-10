import {
  cancelRuntimeDeployTask as cancelRuntimeDeployTaskRequest,
  getRuntimeDeployTask as getRuntimeDeployTaskRequest,
  listRuntimeDeployTasks as listRuntimeDeployTasksRequest,
  stopRuntimeDeployTask as stopRuntimeDeployTaskRequest,
} from "../../shared/api/sdk";

import type {
  RuntimeDeployTaskActionResponse,
  RuntimeDeployTaskListItem,
  RuntimeDeployTask,
} from "./types";

export type RuntimeDeployTaskFilters = {
  status?: "pending" | "running" | "succeeded" | "failed" | "canceled";
  targetEnv?: string;
};

export async function listRuntimeDeployTasks(filters: RuntimeDeployTaskFilters = {}, limit = 30): Promise<RuntimeDeployTaskListItem[]> {
  const resp = await listRuntimeDeployTasksRequest({
    query: {
      limit,
      status: filters.status || undefined,
      target_env: filters.targetEnv?.trim() || undefined,
    },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

export async function getRuntimeDeployTask(runId: string): Promise<RuntimeDeployTask> {
  const resp = await getRuntimeDeployTaskRequest({
    path: { run_id: runId },
    throwOnError: true,
  });
  return resp.data;
}

export async function cancelRuntimeDeployTask(runId: string, reason = ""): Promise<RuntimeDeployTaskActionResponse> {
  const resp = await cancelRuntimeDeployTaskRequest({
    path: { run_id: runId },
    body: {
      reason: reason.trim() || undefined,
      force: false,
    },
    throwOnError: true,
  });
  return resp.data;
}

export async function stopRuntimeDeployTask(runId: string, reason = ""): Promise<RuntimeDeployTaskActionResponse> {
  const resp = await stopRuntimeDeployTaskRequest({
    path: { run_id: runId },
    body: {
      reason: reason.trim() || undefined,
      force: true,
    },
    throwOnError: true,
  });
  return resp.data;
}
