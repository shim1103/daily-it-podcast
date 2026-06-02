import { defineConfig } from 'vitest/config';
import { resolve } from 'path';

export default defineConfig({
  test: {
    // e2e ディレクトリは Playwright で実行するため Vitest から除外
    exclude: ['e2e/**', '**/node_modules/**'],
    environment: 'node',
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
    },
  },
});
