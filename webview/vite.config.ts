import { fileURLToPath } from "node:url";

import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import solid from "vite-plugin-solid";

import packageJson from "../package.json" with { type: "json" };

// The config sits in webview/, but the scripts run from the repository root.
const root = fileURLToPath(new URL(".", import.meta.url));
const { version } = packageJson;

export default defineConfig({
  root,
  define: {
    "import.meta.env.APP_VERSION": JSON.stringify(version),
  },
  plugins: [solid(), tailwindcss()],
  resolve: {
    alias: {
      "~": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    // The app loads this address during `mygo dev` (devUrl in mygo.config.ts).
    // Node may fail to bind the localhost hostname on Windows, so pin IPv4.
    host: "127.0.0.1",
    port: 5173,
    strictPort: true,
    // The development app and the packaged builds are not the page's.
    watch: { ignored: ["**/.mygo/**", "**/build/**"] },
  },
  build: {
    target: "es2022",
  },
});
