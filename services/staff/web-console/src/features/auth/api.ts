import { getMe, logout as logoutRequest } from "../../shared/api/sdk";

import type { MeDto } from "./types";

export async function fetchMe(): Promise<MeDto> {
  const resp = await getMe({ throwOnError: true });
  return resp.data;
}

export async function logout(): Promise<void> {
  await logoutRequest({ throwOnError: true });
}
