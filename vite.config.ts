import { defineConfig } from "vite";
import { join } from "path";
import react from "@vitejs/plugin-react";
import { visualizer } from "rollup-plugin-visualizer";
import banner from "vite-plugin-banner";

function resolve(path: string) {
	return join(__dirname, path);
}

function isDev() {
	const mode = process.env.NODE_ENV || "development";
	return mode === "development";
}

// https://vitejs.dev/config/
export default defineConfig({
	plugins: [
		react(),
		visualizer(),
		banner(`
		@sdttttt: nothing.
	`)
	],
	clearScreen: false,
	build: {
		minify: !isDev(),
		sourcemap: isDev() ? "inline" : false,

		rollupOptions: {
			output: {
				manualChunks(id) {
					if (id.includes("node_modules")) {

						if (id.includes("lodash-es")) {
							return "lodash-es";
						}

						if (id.includes("evergreen-ui")) {
							return "evergreen-ui";
						}

						return "vendor";
					}
				}
			}
		}
	},
	resolve: {
		alias: {
			"@": resolve("src"),
			"@store": resolve("store"),
			"@database": resolve("database")
		}
	}
});
