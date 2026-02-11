import {
  deleteProject as deleteProjectRequest,
  deleteProjectMember as deleteProjectMemberRequest,
  deleteProjectRepository as deleteProjectRepositoryRequest,
  getProject as getProjectRequest,
  listProjectMembers as listProjectMembersRequest,
  listProjectRepositories as listProjectRepositoriesRequest,
  listProjects as listProjectsRequest,
  setProjectMemberLearningModeOverride as setProjectMemberLearningModeOverrideRequest,
  upsertProject as upsertProjectRequest,
  upsertProjectMember as upsertProjectMemberRequest,
  upsertProjectRepository as upsertProjectRepositoryRequest,
} from "../../shared/api/sdk";
import type { Project, ProjectMember, RepositoryBinding } from "./types";

export async function listProjects(limit = 200): Promise<Project[]> {
  const resp = await listProjectsRequest({ query: { limit }, throwOnError: true });
  return resp.data.items ?? [];
}

export async function getProject(projectId: string): Promise<Project> {
  const resp = await getProjectRequest({ path: { project_id: projectId }, throwOnError: true });
  return resp.data;
}

export async function upsertProject(slug: string, name: string): Promise<void> {
  await upsertProjectRequest({ body: { slug, name }, throwOnError: true });
}

export async function deleteProject(projectId: string): Promise<void> {
  await deleteProjectRequest({ path: { project_id: projectId }, throwOnError: true });
}

export async function listProjectRepositories(projectId: string, limit = 200): Promise<RepositoryBinding[]> {
  const resp = await listProjectRepositoriesRequest({
    path: { project_id: projectId },
    query: { limit },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

export async function upsertProjectRepository(params: {
  projectId: string;
  provider: string;
  owner: string;
  name: string;
  token: string;
  servicesYamlPath: string;
}): Promise<void> {
  await upsertProjectRepositoryRequest({
    path: { project_id: params.projectId },
    body: {
      provider: params.provider,
      owner: params.owner,
      name: params.name,
      token: params.token,
      services_yaml_path: params.servicesYamlPath,
    },
    throwOnError: true,
  });
}

export async function deleteProjectRepository(projectId: string, repositoryId: string): Promise<void> {
  await deleteProjectRepositoryRequest({
    path: { project_id: projectId, repository_id: repositoryId },
    throwOnError: true,
  });
}

export async function listProjectMembers(projectId: string, limit = 200): Promise<ProjectMember[]> {
  const resp = await listProjectMembersRequest({
    path: { project_id: projectId },
    query: { limit },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

export async function upsertProjectMember(projectId: string, userId: string, role: ProjectMember["role"]): Promise<void> {
  await upsertProjectMemberRequest({
    path: { project_id: projectId },
    body: { user_id: userId, role },
    throwOnError: true,
  });
}

export async function upsertProjectMemberByEmail(projectId: string, email: string, role: ProjectMember["role"]): Promise<void> {
  await upsertProjectMemberRequest({
    path: { project_id: projectId },
    body: { email, role },
    throwOnError: true,
  });
}

export async function deleteProjectMember(projectId: string, userId: string): Promise<void> {
  await deleteProjectMemberRequest({
    path: { project_id: projectId, user_id: userId },
    throwOnError: true,
  });
}

export async function setProjectMemberLearningModeOverride(projectId: string, userId: string, enabled: boolean | null): Promise<void> {
  await setProjectMemberLearningModeOverrideRequest({
    path: { project_id: projectId, user_id: userId },
    body: { enabled },
    throwOnError: true,
  });
}
