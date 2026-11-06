import { defineConfig } from "mygo-cli";

export default defineConfig({
  name: "Ressu",
  slug: "ressu",
  identifier: "com.sdttttt.ressu",
  version: "0.1.0",
  copyright: "© 2026 sdttttt",
  devUrl: "http://localhost:5173",
  devCommand: "bun run dev:web",
  buildCommand: "bun run build:web",
  frontendDist: "webview/dist",
  bindings: "webview/src/mygo.ts",
  out: "build",
});
