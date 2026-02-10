export type RunDto = {
  id: string;
  correlation_id: string;
  project_id: string | null;
  project_slug: string;
  project_name: string;
  status: string;
  created_at: string;
  started_at: string | null;
  finished_at: string | null;
};

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

export type FlowEventDto = {
  correlation_id: string;
  event_type: string;
  created_at: string;
  payload_json: string;
};

export type FlowEvent = {
  correlationId: string;
  eventType: string;
  createdAt: string;
  payloadJson: string;
};

export type LearningFeedbackDto = {
  id: number;
  run_id: string;
  repository_id: string | null;
  pr_number: number | null;
  file_path: string | null;
  line: number | null;
  kind: string;
  explanation: string;
  created_at: string;
};

export type LearningFeedback = {
  id: number;
  runId: string;
  kind: string;
  explanation: string;
  createdAt: string;
};
