import { getCookie, setCookie } from "../shared/lib/cookies";

export const supportedLocales = ["en", "ru"] as const;
export type Locale = (typeof supportedLocales)[number];

const cookieKey = "codexk8s_locale";

export function readInitialLocale(): Locale {
  const v = (getCookie(cookieKey) || "").toLowerCase();
  if (v === "ru") return "ru";
  return "en";
}

export function persistLocale(locale: Locale): void {
  setCookie(cookieKey, locale, { maxAgeDays: 365, sameSite: "Lax" });
}

