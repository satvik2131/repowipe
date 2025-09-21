import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";
import { componentTagger } from "lovable-tagger";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
  server: {
    host: "::",
    port: 3000,
  },
  plugins: [react(), mode === "development" && componentTagger()].filter(
    Boolean
  ),
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  // Add base path for assets when building
  base: mode === "production" ? "/" : "/",
  build: {
    outDir: "../backend/static", // Build directly into the backend static folder
    emptyOutDir: true,
    assetsDir: "assets", // This ensures assets go into /assets/ folder
  },
}));
