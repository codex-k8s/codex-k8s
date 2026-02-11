import {
  createUser as createUserRequest,
  deleteUser as deleteUserRequest,
  listUsers as listUsersRequest,
} from "../../shared/api/sdk";

import type { User } from "./types";

export async function listUsers(limit = 200): Promise<User[]> {
  const resp = await listUsersRequest({ query: { limit }, throwOnError: true });
  return resp.data.items ?? [];
}

export async function createAllowedUser(email: string, isPlatformAdmin: boolean): Promise<void> {
  await createUserRequest({
    body: { email, is_platform_admin: isPlatformAdmin },
    throwOnError: true,
  });
}

export async function deleteUser(userId: string): Promise<void> {
  await deleteUserRequest({ path: { user_id: userId }, throwOnError: true });
}
