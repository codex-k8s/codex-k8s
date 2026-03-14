import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import vuetify from "vite-plugin-vuetify";

const disableHmr = ["1", "true", "yes"].includes(String(process.env.VITE_DISABLE_HMR || "").toLowerCase());
const enablePollingWatch = ["1", "true", "yes"].includes(String(process.env.VITE_WATCH_USE_POLLING || "").toLowerCase());

function parseOptionalInt(value: string | undefined): number | undefined {
  if (!value) {
    return undefined;
  }
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
}

export default defineConfig({
  plugins: [vue(), vuetify({ autoImport: true })],
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    strictPort: true,
    watch: (() => {
      const interval = parseOptionalInt(process.env.VITE_WATCH_INTERVAL);
      if (!enablePollingWatch && interval === undefined) {
        return undefined;
      }

      return {
        usePolling: enablePollingWatch,
        interval,
      };
    })(),
    allowedHosts: process.env.VITE_ALLOWED_HOSTS ? [process.env.VITE_ALLOWED_HOSTS] : true,
    hmr: (() => {
      if (disableHmr) {
        return false;
      }

      const host = process.env.VITE_HMR_HOST;
      const protocol = process.env.VITE_HMR_PROTOCOL;
      const clientPort = parseOptionalInt(process.env.VITE_HMR_CLIENT_PORT);
      const port = parseOptionalInt(process.env.VITE_HMR_PORT);
      const path = process.env.VITE_HMR_PATH;

      // Keep local `npm run dev` default behavior unless explicitly configured.
      const configured = Boolean(host || protocol || clientPort || port || path);
      if (!configured) {
        return undefined;
      }

      return {
        host,
        protocol,
        clientPort,
        port,
        path,
      };
    })(),
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8080",
        // Keep the original dev host so api-gateway origin validation still matches websocket upgrades.
        ws: true,
      },
      "/metrics": "http://127.0.0.1:8080",
      "/health": "http://127.0.0.1:8080",
    },
  },
});
