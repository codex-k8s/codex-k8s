export type ConfigScope = "platform" | "project" | "repository";
export type ConfigKind = "variable" | "secret";
export type ConfigMutability = "startup_required" | "runtime_mutable";

export type ConfigEntry = {
  id: string;
  scope: ConfigScope;
  kind: ConfigKind;
  project_id: string | null;
  repository_id: string | null;
  key: string;
  value: string | null;
  sync_targets: string[];
  mutability: ConfigMutability;
  is_dangerous: boolean;
  updated_at: string | null;
};
