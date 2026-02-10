import { http } from "../../shared/api/http";

import type { MeDto } from "./types";

export async function fetchMe(): Promise<MeDto> {
  const resp = await http.get("/api/v1/auth/me");
  return resp.data as MeDto;
}

export async function logout(): Promise<void> {
  await http.post("/api/v1/auth/logout");
}

