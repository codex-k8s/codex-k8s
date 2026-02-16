import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { importDocset, listDocsetGroups, runRepositoryPreflight, syncDocset, upsertRepositoryBotParams, deleteProjectRepository, listProjectRepositories, upsertProjectRepository } from "./api";
import type { DocsetGroup, ImportDocsetResponse, RepositoryBinding, RunRepositoryPreflightResponse, SyncDocsetResponse } from "./types";

export const useProjectRepositoriesStore = defineStore("projectRepositories", {
  state: () => ({
    projectId: "" as string,
    items: [] as RepositoryBinding[],
    loading: false,
    error: null as ApiError | null,
    attaching: false,
    attachError: null as ApiError | null,
    removing: false,
    botUpdating: false,
    botUpdateError: null as ApiError | null,
    preflighting: false,
    preflightError: null as ApiError | null,
    docsetLoadingGroups: false,
    docsetGroupsError: null as ApiError | null,
    docsetImporting: false,
    docsetImportError: null as ApiError | null,
    docsetSyncing: false,
    docsetSyncError: null as ApiError | null,
  }),
  actions: {
    async load(projectId: string): Promise<void> {
      this.projectId = projectId;
      this.loading = true;
      this.error = null;
      try {
        this.items = await listProjectRepositories(projectId);
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

    async upsertBotParams(repositoryId: string, params: { botToken: string | null; botUsername: string | null; botEmail: string | null }): Promise<void> {
      if (!this.projectId) return;
      this.botUpdating = true;
      this.botUpdateError = null;
      try {
        await upsertRepositoryBotParams({
          projectId: this.projectId,
          repositoryId,
          botToken: params.botToken,
          botUsername: params.botUsername,
          botEmail: params.botEmail,
        });
        await this.load(this.projectId);
      } catch (e) {
        this.botUpdateError = normalizeApiError(e);
      } finally {
        this.botUpdating = false;
      }
    },

    async runPreflight(repositoryId: string): Promise<RunRepositoryPreflightResponse | null> {
      if (!this.projectId) return null;
      this.preflighting = true;
      this.preflightError = null;
      try {
        return await runRepositoryPreflight(this.projectId, repositoryId);
      } catch (e) {
        this.preflightError = normalizeApiError(e);
        return null;
      } finally {
        this.preflighting = false;
      }
    },

    async loadDocsetGroups(params: { docsetRef?: string; locale?: "ru" | "en" }): Promise<DocsetGroup[]> {
      this.docsetLoadingGroups = true;
      this.docsetGroupsError = null;
      try {
        return await listDocsetGroups(params);
      } catch (e) {
        this.docsetGroupsError = normalizeApiError(e);
        return [];
      } finally {
        this.docsetLoadingGroups = false;
      }
    },

    async importDocset(params: { repositoryId: string; docsetRef: string; locale: "ru" | "en"; groupIds: string[] }): Promise<ImportDocsetResponse | null> {
      if (!this.projectId) return null;
      this.docsetImporting = true;
      this.docsetImportError = null;
      try {
        return await importDocset({
          projectId: this.projectId,
          repositoryId: params.repositoryId,
          docsetRef: params.docsetRef,
          locale: params.locale,
          groupIds: params.groupIds,
        });
      } catch (e) {
        this.docsetImportError = normalizeApiError(e);
        return null;
      } finally {
        this.docsetImporting = false;
      }
    },

    async syncDocset(params: { repositoryId: string; docsetRef: string }): Promise<SyncDocsetResponse | null> {
      if (!this.projectId) return null;
      this.docsetSyncing = true;
      this.docsetSyncError = null;
      try {
        return await syncDocset({
          projectId: this.projectId,
          repositoryId: params.repositoryId,
          docsetRef: params.docsetRef,
        });
      } catch (e) {
        this.docsetSyncError = normalizeApiError(e);
        return null;
      } finally {
        this.docsetSyncing = false;
      }
    },
  },
});
