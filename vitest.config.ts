/// <reference types="vitest" />

import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    // ...
    includeSource: ['src/**/*.{js,ts}', 'store/**/*.{js,ts}', "database/*.{js,ts}"],
    environment: "happy-dom"
  },
})