import { readInitialLocale } from "../../i18n/locale";
import { client } from "./generated/client.gen";
import { normalizeApiError } from "./errors";

client.setConfig({
  baseURL: "/",
  withCredentials: true,
  timeout: 15000,
  throwOnError: true,
});

client.instance.interceptors.request.use((config) => {
  config.headers = config.headers ?? {};
  // Backend may use it later; for now it's required by frontend guidelines.
  config.headers["Accept-Language"] = readInitialLocale();
  return config;
});

client.instance.interceptors.response.use(
  (resp) => resp,
  (err) => Promise.reject(normalizeApiError(err)),
);

export const http = client.instance;
export const apiClient = client;
