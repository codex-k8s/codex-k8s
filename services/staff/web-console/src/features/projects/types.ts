export type ProjectDto = {
  id: string;
  slug: string;
  name: string;
  role: "read" | "read_write" | "admin" | string;
};

export type Project = {
  id: string;
  slug: string;
  name: string;
  role: string;
};

export type RepositoryBindingDto = {
  id: string;
  project_id: string;
  provider: string;
  external_id: number;
  owner: string;
  name: string;
  services_yaml_path: string;
};

export type RepositoryBinding = {
  id: string;
  projectId: string;
  provider: string;
  externalId: number;
  owner: string;
  name: string;
  servicesYamlPath: string;
};

export type ProjectMemberDto = {
  project_id: string;
  user_id: string;
  email: string;
  role: "read" | "read_write" | "admin";
  learning_mode_override: boolean | null;
};

export type ProjectMember = {
  projectId: string;
  userId: string;
  email: string;
  role: "read" | "read_write" | "admin";
  learningModeOverride: boolean | null;
};

