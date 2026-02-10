import { http } from "../../shared/api/http";

import type { ProjectDto, ProjectMemberDto, RepositoryBindingDto } from "./types";

type ItemsResponse<T> = { items: T[] };

export async function listProjects(limit = 200): Promise<ProjectDto[]> {
  const resp = await http.get("/api/v1/staff/projects", { params: { limit } });
  return (resp.data as ItemsResponse<ProjectDto>).items ?? [];
}

export async function upsertProject(slug: string, name: string): Promise<void> {
  await http.post("/api/v1/staff/projects", { slug, name });
}

export async function listProjectRepositories(projectId: string, limit = 200): Promise<RepositoryBindingDto[]> {
  const resp = await http.get(`/api/v1/staff/projects/${projectId}/repositories`, { params: { limit } });
  return (resp.data as ItemsResponse<RepositoryBindingDto>).items ?? [];
}

export async function upsertProjectRepository(params: {
  projectId: string;
  provider: string;
  owner: string;
  name: string;
  token: string;
  servicesYamlPath: string;
}): Promise<void> {
  await http.post(`/api/v1/staff/projects/${params.projectId}/repositories`, {
    provider: params.provider,
    owner: params.owner,
    name: params.name,
    token: params.token,
    services_yaml_path: params.servicesYamlPath,
  });
}

export async function deleteProjectRepository(projectId: string, repositoryId: string): Promise<void> {
  await http.delete(`/api/v1/staff/projects/${projectId}/repositories/${repositoryId}`);
}

export async function listProjectMembers(projectId: string, limit = 200): Promise<ProjectMemberDto[]> {
  const resp = await http.get(`/api/v1/staff/projects/${projectId}/members`, { params: { limit } });
  return (resp.data as ItemsResponse<ProjectMemberDto>).items ?? [];
}

export async function upsertProjectMember(projectId: string, userId: string, role: string): Promise<void> {
  await http.post(`/api/v1/staff/projects/${projectId}/members`, { user_id: userId, role });
}

export async function setProjectMemberLearningModeOverride(projectId: string, userId: string, enabled: boolean | null): Promise<void> {
  await http.put(`/api/v1/staff/projects/${projectId}/members/${userId}/learning-mode`, { enabled });
}

