import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { deleteProjectRepository, listProjectRepositories, upsertProjectRepository } from "./api";
import type { RepositoryBinding } from "./types";

export const useProjectRepositoriesStore = defineStore("projectRepositories", {
  state: () => ({
    projectId: "" as string,
    items: [] as RepositoryBinding[],
    loading: false,
    error: null as ApiError | null,
    attaching: false,
    attachError: null as ApiError | null,
    removing: false,
  }),
  actions: {
    async load(projectId: string): Promise<void> {
      this.projectId = projectId;
      this.loading = true;
      this.error = null;
      try {
        const dtos = await listProjectRepositories(projectId);
        this.items = dtos.map((r) => ({
          id: r.id,
          projectId: r.project_id,
          provider: r.provider,
          externalId: r.external_id,
          owner: r.owner,
          name: r.name,
          servicesYamlPath: r.services_yaml_path,
        }));
      } catch (e) {
        this.error = normalizeApiError(e);
      } finally {
        this.loading = false;
      }
    },

    async attach(params: { owner: string; name: string; token: string; servicesYamlPath: string }): Promise<void> {
      if (!this.projectId) return;
      this.attaching = true;
      this.attachError = null;
      try {
        await upsertProjectRepository({
          projectId: this.projectId,
          provider: "github",
          owner: params.owner,
          name: params.name,
          token: params.token,
          servicesYamlPath: params.servicesYamlPath,
        });
        await this.load(this.projectId);
      } catch (e) {
        this.attachError = normalizeApiError(e);
      } finally {
        this.attaching = false;
      }
    },

    async remove(repositoryId: string): Promise<void> {
      if (!this.projectId) return;
      this.removing = true;
      try {
        await deleteProjectRepository(this.projectId, repositoryId);
        await this.load(this.projectId);
      } catch (e) {
        this.error = normalizeApiError(e);
      } finally {
        this.removing = false;
      }
    },
  },
});

