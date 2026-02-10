import type { AxiosError } from "axios";
import { isAxiosError } from "axios";

import { isErrorResponseDto } from "./types";

export type ApiErrorKind = "network" | "timeout" | "http" | "unknown";

export class ApiError extends Error {
  kind: ApiErrorKind;
  status?: number;
  code?: string;
  field?: string;
  messageKey: string;

  constructor(init: { kind: ApiErrorKind; messageKey: string; status?: number; code?: string; field?: string }) {
    super(init.messageKey);
    this.kind = init.kind;
    this.status = init.status;
    this.code = init.code;
    this.field = init.field;
    this.messageKey = init.messageKey;
  }
}

function keyForHttp(status: number | undefined, code: string | undefined, message: string | undefined): string {
  if (status === 403 && message === "email is not allowed") return "errors.emailNotAllowed";
  if (status === 403 && message === "platform admin required") return "errors.platformAdminRequired";
  if (status === 403 && message === "platform owner required") return "errors.platformOwnerRequired";
  if (status === 403 && message === "cannot delete self") return "errors.cannotDeleteSelf";
  if (status === 403 && message === "cannot delete platform admin") return "errors.cannotDeletePlatformAdmin";
  if (status === 403 && message === "cannot remove platform owner from project") return "errors.cannotRemovePlatformOwner";

  if (code === "invalid_argument") return "errors.invalidArgument";
  if (code === "unauthorized") return "errors.unauthorized";
  if (code === "forbidden") return "errors.forbidden";
  if (code === "conflict") return "errors.conflict";
  return "errors.unknown";
}

export function normalizeApiError(err: unknown): ApiError {
  if (err instanceof ApiError) return err;

  if (isAxiosError(err)) {
    const ax = err as AxiosError;

    // Axios uses `code` for network-ish errors too.
    if (ax.code === "ECONNABORTED") {
      return new ApiError({ kind: "timeout", messageKey: "errors.timeout" });
    }
    if (!ax.response) {
      return new ApiError({ kind: "network", messageKey: "errors.network" });
    }

    const status = ax.response.status;
    const data = ax.response.data;
    if (isErrorResponseDto(data)) {
      return new ApiError({
        kind: "http",
        status,
        code: data.code,
        field: data.field,
        messageKey: keyForHttp(status, data.code, data.message),
      });
    }

    return new ApiError({ kind: "http", status, messageKey: keyForHttp(status, undefined, undefined) });
  }

  return new ApiError({ kind: "unknown", messageKey: "errors.unknown" });
}
