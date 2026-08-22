import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    passWithNoTests: true,
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
            "test/**/*system_e2e*.test.ts",
          ],
          passWithNoTests: true,
        },
      },
    ],
  },
});
