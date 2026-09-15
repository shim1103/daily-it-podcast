import type { LocalR2BindingHandle } from "./create-local-r2-binding.ts";
import type { R2BucketBinding } from "../../worker/src/infrastructure/r2/r2-episode-repository.ts";

/**
 * SU 用 Fake local R2 binding。実 getPlatformProxy / wrangler を起動しない。
 *
 * 正: docs/decisions/2026-09-16T00-20-08-feature-generator-r2-test-peer-scope.md
 *
 * @ensure 空 list / null get の bucket と no-op dispose を返す。
 */
export async function createFakeLocalR2Binding(): Promise<LocalR2BindingHandle> {
  const bucket: R2BucketBinding = {
    async get() {
      return null;
    },
    async list() {
      return { objects: [] };
    },
  };
  return {
    bucket,
    async dispose() {},
  };
}
