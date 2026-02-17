import { deleteConfigEntry as deleteConfigEntryRequest, listConfigEntries as listConfigEntriesRequest, upsertConfigEntry as upsertConfigEntryRequest } from "../../shared/api/sdk";
import type { ConfigEntry } from "./types";

export async function listConfigEntries(params: {
  scope: "platform" | "project" | "repository";
  projectId?: string;
  repositoryId?: string;
  limit?: number;
}): Promise<ConfigEntry[]> {
  const resp = await listConfigEntriesRequest({
    query: {
      scope: params.scope,
      project_id: params.projectId,
      repository_id: params.repositoryId,
      limit: params.limit,
    },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

export async function upsertConfigEntry(params: {
  scope: "platform" | "project" | "repository";
  kind: "variable" | "secret";
  projectId: string | null;
  repositoryId: string | null;
  key: string;
  valuePlain: string | null;
  valueSecret: string | null;
  syncTargets: string[];
  mutability: "startup_required" | "runtime_mutable";
  isDangerous: boolean;
  dangerousConfirmed: boolean;
}): Promise<ConfigEntry> {
  const resp = await upsertConfigEntryRequest({
    body: {
      scope: params.scope,
      kind: params.kind,
      project_id: params.projectId,
      repository_id: params.repositoryId,
      key: params.key,
      value_plain: params.valuePlain,
      value_secret: params.valueSecret,
      sync_targets: params.syncTargets,
      mutability: params.mutability,
      is_dangerous: params.isDangerous,
      dangerous_confirmed: params.dangerousConfirmed,
    },
    throwOnError: true,
  });
  return resp.data;
}

export async function deleteConfigEntry(id: string): Promise<void> {
  await deleteConfigEntryRequest({ path: { config_entry_id: id }, throwOnError: true });
}
