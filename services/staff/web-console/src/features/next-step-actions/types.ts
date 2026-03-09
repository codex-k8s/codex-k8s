import type {
  NextStepActionRequest as NextStepActionRequestDto,
  NextStepActionResponse as NextStepActionResponseDto,
} from "../../shared/api/generated";

export type NextStepActionKind = "issue_stage_transition" | "pull_request_label_add";

export type NextStepActionDisplayVariant =
  | "revise"
  | "full_flow"
  | "shortened_flow"
  | "very_short_flow"
  | "full_or_shortened_flow"
  | "full_or_very_short_flow"
  | "shortened_or_very_short_flow"
  | "all_flows"
  | "reviewer"
  | "rethink"
  | "doc_audit"
  | "self_improve"
  | "prepare_plan"
  | "go_to_dev"
  | "go_to_qa"
  | "restart_full"
  | "restart_shortened"
  | "restart_very_short";

export type NextStepActionQuery = {
  repositoryFullName: string;
  issueNumber?: number;
  pullRequestNumber?: number;
  actionKind: NextStepActionKind;
  targetLabel: string;
  displayVariant: NextStepActionDisplayVariant;
};

export type NextStepActionRequest = NextStepActionRequestDto;
export type NextStepActionPreview = NextStepActionResponseDto;
