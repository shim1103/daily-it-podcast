import { defineConfig } from "vitest/config";

// 進捗 D1 疎通（dispatch 専用）の収集境界。共有の vitest.config.mjs の project へは載せない。
export default defineConfig({
  root: new URL("../..", import.meta.url).pathname,
  test: {
    include: ["test/dispatch/**/*narrow_integration*.test.ts"],
    environment: "node",
    // why: remote binding の proxy 起動が既定の 10 秒を超えうる
    testTimeout: 60_000,
    hookTimeout: 60_000,
  },
});
