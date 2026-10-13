import { defineConfig } from "vite";

export default defineConfig(async () => {
  // dynamically import the ESM-only plugin so it is not required synchronously
  const reactPlugin = (await import("@vitejs/plugin-react")).default;

  return {
    plugins: [reactPlugin()],
    server: { port: 5173 },
    // optional: optimizeDeps.exclude if you still see issues
    // optimizeDeps: { exclude: ["@vitejs/plugin-react"] },
  };
});

