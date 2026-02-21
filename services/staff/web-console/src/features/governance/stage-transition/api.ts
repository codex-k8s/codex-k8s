import { transitionIssueStageLabel as transitionIssueStageLabelRequest } from "../../../shared/api/sdk";

import type { TransitionIssueStageLabelResponse } from "../../../shared/api/generated";

export type StageTransitionParams = {
  repositoryFullName: string;
  issueNumber: number;
  targetLabel: string;
};

export async function transitionIssueStageLabel(params: StageTransitionParams): Promise<TransitionIssueStageLabelResponse> {
  const response = await transitionIssueStageLabelRequest({
    body: {
      repository_full_name: params.repositoryFullName,
      issue_number: params.issueNumber,
      target_label: params.targetLabel,
    },
    throwOnError: true,
  });
  return response.data;
}
