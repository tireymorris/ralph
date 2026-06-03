import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "node",
    environmentMatchGlobs: [["**/*.test.tsx", "jsdom"]],
    setupFiles: ["./src/test/setup.ts"],
  },
  build: {
    outDir: "../internal/web/static/dist",
    emptyOutDir: true,
  },
});
