import { fileURLToPath } from "node:url";

import { defineConfig } from "vite";
import solid from "@solidjs/vite-plugin";

// The config sits in webview/, but the scripts run from the repository root.
const root = fileURLToPath(new URL(".", import.meta.url));

export default defineConfig({
  root,
  plugins: [solid()],
  server: {
    // The app loads this address during `mygo dev` (devUrl in mygo.json).
    port: 1886,
    strictPort: true,
    // The development app and the packaged builds are not the page's.
    watch: { ignored: ["**/.mygo/**", "**/build/**"] },
  },
  build: {
    target: "es2022",
    outDir: "dist",
    rollupOptions: {
      input: {
        index: fileURLToPath(new URL("./index.html", import.meta.url)),
        snapshot: fileURLToPath(new URL("./snapshot.html", import.meta.url)),
      },
    },
  },
});
