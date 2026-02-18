import { deleteConfigEntry as deleteConfigEntryRequest, listConfigEntries as listConfigEntriesRequest, upsertConfigEntry as upsertConfigEntryRequest } from "../../shared/api/sdk";
import type { ConfigEntry as ApiConfigEntry } from "../../shared/api/generated/types.gen";
import type { ConfigEntry, ConfigKind, ConfigMutability, ConfigScope } from "./types";

function mapApiConfigEntry(item: ApiConfigEntry): ConfigEntry {
  return {
    id: String(item.id ?? ""),
    scope: item.scope,
    kind: item.kind,
    project_id: item.project_id ?? null,
    repository_id: item.repository_id ?? null,
    key: String(item.key ?? ""),
    value: item.value ?? null,
    sync_targets: Array.isArray(item.sync_targets) ? item.sync_targets.filter((entry): entry is string => typeof entry === "string") : [],
    mutability: String(item.mutability ?? "startup_required"),
    is_dangerous: Boolean(item.is_dangerous),
    updated_at: item.updated_at ?? null,
  };
}

export async function listConfigEntries(params: {
  scope: ConfigScope;
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
  const items = resp.data.items ?? [];
  return items.map(mapApiConfigEntry);
}

export async function upsertConfigEntry(params: {
  scope: ConfigScope;
  kind: ConfigKind;
  projectId: string | null;
  repositoryId: string | null;
  key: string;
  valuePlain: string | null;
  valueSecret: string | null;
  syncTargets: string[];
  mutability: ConfigMutability;
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
  return mapApiConfigEntry(resp.data);
}

export async function deleteConfigEntry(id: string): Promise<void> {
  await deleteConfigEntryRequest({ path: { config_entry_id: id }, throwOnError: true });
}
