import type {
  FlowEvent as FlowEventAPIModel,
  LearningFeedback as LearningFeedbackAPIModel,
  Run as RunAPIModel,
} from "../../shared/api/generated";

export type RunDto = RunAPIModel;

export type Run = {
  id: string;
  correlationId: string;
  projectId: string | null;
  projectSlug: string;
  projectName: string;
  status: string;
  createdAt: string;
  startedAt: string | null;
  finishedAt: string | null;
};

export type FlowEventDto = FlowEventAPIModel;

export type FlowEvent = {
  correlationId: string;
  eventType: string;
  createdAt: string;
  payloadJson: string;
};

export type LearningFeedbackDto = LearningFeedbackAPIModel;

export type LearningFeedback = {
  id: number;
  runId: string;
  kind: string;
  explanation: string;
  createdAt: string;
};
