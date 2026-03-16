import type { RealtimeConnectionState } from "../../shared/ws/realtime-client";
import type { SystemSetting as GeneratedSystemSetting } from "../../shared/api/sdk";

export type SystemSetting = GeneratedSystemSetting;

export type SystemSettingsRealtimeState = Exclude<RealtimeConnectionState, "closed">;

export type SystemSettingsRealtimeMessage = {
  type: "snapshot" | "error";
  items?: SystemSetting[];
  message?: string;
  sent_at: string;
};
