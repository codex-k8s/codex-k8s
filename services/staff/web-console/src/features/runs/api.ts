import { http } from "../../shared/api/http";

import type { FlowEventDto, LearningFeedbackDto, RunDto } from "./types";

type ItemsResponse<T> = { items: T[] };

export async function listRuns(limit = 200): Promise<RunDto[]> {
  const resp = await http.get("/api/v1/staff/runs", { params: { limit } });
  return (resp.data as ItemsResponse<RunDto>).items ?? [];
}

export async function listRunEvents(runId: string, limit = 500): Promise<FlowEventDto[]> {
  const resp = await http.get(`/api/v1/staff/runs/${runId}/events`, { params: { limit } });
  return (resp.data as ItemsResponse<FlowEventDto>).items ?? [];
}

export async function listRunLearningFeedback(runId: string, limit = 200): Promise<LearningFeedbackDto[]> {
  const resp = await http.get(`/api/v1/staff/runs/${runId}/learning-feedback`, { params: { limit } });
  return (resp.data as ItemsResponse<LearningFeedbackDto>).items ?? [];
}

