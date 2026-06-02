// @ts-check
import tseslint from 'typescript-eslint';

/**
 * 全パッケージ共通のベースルール
 * flat config 形式（ESLint v9+）
 */
const baseConfig = tseslint.config(
  // 共通対象ファイル
  {
    files: [
      'packages/*/src/**/*.ts',
      'apps/*/src/**/*.{ts,tsx}',
      'config/**/*.ts',
    ],
    extends: [tseslint.configs.recommended],
    rules: {
      // console.log は warn（orchestrator エントリポイントは個別で off）
      'no-console': 'warn',
      // unused-vars は @typescript-eslint 側に一本化
      'no-unused-vars': 'off',
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
      // 明示的な any は禁止
      '@typescript-eslint/no-explicit-any': 'error',
    },
  },
  // orchestrator エントリポイント: console は業務ログとして許可
  {
    files: [
      'packages/orchestrator/src/main.ts',
      'packages/orchestrator/src/orchestrator.ts',
    ],
    rules: {
      'no-console': 'off',
    },
  },
  // テストファイル: no-explicit-any を warn に緩和
  {
    files: [
      'packages/*/src/__tests__/**/*.ts',
      'apps/*/src/__tests__/**/*.ts',
    ],
    rules: {
      '@typescript-eslint/no-explicit-any': 'warn',
    },
  },
);

export default baseConfig;
