import { format, isValid, parseISO } from "date-fns";

// formatDateTime formats an RFC3339 datetime to a stable, locale-specific string:
// - en: YYYY-MM-DD HH:MM
// - ru: DD.MM.YYYY HH:MM
export function formatDateTime(value: string | null | undefined, locale: string): string {
  if (!value) return "-";
  const d = parseISO(value);
  if (!isValid(d)) return value;

  const pattern = locale === "ru" ? "dd.MM.yyyy HH:mm" : "yyyy-MM-dd HH:mm";
  return format(d, pattern);
}

// formatDurationSince returns compact SLA-like elapsed time from value until now.
export function formatDurationSince(value: string | null | undefined, locale: string): string {
  if (!value) return "-";
  const startedAt = parseISO(value);
  if (!isValid(startedAt)) return "-";

  const diffMs = Date.now() - startedAt.getTime();
  if (diffMs <= 0) {
    return locale === "ru" ? "<1м" : "<1m";
  }

  const totalMinutes = Math.floor(diffMs / 60000);
  const days = Math.floor(totalMinutes / (24 * 60));
  const hours = Math.floor((totalMinutes % (24 * 60)) / 60);
  const minutes = totalMinutes % 60;

  if (locale === "ru") {
    if (days > 0) return `${days}д ${hours}ч`;
    if (hours > 0) return `${hours}ч ${minutes}м`;
    return `${minutes}м`;
  }

  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}
