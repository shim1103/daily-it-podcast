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
            "test/**/*narrow_integration*.test.ts",
          ],
          passWithNoTests: true,
        },
      },
      {
        test: {
          name: "integration",
          include: [
            "test/**/*narrow_integration*.test.ts",
            "test/**/*broad_integration*.test.ts",
            "test/**/*contract*.test.ts",
          ],
          passWithNoTests: true,
        },
      },
    ],
  },
});
