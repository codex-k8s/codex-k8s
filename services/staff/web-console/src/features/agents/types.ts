export type {
  ActivatePromptTemplateVersionRequest,
  Agent,
  AgentSettings,
  PromptTemplateAuditEvent,
  PromptTemplateDiffResponse,
  PromptTemplateKey,
  PromptTemplateScope,
  PromptTemplateSource,
  PromptTemplateStatus,
  PromptTemplateVersion,
  PreviewPromptTemplateResponse,
} from "../../shared/api/generated";

export type PromptTemplateKeyFilters = {
  limit?: number;
  scope?: "global" | "project";
  projectId?: string;
  role?: string;
  kind?: "work" | "revise";
  locale?: string;
};

export type PromptTemplateAuditFilters = {
  limit?: number;
  projectId?: string;
  templateKey?: string;
  actorId?: string;
};

