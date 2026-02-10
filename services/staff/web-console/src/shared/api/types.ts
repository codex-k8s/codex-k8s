export type ErrorResponseDto = {
  code: string;
  message: string;
  field?: string;
};

export function isErrorResponseDto(v: unknown): v is ErrorResponseDto {
  if (!v || typeof v !== "object") return false;
  const o = v as Record<string, unknown>;
  return typeof o.code === "string" && typeof o.message === "string";
}

