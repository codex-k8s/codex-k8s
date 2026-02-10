export type UserDto = {
  id: string;
  email: string;
  github_user_id: number | null;
  github_login: string | null;
  is_platform_admin: boolean;
};

export type User = {
  id: string;
  email: string;
  githubLogin: string | null;
  isPlatformAdmin: boolean;
};

