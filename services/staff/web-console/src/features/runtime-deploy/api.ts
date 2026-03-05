import {
  getRuntimeDeployTask as getRuntimeDeployTaskRequest,
  listRuntimeDeployTasks as listRuntimeDeployTasksRequest,
} from "../../shared/api/sdk";

import type {
  RuntimeDeployTaskListItem,
  RuntimeDeployTask,
} from "./types";

export type RuntimeDeployTaskFilters = {
  status?: "pending" | "running" | "succeeded" | "failed";
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
