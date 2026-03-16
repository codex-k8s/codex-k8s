import {
  listSystemSettings as listSystemSettingsRequest,
  resetSystemSetting as resetSystemSettingRequest,
  updateSystemSettingBoolean as updateSystemSettingBooleanRequest,
} from "../../shared/api/sdk";

import type { SystemSetting } from "./types";

export async function listSystemSettings(): Promise<SystemSetting[]> {
  const resp = await listSystemSettingsRequest({ throwOnError: true });
  return resp.data.items ?? [];
}

export async function updateSystemSettingBoolean(settingKey: string, booleanValue: boolean): Promise<SystemSetting> {
  const resp = await updateSystemSettingBooleanRequest({
    path: { setting_key: settingKey },
    body: { boolean_value: booleanValue },
    throwOnError: true,
  });
  return resp.data;
}

export async function resetSystemSetting(settingKey: string): Promise<SystemSetting> {
  const resp = await resetSystemSettingRequest({
    path: { setting_key: settingKey },
    throwOnError: true,
  });
  return resp.data;
}
