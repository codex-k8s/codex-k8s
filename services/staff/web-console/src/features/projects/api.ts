import {
  deleteProject as deleteProjectRequest,
  deleteProjectMember as deleteProjectMemberRequest,
  deleteProjectRepository as deleteProjectRepositoryRequest,
  getProject as getProjectRequest,
  importDocset as importDocsetRequest,
  listProjectMembers as listProjectMembersRequest,
  listProjectRepositories as listProjectRepositoriesRequest,
  listProjects as listProjectsRequest,
  listDocsetGroups as listDocsetGroupsRequest,
  runRepositoryPreflight as runRepositoryPreflightRequest,
  setProjectMemberLearningModeOverride as setProjectMemberLearningModeOverrideRequest,
  syncDocset as syncDocsetRequest,
  upsertProject as upsertProjectRequest,
  upsertProjectMember as upsertProjectMemberRequest,
  upsertProjectRepository as upsertProjectRepositoryRequest,
  upsertRepositoryBotParams as upsertRepositoryBotParamsRequest,
} from "../../shared/api/sdk";
import type { DocsetGroup, ImportDocsetResponse, Project, ProjectMember, RepositoryBinding, RunRepositoryPreflightResponse, SyncDocsetResponse } from "./types";

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

export async function upsertRepositoryBotParams(params: {
  projectId: string;
  repositoryId: string;
  botToken: string | null;
  botUsername: string | null;
  botEmail: string | null;
}): Promise<void> {
  await upsertRepositoryBotParamsRequest({
    path: { project_id: params.projectId, repository_id: params.repositoryId },
    body: {
      bot_token: params.botToken,
      bot_username: params.botUsername,
      bot_email: params.botEmail,
    },
    throwOnError: true,
  });
}

export async function runRepositoryPreflight(projectId: string, repositoryId: string): Promise<RunRepositoryPreflightResponse> {
  const resp = await runRepositoryPreflightRequest({
    path: { project_id: projectId, repository_id: repositoryId },
    throwOnError: true,
  });
  return resp.data;
}

export async function listDocsetGroups(params: { docsetRef?: string; locale?: "ru" | "en" }): Promise<DocsetGroup[]> {
  const resp = await listDocsetGroupsRequest({
    query: { docset_ref: params.docsetRef, locale: params.locale },
    throwOnError: true,
  });
  return resp.data.groups ?? [];
}

export async function importDocset(params: {
  projectId: string;
  repositoryId: string;
  docsetRef: string;
  locale: "ru" | "en";
  groupIds: string[];
}): Promise<ImportDocsetResponse> {
  const resp = await importDocsetRequest({
    path: { project_id: params.projectId },
    body: {
      repository_id: params.repositoryId,
      docset_ref: params.docsetRef,
      locale: params.locale,
      group_ids: params.groupIds,
    },
    throwOnError: true,
  });
  return resp.data;
}

export async function syncDocset(params: { projectId: string; repositoryId: string; docsetRef: string }): Promise<SyncDocsetResponse> {
  const resp = await syncDocsetRequest({
    path: { project_id: params.projectId },
    body: { repository_id: params.repositoryId, docset_ref: params.docsetRef },
    throwOnError: true,
  });
  return resp.data;
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
