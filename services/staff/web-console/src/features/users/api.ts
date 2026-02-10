import { http } from "../../shared/api/http";

import type { UserDto } from "./types";

type ItemsResponse<T> = { items: T[] };

export async function listUsers(limit = 200): Promise<UserDto[]> {
  const resp = await http.get("/api/v1/staff/users", { params: { limit } });
  return (resp.data as ItemsResponse<UserDto>).items ?? [];
}

export async function createAllowedUser(email: string, isPlatformAdmin: boolean): Promise<void> {
  await http.post("/api/v1/staff/users", { email, is_platform_admin: isPlatformAdmin });
}

