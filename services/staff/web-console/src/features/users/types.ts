import type { User as UserAPIModel } from "../../shared/api/generated";

export type UserDto = UserAPIModel;

export type User = {
  id: string;
  email: string;
  githubLogin: string | null;
  isPlatformAdmin: boolean;
  isPlatformOwner: boolean;
};
