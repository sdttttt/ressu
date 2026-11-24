import { defineConfig } from "mygo-cli";

import pkg from "./package.json";

export default defineConfig({
  name: "Ressu",
  slug: "ressu",
  identifier: "com.sdttttt.ressu",
  version: pkg.version,
  copyright: "© 2026 sdttttt",
  bindings: "webview/src/mygo.ts",
  frontendDist: "webview/dist",
  devUrl: "http://localhost:5173",
  devCommand: "bun run dev:web",
  buildCommand: "bun run build:web",
  out: "build",
});
