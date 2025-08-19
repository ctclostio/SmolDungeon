import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    include: ['tests/**/*.test.ts', 'packages/*/src/**/*.test.ts'],
    exclude: ['node_modules', 'dist'],
  },
  resolve: {
    alias: {
      '@smol-dungeon/schema': './packages/schema/src/index.ts',
      '@smol-dungeon/core': './packages/core/src/index.ts',
      '@smol-dungeon/persistence': './packages/persistence/src/index.ts',
      '@smol-dungeon/adapters': './packages/adapters/src/index.ts',
    },
  },
});