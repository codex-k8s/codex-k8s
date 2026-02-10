export type MeDto = {
  user: {
    id: string;
    email: string;
    github_login: string;
    is_platform_admin: boolean;
    is_platform_owner: boolean;
  };
};

export type UserIdentity = {
  id: string;
  email: string;
  githubLogin: string;
  isPlatformAdmin: boolean;
  isPlatformOwner: boolean;
};
