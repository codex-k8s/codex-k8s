function normalize(value: string | null | undefined): string {
  return String(value || "")
    .trim()
    .toLowerCase();
}

export function colorForRunStatus(status: string | null | undefined): string {
  const s = normalize(status);
  switch (s) {
    case "succeeded":
    case "success":
      return "success";
    case "failed":
    case "error":
      return "error";
    case "running":
      return "info";
    case "pending":
    case "waiting":
      return "warning";
    case "canceled":
    case "cancelled":
      return "secondary";
    default:
      return "secondary";
  }
}

export function colorForProjectRole(role: string | null | undefined): string {
  const r = normalize(role);
  switch (r) {
    case "admin":
      return "warning";
    case "read_write":
      return "info";
    case "read":
      return "secondary";
    default:
      return "secondary";
  }
}

export function colorForYesNo(value: boolean): string {
  return value ? "success" : "secondary";
}

