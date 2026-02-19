import { defineStore } from "pinia";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import {
  deleteRunNamespace,
  getRun,
  getRunAccessKeyStatus,
  getRunLogs,
  listPendingApprovals,
  listRunJobs,
  listRunEvents,
  listRunWaits,
  listRuns,
  regenerateRunAccessKey,
  resolveApprovalDecision,
  revokeRunAccessKey,
  type RunListFilters,
  type RunWaitFilters,
} from "./api";
import type {
  ApprovalRequest,
  FlowEvent,
  RunAccessKeyStatus,
  ResolveApprovalDecisionResponse,
  Run,
  RunLogs,
  RunNamespaceCleanupResponse,
} from "./types";

const errorAutoHideMs = 5000;

export const useRunsStore = defineStore("runs", {
  state: () => ({
    items: [] as Run[],
    runningJobs: [] as Run[],
    waitQueue: [] as Run[],
    pendingApprovals: [] as ApprovalRequest[],
    jobsFilters: {
      triggerKind: "",
      status: "",
      agentKey: "",
    } as RunListFilters,
    waitsFilters: {
      triggerKind: "",
      status: "",
      agentKey: "",
      waitState: "",
    } as RunWaitFilters,
    loading: false,
    jobsLoading: false,
    waitsLoading: false,
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
    async loadRuntimeViews(): Promise<void> {
      await Promise.all([this.loadRunJobs(), this.loadRunWaits()]);
    },
    async loadRunJobs(): Promise<void> {
      this.jobsLoading = true;
      this.error = null;
      try {
        this.runningJobs = await listRunJobs(this.jobsFilters, 200);
      } catch (e) {
        this.error = normalizeApiError(e);
        this.scheduleErrorHide();
      } finally {
        this.jobsLoading = false;
      }
    },
    async loadRunWaits(): Promise<void> {
      this.waitsLoading = true;
      this.error = null;
      try {
        this.waitQueue = await listRunWaits(this.waitsFilters, 200);
      } catch (e) {
        this.error = normalizeApiError(e);
        this.scheduleErrorHide();
      } finally {
        this.waitsLoading = false;
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
    logs: null as RunLogs | null,
    runAccessKeyStatus: null as RunAccessKeyStatus | null,
    runAccessKeyPlaintext: "" as string,
    regeneratingAccessKey: false,
    revokingAccessKey: false,
    accessKeyError: null as ApiError | null,
    deletingNamespace: false,
    deleteNamespaceError: null as ApiError | null,
    namespaceDeleteResult: null as RunNamespaceCleanupResponse | null,
    errorTimerId: null as number | null,
    accessKeyErrorTimerId: null as number | null,
    deleteNamespaceErrorTimerId: null as number | null,
  }),
  actions: {
    clearErrorTimer(timerField: "errorTimerId" | "accessKeyErrorTimerId" | "deleteNamespaceErrorTimerId"): void {
      const timerId = this[timerField];
      if (timerId !== null) {
        window.clearTimeout(timerId);
        this[timerField] = null;
      }
    },
    scheduleErrorHide(
      errorField: "error" | "accessKeyError" | "deleteNamespaceError",
      timerField: "errorTimerId" | "accessKeyErrorTimerId" | "deleteNamespaceErrorTimerId",
    ): void {
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
        const [run, events, logs] = await Promise.all([
          getRun(runId),
          listRunEvents(runId),
          getRunLogs(runId, 200),
        ]);
        this.run = run;
        this.events = [...events].sort((a, b) => (a.created_at < b.created_at ? 1 : a.created_at > b.created_at ? -1 : 0));
        this.logs = logs;
        this.runAccessKeyStatus = await getRunAccessKeyStatus(runId);
        this.runAccessKeyPlaintext = "";
      } catch (e) {
        this.run = null;
        this.events = [];
        this.logs = null;
        this.runAccessKeyStatus = null;
        this.runAccessKeyPlaintext = "";
        this.error = normalizeApiError(e);
        this.scheduleErrorHide("error", "errorTimerId");
      } finally {
        this.loading = false;
      }
    },

    async refreshLogs(runId: string, tailLines = 200): Promise<void> {
      try {
        this.logs = await getRunLogs(runId, tailLines);
      } catch (e) {
        this.error = normalizeApiError(e);
        this.scheduleErrorHide("error", "errorTimerId");
      }
    },

    async regenerateAccessKey(runId: string, ttlSeconds?: number): Promise<void> {
      this.regeneratingAccessKey = true;
      this.accessKeyError = null;
      try {
        const issued = await regenerateRunAccessKey(runId, ttlSeconds);
        this.runAccessKeyStatus = issued.status;
        this.runAccessKeyPlaintext = issued.access_key;
      } catch (e) {
        this.accessKeyError = normalizeApiError(e);
        this.scheduleErrorHide("accessKeyError", "accessKeyErrorTimerId");
      } finally {
        this.regeneratingAccessKey = false;
      }
    },

    async revokeAccessKey(runId: string): Promise<void> {
      this.revokingAccessKey = true;
      this.accessKeyError = null;
      try {
        this.runAccessKeyStatus = await revokeRunAccessKey(runId);
        this.runAccessKeyPlaintext = "";
      } catch (e) {
        this.accessKeyError = normalizeApiError(e);
        this.scheduleErrorHide("accessKeyError", "accessKeyErrorTimerId");
      } finally {
        this.revokingAccessKey = false;
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
