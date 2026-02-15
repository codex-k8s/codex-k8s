import "vuetify/styles";
import "@mdi/font/css/materialdesignicons.css";

import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { aliases, mdi } from "vuetify/iconsets/mdi";

export const vuetify = createVuetify({
  components,
  directives,
  icons: {
    defaultSet: "mdi",
    aliases,
    sets: { mdi },
  },
  theme: {
    defaultTheme: "codexLight",
    themes: {
      codexLight: {
        dark: false,
        colors: {
          primary: "#0b5a83",
          secondary: "#111827",
          success: "#0e7c3a",
          warning: "#b45309",
          error: "#b42318",
          info: "#1d4ed8",
        },
      },
    },
  },
});

