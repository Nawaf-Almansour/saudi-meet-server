import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    host: '0.0.0.0',
    proxy: {
      '/api/ws': {
        target: 'ws://diwan-api:4000',
        ws: true,
      },
      '/api': {
        target: 'http://diwan-api:4000',
        changeOrigin: true,
      },
    },
  },
});
