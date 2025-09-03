import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@smol-dungeon/schema': '../../../packages/schema/src/index.ts',
      '@smol-dungeon/core': '../../../packages/core/src/index.ts',
      '@smol-dungeon/persistence': '../../../packages/persistence/src/index.browser.ts',
      '@smol-dungeon/adapters': '../../../packages/adapters/src/index.ts',
    },
  },
  define: {
    global: 'globalThis',
  },
});