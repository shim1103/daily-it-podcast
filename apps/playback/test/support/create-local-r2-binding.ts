import type { R2BucketBinding } from "../../worker/src/infrastructure/r2/r2-episode-repository.ts";

/**
 * local R2 binding（getPlatformProxy）の注入ハンドル。
 * 本番 Worker env ではない。列 5 が Adapter に渡す入口。
 *
 * 正: docs/decisions/2026-09-15T12-02-48-feature-generator-r2-write-adapter.md
 */
export type LocalR2BindingHandle = {
  bucket: R2BucketBinding;
  dispose: () => Promise<void>;
};

/**
 * getPlatformProxy 経由で local R2 binding を用意する。
 *
 * @require C で wrangler / proxy を実起動する。本 stub は起動しない。
 * @ensure 本 stub は空操作の bucket と no-op dispose を返す。
 */
export async function createLocalR2Binding(): Promise<LocalR2BindingHandle> {
  const bucket: R2BucketBinding = {
    async get() {
      return null;
    },
    async put() {
      return undefined;
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
