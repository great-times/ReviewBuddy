import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    chunkSizeWarningLimit: 1200,
    rollupOptions: {
      output: {
        manualChunks: {
          react: ['react', 'react-dom', 'react-router-dom', 'zustand'],
          antd: ['antd', '@ant-design/icons'],
          markdown: ['react-markdown', 'remark-gfm'],
          axios: ['axios'],
        },
      },
    },
  },
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
