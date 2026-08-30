import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react()],
  worker: {
    // le worker importe wasm_exec.js et opfs.ts comme des modules ES
    format: "es",
  },
});
