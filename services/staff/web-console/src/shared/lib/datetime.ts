function pad2(n: number): string {
  return String(n).padStart(2, "0");
}

function formatOffset(date: Date): string {
  const offMin = date.getTimezoneOffset(); // minutes: UTC - local
  const sign = offMin > 0 ? "-" : "+";
  const abs = Math.abs(offMin);
  const hh = pad2(Math.floor(abs / 60));
  const mm = pad2(abs % 60);
  return `${sign}${hh}:${mm}`;
}

// formatDateTime formats an RFC3339 datetime to a stable, locale-specific string:
// - en: YYYY-MM-DD HH:MM±HH:MM
// - ru: DD.MM.YYYY HH:MM±HH:MM
export function formatDateTime(value: string | null | undefined, locale: string): string {
  if (!value) return "-";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;

  const yyyy = d.getFullYear();
  const MM = pad2(d.getMonth() + 1);
  const dd = pad2(d.getDate());
  const HH = pad2(d.getHours());
  const mm = pad2(d.getMinutes());
  const off = formatOffset(d);

  if (locale === "ru") return `${dd}.${MM}.${yyyy} ${HH}:${mm}${off}`;
  return `${yyyy}-${MM}-${dd} ${HH}:${mm}${off}`;
}

