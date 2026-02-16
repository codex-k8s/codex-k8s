import {
  cleanupRegistryImages as cleanupRegistryImagesRequest,
  deleteRegistryImageTag as deleteRegistryImageTagRequest,
  getRuntimeDeployTask as getRuntimeDeployTaskRequest,
  listRegistryImages as listRegistryImagesRequest,
  listRuntimeDeployTasks as listRuntimeDeployTasksRequest,
} from "../../shared/api/sdk";

import type {
  CleanupRegistryImagesResponse,
  RegistryImageDeleteResult,
  RegistryImageRepository,
  RuntimeDeployTask,
} from "./types";

export type RuntimeDeployTaskFilters = {
  status?: "pending" | "running" | "succeeded" | "failed";
  targetEnv?: string;
};

export type RegistryImagesFilters = {
  repository?: string;
  limitRepositories?: number;
  limitTags?: number;
};

export type CleanupRegistryImagesParams = {
  repositoryPrefix?: string;
  keepTags?: number;
  limitRepositories?: number;
  dryRun: boolean;
};

export async function listRuntimeDeployTasks(filters: RuntimeDeployTaskFilters = {}, limit = 200): Promise<RuntimeDeployTask[]> {
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

export async function listRegistryImages(filters: RegistryImagesFilters = {}): Promise<RegistryImageRepository[]> {
  const resp = await listRegistryImagesRequest({
    query: {
      repository: filters.repository?.trim() || undefined,
      limit_repositories: filters.limitRepositories,
      limit_tags: filters.limitTags,
    },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

export async function deleteRegistryImageTag(repository: string, tag: string): Promise<RegistryImageDeleteResult> {
  const resp = await deleteRegistryImageTagRequest({
    body: {
      repository: repository.trim(),
      tag: tag.trim(),
    },
    throwOnError: true,
  });
  return resp.data;
}

export async function cleanupRegistryImages(params: CleanupRegistryImagesParams): Promise<CleanupRegistryImagesResponse> {
  const resp = await cleanupRegistryImagesRequest({
    body: {
      repository_prefix: params.repositoryPrefix?.trim() || undefined,
      keep_tags: params.keepTags,
      limit_repositories: params.limitRepositories,
      dry_run: params.dryRun,
    },
    throwOnError: true,
  });
  return resp.data;
}
