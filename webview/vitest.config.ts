import { defineConfig } from "vitest/config";

// The config sits in webview/, but the scripts run from the repository root.
export default defineConfig({
  root: new URL(".", import.meta.url).pathname,
  test: {
    include: ["src/**/*.test.ts"],
    environment: "node",
  },
});
