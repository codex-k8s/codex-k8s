export type MeDto = {
  user: {
    id: string;
    email: string;
    github_login: string;
    is_platform_admin: boolean;
  };
};

export type UserIdentity = {
  id: string;
  email: string;
  githubLogin: string;
  isPlatformAdmin: boolean;
};

