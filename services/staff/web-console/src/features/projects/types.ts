import type {
  Project as ProjectAPIModel,
  ProjectMember as ProjectMemberAPIModel,
  RepositoryBinding as RepositoryBindingAPIModel,
} from "../../shared/api/generated";

export type ProjectDto = ProjectAPIModel;

export type Project = {
  id: string;
  slug: string;
  name: string;
  role: string;
};

export type RepositoryBindingDto = RepositoryBindingAPIModel;

export type RepositoryBinding = {
  id: string;
  projectId: string;
  provider: string;
  externalId: number;
  owner: string;
  name: string;
  servicesYamlPath: string;
};

export type ProjectMemberDto = ProjectMemberAPIModel;

export type ProjectMember = {
  projectId: string;
  userId: string;
  email: string;
  role: "read" | "read_write" | "admin";
  learningModeOverride: boolean | null;
};
