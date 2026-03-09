import {
  executeNextStepAction as executeNextStepActionRequest,
  previewNextStepAction as previewNextStepActionRequest,
} from "../../shared/api/sdk";

import type { NextStepActionPreview, NextStepActionQuery } from "./types";

function toRequestBody(query: NextStepActionQuery) {
  return {
    repository_full_name: query.repositoryFullName,
    issue_number: typeof query.issueNumber === "number" ? query.issueNumber : undefined,
    pull_request_number: typeof query.pullRequestNumber === "number" ? query.pullRequestNumber : undefined,
    action_kind: query.actionKind,
    target_label: query.targetLabel,
  };
}

export async function previewNextStepAction(query: NextStepActionQuery): Promise<NextStepActionPreview> {
  const resp = await previewNextStepActionRequest({
    body: toRequestBody(query),
    throwOnError: true,
  });
  return resp.data;
}

export async function executeNextStepAction(query: NextStepActionQuery): Promise<NextStepActionPreview> {
  const resp = await executeNextStepActionRequest({
    body: toRequestBody(query),
    throwOnError: true,
  });
  return resp.data;
}
