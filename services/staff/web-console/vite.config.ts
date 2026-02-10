import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

const disableHmr = ["1", "true", "yes"].includes(String(process.env.VITE_DISABLE_HMR || "").toLowerCase());

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
    hmr: disableHmr ? false : undefined,
    proxy: {
      "/api": "http://127.0.0.1:8080",
      "/metrics": "http://127.0.0.1:8080",
      "/health": "http://127.0.0.1:8080",
    },
  },
});
