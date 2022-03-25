import { defineConfig } from 'vite'
import { join } from "path";
import react from '@vitejs/plugin-react'

function resolve(path: string) {
	return join(__dirname, path);
}

// https://vitejs.dev/config/
export default defineConfig({
	plugins: [react()],

	resolve: {
		alias: {
			"@": resolve("src"),
			"redux": resolve("redux"),
		}
	}
})
