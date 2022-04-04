import { defineConfig } from 'vite'
import { join } from "path";
import react from '@vitejs/plugin-react'
import { visualizer } from "rollup-plugin-visualizer";

function resolve(path: string) {
	return join(__dirname, path);
}

const mode = process.env.NODE_ENV || 'development'

// https://vitejs.dev/config/
export default defineConfig({
	plugins: [react(), visualizer()],
	clearScreen: false,
	build: {
		minify: mode === 'development' ? false : "terser",
		sourcemap: mode === 'development' ? "inline" : false,
	},
	resolve: {
		alias: {
			"@": resolve("src"),
			"@store": resolve("store"),
		}
	}
})
