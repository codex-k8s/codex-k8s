import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { listProjectMembers, setProjectMemberLearningModeOverride, upsertProjectMember } from "./api";
import type { ProjectMember } from "./types";

export const useProjectMembersStore = defineStore("projectMembers", {
  state: () => ({
    projectId: "" as string,
    items: [] as ProjectMember[],
    loading: false,
    error: null as ApiError | null,
    saving: false,
  }),
  actions: {
    async load(projectId: string): Promise<void> {
      this.projectId = projectId;
      this.loading = true;
      this.error = null;
      try {
        const dtos = await listProjectMembers(projectId);
        this.items = dtos.map((m) => ({
          projectId: m.project_id,
          userId: m.user_id,
          email: m.email,
          role: m.role,
          learningModeOverride: m.learning_mode_override,
        }));
      } catch (e) {
        this.error = normalizeApiError(e);
      } finally {
        this.loading = false;
      }
    },

    async save(member: { userId: string; role: string; learningModeOverride: boolean | null }): Promise<void> {
      if (!this.projectId) return;
      this.saving = true;
      this.error = null;
      try {
        await upsertProjectMember(this.projectId, member.userId, member.role);
        await setProjectMemberLearningModeOverride(this.projectId, member.userId, member.learningModeOverride);
        await this.load(this.projectId);
      } catch (e) {
        this.error = normalizeApiError(e);
      } finally {
        this.saving = false;
      }
    },
  },
});

