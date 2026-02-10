export type CookieSameSite = "Lax" | "Strict" | "None";

export type CookieOptions = {
  maxAgeDays?: number;
  path?: string;
  secure?: boolean;
  sameSite?: CookieSameSite;
};

function encodeCookieValue(v: string): string {
  return encodeURIComponent(v);
}

function decodeCookieValue(v: string): string {
  try {
    return decodeURIComponent(v);
  } catch {
    return v;
  }
}

export function getCookie(name: string): string | undefined {
  if (typeof document === "undefined") return undefined;
  const needle = `${encodeURIComponent(name)}=`;
  const parts = document.cookie.split(";").map((p) => p.trim());
  for (const p of parts) {
    if (!p.startsWith(needle)) continue;
    return decodeCookieValue(p.slice(needle.length));
  }
  return undefined;
}

export function setCookie(name: string, value: string, opts: CookieOptions = {}): void {
  if (typeof document === "undefined") return;
  const path = opts.path ?? "/";
  const sameSite = opts.sameSite ?? "Lax";
  const secure = opts.secure ?? (typeof location !== "undefined" && location.protocol === "https:");

  const pieces: string[] = [];
  pieces.push(`${encodeURIComponent(name)}=${encodeCookieValue(value)}`);
  pieces.push(`Path=${path}`);
  if (typeof opts.maxAgeDays === "number") {
    pieces.push(`Max-Age=${Math.floor(opts.maxAgeDays * 24 * 60 * 60)}`);
  }
  pieces.push(`SameSite=${sameSite}`);
  if (secure) pieces.push("Secure");

  document.cookie = pieces.join("; ");
}

export function deleteCookie(name: string, opts: Pick<CookieOptions, "path"> = {}): void {
  setCookie(name, "", { ...opts, maxAgeDays: -1 });
}

