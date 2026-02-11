import type { MeResponse } from "../../shared/api/generated";

export type MeDto = MeResponse;

export type UserIdentity = {
  id: string;
  email: string;
  githubLogin: string;
  isPlatformAdmin: boolean;
  isPlatformOwner: boolean;
};
