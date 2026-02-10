import axios from "axios";

import { readInitialLocale } from "../../i18n/locale";
import { normalizeApiError } from "./errors";

export const http = axios.create({
  baseURL: "/",
  withCredentials: true,
  timeout: 15000,
});

http.interceptors.request.use((config) => {
  config.headers = config.headers ?? {};
  // Backend may use it later; for now it's required by frontend guidelines.
  config.headers["Accept-Language"] = readInitialLocale();
  return config;
});

http.interceptors.response.use(
  (resp) => resp,
  (err) => Promise.reject(normalizeApiError(err)),
);

