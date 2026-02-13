import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import {
  deleteRunNamespace,
  getRun,
  listPendingApprovals,
  listRunEvents,
  listRuns,
  resolveApprovalDecision,
} from "./api";
import type {
  ApprovalRequest,
  FlowEvent,
  ResolveApprovalDecisionResponse,
  Run,
  RunNamespaceCleanupResponse,
} from "./types";

const errorAutoHideMs = 5000;

export const useRunsStore = defineStore("runs", {
  state: () => ({
    items: [] as Run[],
    pendingApprovals: [] as ApprovalRequest[],
    loading: false,
    approvalsLoading: false,
    resolvingApprovalID: null as number | null,
    error: null as ApiError | null,
    approvalsError: null as ApiError | null,
    errorTimerId: null as number | null,
    approvalsErrorTimerId: null as number | null,
  }),
  actions: {
    clearErrorTimer(): void {
      if (this.errorTimerId !== null) {
        window.clearTimeout(this.errorTimerId);
        this.errorTimerId = null;
      }
    },
    scheduleErrorHide(): void {
      this.clearErrorTimer();
      this.errorTimerId = window.setTimeout(() => {
        this.error = null;
        this.errorTimerId = null;
      }, errorAutoHideMs);
    },
    clearApprovalsErrorTimer(): void {
      if (this.approvalsErrorTimerId !== null) {
        window.clearTimeout(this.approvalsErrorTimerId);
        this.approvalsErrorTimerId = null;
      }
    },
    scheduleApprovalsErrorHide(): void {
      this.clearApprovalsErrorTimer();
      this.approvalsErrorTimerId = window.setTimeout(() => {
        this.approvalsError = null;
        this.approvalsErrorTimerId = null;
      }, errorAutoHideMs);
    },
    async load(): Promise<void> {
      this.loading = true;
      this.error = null;
      try {
        this.items = await listRuns();
      } catch (e) {
        this.error = normalizeApiError(e);
        this.scheduleErrorHide();
      } finally {
        this.loading = false;
      }
    },
    async loadPendingApprovals(): Promise<void> {
      this.approvalsLoading = true;
      this.approvalsError = null;
      try {
        this.pendingApprovals = await listPendingApprovals();
      } catch (e) {
        this.approvalsError = normalizeApiError(e);
        this.scheduleApprovalsErrorHide();
      } finally {
        this.approvalsLoading = false;
      }
    },
    async resolvePendingApproval(
      approvalRequestId: number,
      decision: "approved" | "denied" | "expired" | "failed",
      reason = "",
    ): Promise<ResolveApprovalDecisionResponse | null> {
      this.resolvingApprovalID = approvalRequestId;
      this.approvalsError = null;
      try {
        const response = await resolveApprovalDecision(approvalRequestId, decision, reason);
        await this.loadPendingApprovals();
        return response;
      } catch (e) {
        this.approvalsError = normalizeApiError(e);
        this.scheduleApprovalsErrorHide();
        return null;
      } finally {
        this.resolvingApprovalID = null;
      }
    },
  },
});

export const useRunDetailsStore = defineStore("runDetails", {
  state: () => ({
    runId: "" as string,
    run: null as Run | null,
    loading: false,
    error: null as ApiError | null,
    events: [] as FlowEvent[],
    deletingNamespace: false,
    deleteNamespaceError: null as ApiError | null,
    namespaceDeleteResult: null as RunNamespaceCleanupResponse | null,
    errorTimerId: null as number | null,
    deleteNamespaceErrorTimerId: null as number | null,
  }),
  actions: {
    clearErrorTimer(timerField: "errorTimerId" | "deleteNamespaceErrorTimerId"): void {
      const timerId = this[timerField];
      if (timerId !== null) {
        window.clearTimeout(timerId);
        this[timerField] = null;
      }
    },
    scheduleErrorHide(errorField: "error" | "deleteNamespaceError", timerField: "errorTimerId" | "deleteNamespaceErrorTimerId"): void {
      this.clearErrorTimer(timerField);
      this[timerField] = window.setTimeout(() => {
        this[errorField] = null;
        this[timerField] = null;
      }, errorAutoHideMs);
    },
    async load(runId: string): Promise<void> {
      this.runId = runId;
      this.loading = true;
      this.error = null;
      try {
        const [run, events] = await Promise.all([
          getRun(runId),
          listRunEvents(runId),
        ]);
        this.run = run;
        this.events = [...events].sort((a, b) => (a.created_at < b.created_at ? 1 : a.created_at > b.created_at ? -1 : 0));
      } catch (e) {
        this.run = null;
        this.events = [];
        this.error = normalizeApiError(e);
        this.scheduleErrorHide("error", "errorTimerId");
      } finally {
        this.loading = false;
      }
    },

    async deleteNamespace(runId: string): Promise<void> {
      this.deletingNamespace = true;
      this.deleteNamespaceError = null;
      this.namespaceDeleteResult = null;
      try {
        this.namespaceDeleteResult = await deleteRunNamespace(runId);
        await this.load(runId);
      } catch (e) {
        this.deleteNamespaceError = normalizeApiError(e);
        this.scheduleErrorHide("deleteNamespaceError", "deleteNamespaceErrorTimerId");
      } finally {
        this.deletingNamespace = false;
      }
    },
  },
});
