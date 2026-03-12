import test from "node:test";
import assert from "node:assert/strict";

import { ApiError } from "../api/errors.ts";
import {
  hasActiveRuntimeDeployLease,
  localizeRuntimeDeployActionError,
  resolveRuntimeDeployLeaseStatus,
} from "./runtime-deploy-task-actions.ts";

test("resolveRuntimeDeployLeaseStatus returns active for future lease with owner", () => {
  const status = resolveRuntimeDeployLeaseStatus(
    {
      lease_owner: "worker-1",
      lease_until: "2026-03-12T10:05:00Z",
    },
    new Date("2026-03-12T10:00:00Z"),
  );

  assert.equal(status, "active");
  assert.equal(
    hasActiveRuntimeDeployLease(
      {
        lease_owner: "worker-1",
        lease_until: "2026-03-12T10:05:00Z",
      },
      new Date("2026-03-12T10:00:00Z"),
    ),
    true,
  );
});

test("resolveRuntimeDeployLeaseStatus returns expired when lease time is in the past", () => {
  const status = resolveRuntimeDeployLeaseStatus(
    {
      lease_owner: "worker-1",
      lease_until: "2026-03-12T09:55:00Z",
    },
    new Date("2026-03-12T10:00:00Z"),
  );

  assert.equal(status, "expired");
});

test("resolveRuntimeDeployLeaseStatus returns missing without complete lease metadata", () => {
  assert.equal(
    resolveRuntimeDeployLeaseStatus(
      {
        lease_owner: null,
        lease_until: "2026-03-12T10:05:00Z",
      },
      new Date("2026-03-12T10:00:00Z"),
    ),
    "missing",
  );
  assert.equal(
    resolveRuntimeDeployLeaseStatus(
      {
        lease_owner: "worker-1",
        lease_until: null,
      },
      new Date("2026-03-12T10:00:00Z"),
    ),
    "missing",
  );
});

test("localizeRuntimeDeployActionError remaps failed_precondition for operator-friendly messages", () => {
  const baseError = new ApiError({
    kind: "http",
    status: 409,
    code: "failed_precondition",
    messageKey: "errors.failedPrecondition",
  });

  assert.equal(
    localizeRuntimeDeployActionError(baseError, "cancel").messageKey,
    "pages.runtimeDeployTaskDetails.cancelUnavailableError",
  );
  assert.equal(
    localizeRuntimeDeployActionError(baseError, "stop").messageKey,
    "pages.runtimeDeployTaskDetails.stopUnavailableError",
  );
});

test("localizeRuntimeDeployActionError keeps unrelated errors intact", () => {
  const baseError = new ApiError({
    kind: "http",
    status: 500,
    code: "unknown",
    messageKey: "errors.unknown",
  });

  assert.equal(localizeRuntimeDeployActionError(baseError, "stop"), baseError);
});
