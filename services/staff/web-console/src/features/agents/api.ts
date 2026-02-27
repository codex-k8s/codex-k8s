import {
  activatePromptTemplateVersion as activatePromptTemplateVersionRequest,
  createPromptTemplateVersion as createPromptTemplateVersionRequest,
  diffPromptTemplateVersions as diffPromptTemplateVersionsRequest,
  getAgent as getAgentRequest,
  listAgents as listAgentsRequest,
  listPromptTemplateAuditEvents as listPromptTemplateAuditEventsRequest,
  listPromptTemplateKeys as listPromptTemplateKeysRequest,
  listPromptTemplateVersions as listPromptTemplateVersionsRequest,
  previewPromptTemplate as previewPromptTemplateRequest,
  updateAgentSettings as updateAgentSettingsRequest,
} from "../../shared/api/sdk";
import type {
  Agent,
  AgentSettings,
  PromptTemplateAuditEvent,
  PromptTemplateDiffResponse,
  PromptTemplateKey,
  PromptTemplateSource,
  PromptTemplateVersion,
  PreviewPromptTemplateResponse,
} from "./types";
import type { PromptTemplateAuditFilters, PromptTemplateKeyFilters } from "./types";

export async function listAgents(limit = 200): Promise<Agent[]> {
  const resp = await listAgentsRequest({
    query: { limit },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

export async function getAgent(agentId: string): Promise<Agent> {
  const resp = await getAgentRequest({
    path: { agent_id: agentId },
    throwOnError: true,
  });
  return resp.data;
}

export async function updateAgentSettings(params: {
  agentId: string;
  expectedVersion: number;
  settings: AgentSettings;
}): Promise<Agent> {
  const resp = await updateAgentSettingsRequest({
    path: { agent_id: params.agentId },
    body: {
      expected_version: params.expectedVersion,
      settings: params.settings,
    },
    throwOnError: true,
  });
  return resp.data;
}

export async function listPromptTemplateKeys(filters: PromptTemplateKeyFilters = {}): Promise<PromptTemplateKey[]> {
  const resp = await listPromptTemplateKeysRequest({
    query: {
      limit: filters.limit ?? 500,
      scope: filters.scope,
      project_id: filters.projectId?.trim() || undefined,
      role: filters.role?.trim() || undefined,
      kind: filters.kind,
      locale: filters.locale?.trim() || undefined,
    },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

export async function listPromptTemplateVersions(templateKey: string, limit = 200): Promise<PromptTemplateVersion[]> {
  const resp = await listPromptTemplateVersionsRequest({
    path: { template_key: templateKey },
    query: { limit },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

export async function createPromptTemplateVersion(params: {
  templateKey: string;
  bodyMarkdown: string;
  expectedVersion: number;
  source?: PromptTemplateSource;
  changeReason?: string;
}): Promise<PromptTemplateVersion> {
  const resp = await createPromptTemplateVersionRequest({
    path: { template_key: params.templateKey },
    body: {
      body_markdown: params.bodyMarkdown,
      expected_version: params.expectedVersion,
      source: params.source,
      change_reason: params.changeReason?.trim() || undefined,
    },
    throwOnError: true,
  });
  return resp.data;
}

export async function activatePromptTemplateVersion(params: {
  templateKey: string;
  version: number;
  expectedVersion: number;
  changeReason: string;
}): Promise<PromptTemplateVersion> {
  const resp = await activatePromptTemplateVersionRequest({
    path: {
      template_key: params.templateKey,
      version: params.version,
    },
    body: {
      expected_version: params.expectedVersion,
      change_reason: params.changeReason.trim(),
    },
    throwOnError: true,
  });
  return resp.data;
}

export async function previewPromptTemplate(params: {
  templateKey: string;
  projectId?: string;
  version?: number;
}): Promise<PreviewPromptTemplateResponse> {
  const resp = await previewPromptTemplateRequest({
    path: { template_key: params.templateKey },
    body: {
      project_id: params.projectId?.trim() || undefined,
      version: params.version,
    },
    throwOnError: true,
  });
  return resp.data;
}

export async function diffPromptTemplateVersions(params: {
  templateKey: string;
  fromVersion: number;
  toVersion: number;
}): Promise<PromptTemplateDiffResponse> {
  const resp = await diffPromptTemplateVersionsRequest({
    path: { template_key: params.templateKey },
    query: {
      from_version: params.fromVersion,
      to_version: params.toVersion,
    },
    throwOnError: true,
  });
  return resp.data;
}

export async function listPromptTemplateAuditEvents(filters: PromptTemplateAuditFilters = {}): Promise<PromptTemplateAuditEvent[]> {
  const resp = await listPromptTemplateAuditEventsRequest({
    query: {
      limit: filters.limit ?? 200,
      project_id: filters.projectId?.trim() || undefined,
      template_key: filters.templateKey?.trim() || undefined,
      actor_id: filters.actorId?.trim() || undefined,
    },
    throwOnError: true,
  });
  return resp.data.items ?? [];
}

