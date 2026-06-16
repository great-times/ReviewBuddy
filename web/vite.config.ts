import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 26406,
    strictPort: true,
    proxy: {
      '/api': {
        target: 'http://localhost:26405',
        changeOrigin: true,
      },
    },
  },
});
