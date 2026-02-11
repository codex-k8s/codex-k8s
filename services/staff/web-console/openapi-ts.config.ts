import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../../external/api-gateway/api/server/api.yaml",
  output: "src/shared/api/generated",
  plugins: ["@hey-api/client-axios"],
});
