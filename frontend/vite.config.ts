import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import checker from 'vite-plugin-checker'

export default defineConfig({
  plugins: [
    react(),
    checker({
      typescript: true,
      eslint: { lintCommand: 'eslint ."', useFlatConfig: true },
      enableBuild: false,
    }),
  ],
  resolve: { tsconfigPaths: true },
  server: { port: 5173 },
  build: { sourcemap: false },
})
