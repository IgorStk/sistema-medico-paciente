import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    host: "localhost",
    port: 5173, // porta liberada no CORS dos serviços
    strictPort: true, // falha em vez de trocar de porta (outra porta seria bloqueada pelo CORS)
  },
});
