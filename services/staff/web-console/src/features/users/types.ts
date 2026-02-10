export type UserDto = {
  id: string;
  email: string;
  github_user_id: number | null;
  github_login: string | null;
  is_platform_admin: boolean;
  is_platform_owner: boolean;
};

export type User = {
  id: string;
  email: string;
  githubLogin: string | null;
  isPlatformAdmin: boolean;
  isPlatformOwner: boolean;
};
