import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    passWithNoTests: true,
    projects: [
      {
        test: {
          name: "unit",
          include: [
            "contracts/**/*.test.ts",
            "web/src/**/*.test.ts",
            "worker/src/**/*.test.ts",
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
