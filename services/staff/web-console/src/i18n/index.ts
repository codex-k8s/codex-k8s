import { createI18n } from "vue-i18n";

import { readInitialLocale } from "./locale";
import { en } from "./messages/en";
import { ru } from "./messages/ru";

export const i18n = createI18n({
  legacy: false,
  locale: readInitialLocale(),
  fallbackLocale: "en",
  globalInjection: true,
  messages: {
    en,
    ru,
  },
});

