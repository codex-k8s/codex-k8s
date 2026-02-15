import { createApp } from "vue";
import { createPinia } from "pinia";
import VueCookies, { globalCookiesConfig } from "vue3-cookies";

import App from "./app/App.vue";
import { vuetify } from "./app/plugins/vuetify";
import { i18n } from "./i18n";
import { createAppRouter } from "./router";

import "./app/styles/global.css";

globalCookiesConfig({
  expireTimes: "365d",
  path: "/",
  domain: "",
  secure: false,
  sameSite: "Lax",
});

const app = createApp(App);
const pinia = createPinia();
app.use(pinia);
app.use(vuetify);
app.use(i18n);
app.use(VueCookies);

const router = createAppRouter(pinia);
app.use(router);

app.mount("#app");
