import { isValid, parseISO } from "date-fns";

import { ApiError } from "../api/errors.ts";

export type RuntimeDeployActionKind = "cancel" | "stop";

export type RuntimeDeployLeaseStateInput = {
  lease_owner?: string | null;
  lease_until?: string | null;
};

export type RuntimeDeployLeaseStatus = "active" | "expired" | "missing";

export function resolveRuntimeDeployLeaseStatus(
  task: RuntimeDeployLeaseStateInput | null | undefined,
  now: Date = new Date(),
): RuntimeDeployLeaseStatus {
  const leaseOwner = String(task?.lease_owner || "").trim();
  const leaseUntil = String(task?.lease_until || "").trim();
  if (!leaseOwner || !leaseUntil) {
    return "missing";
  }

  const leaseUntilDate = parseISO(leaseUntil);
  if (!isValid(leaseUntilDate)) {
    return "missing";
  }

  return leaseUntilDate.getTime() > now.getTime() ? "active" : "expired";
}

export function hasActiveRuntimeDeployLease(
  task: RuntimeDeployLeaseStateInput | null | undefined,
  now: Date = new Date(),
): boolean {
  return resolveRuntimeDeployLeaseStatus(task, now) === "active";
}

export function localizeRuntimeDeployActionError(
  error: ApiError,
  action: RuntimeDeployActionKind,
): ApiError {
  if (error.code !== "failed_precondition") {
    return error;
  }

  return new ApiError({
    kind: error.kind,
    status: error.status,
    code: error.code,
    field: error.field,
    messageKey: action === "stop"
      ? "pages.runtimeDeployTaskDetails.stopUnavailableError"
      : "pages.runtimeDeployTaskDetails.cancelUnavailableError",
  });
}
