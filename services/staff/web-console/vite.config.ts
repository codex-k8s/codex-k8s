import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

const disableHmr = ["1", "true", "yes"].includes(String(process.env.VITE_DISABLE_HMR || "").toLowerCase());

function parseOptionalInt(value: string | undefined): number | undefined {
  if (!value) {
    return undefined;
  }
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
}

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    strictPort: true,
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
      "/api": "http://127.0.0.1:8080",
      "/metrics": "http://127.0.0.1:8080",
      "/health": "http://127.0.0.1:8080",
    },
  },
});
