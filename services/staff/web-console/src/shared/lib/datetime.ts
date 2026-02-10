import { format, isValid, parseISO } from "date-fns";

// formatDateTime formats an RFC3339 datetime to a stable, locale-specific string:
// - en: YYYY-MM-DD HH:MM±HH:MM
// - ru: DD.MM.YYYY HH:MM±HH:MM
export function formatDateTime(value: string | null | undefined, locale: string): string {
  if (!value) return "-";
  const d = parseISO(value);
  if (!isValid(d)) return value;

  const pattern = locale === "ru" ? "dd.MM.yyyy HH:mmXXX" : "yyyy-MM-dd HH:mmXXX";
  return format(d, pattern);
}
