import { coverageConfigDefaults, defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    passWithNoTests: true,
    coverage: {
      provider: "v8",
      include: ["contracts/**/*.ts", "web/src/**/*.ts", "web/src/**/*.tsx", "worker/src/**/*.ts"],
      exclude: [
        ...coverageConfigDefaults.exclude,
        "web/main.ts",
        "web/src/api/api-result.ts",
        "worker/src/application/ports/**",
        // why: 実行文を持たない型宣言のみ file は v8 が分母 0 の 0% 行として表示し、
        //   coverage 表の見た目を崩す。型の回帰は typecheck が担う（Decision 2026-09-04T18-30-01）
        "worker/src/composition/runtime-config-bindings.ts",
      ],
      thresholds: {
        branches: 100,
        "worker/src/routes/**": { branches: 90 },
        "worker/src/composition/**": { branches: 90 },
        "worker/src/infrastructure/drive/google-drive-episode-repository.ts": {
          branches: 90,
        },
        "worker/src/controllers/map-internal-error.ts": { branches: 90 },
        "web/src/pages/**": { branches: 90 },
        "web/src/view-models/**": { branches: 90 },
      },
    },
    projects: [
      {
        test: {
          name: "unit",
          environment: "happy-dom",
          include: [
            "contracts/**/*sociable_unit*.test.ts",
            "web/vite-config.sociable_unit.test.ts",
            "web/src/**/*sociable_unit*.test.ts",
            "worker/src/**/*sociable_unit*.test.ts",
            // why: secret なし NI を Unit coverage 分母へ算入する（Decision 2026-08-30T16-20-01）
            "test/integration/**/*narrow_integration*.test.ts",
            // why: frontend Broad は真の外部を Stub せず経路がそのまま製品ロジックの実行になるため
            //   Unit coverage 分母へ算入する（Decision 2026-09-04T18-30-00 §5）。worker Broad は
            //   Drive Stub を挟むため分母から外す方針を維持し、frontend の合成入口 file だけを足す
            "test/integration/episode_list_page.broad_integration.test.ts",
          ],
          passWithNoTests: true,
        },
      },
      {
        test: {
          name: "integration",
          // why: frontend Broad（episode_list_page）は React を render するため happy-dom が要る。
          //   node 前提の worker Broad / Narrow は各 file 冒頭の `@vitest-environment node` で戻す
          environment: "happy-dom",
          include: [
            "test/integration/**/*narrow_integration*.test.ts",
            "test/integration/**/*broad_integration*.test.ts",
          ],
          passWithNoTests: true,
        },
      },
    ],
  },
});
