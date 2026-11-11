import { defineConfig } from "mygo-cli";

export default defineConfig({
  name: "Ressu",
  slug: "ressu",
  identifier: "com.sdttttt.ressu",
  version: "26.11.8",
  copyright: "© 2026 sdttttt",
  bindings: ".mygo/mygo.ts",
  frontendDist: "webview/static",
  out: "build",
});
