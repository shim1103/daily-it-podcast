import type { R2BucketBinding } from "../../worker/src/infrastructure/r2/r2-episode-repository.ts";

/**
 * local R2 binding（getPlatformProxy）の注入ハンドル。
 * 本番 Worker env ではない。列 5 が Adapter に渡す入口。
 *
 * 正: docs/decisions/2026-09-16T00-20-08-feature-generator-r2-test-peer-scope.md
 */
export type LocalR2BindingHandle = {
  bucket: R2BucketBinding;
  dispose: () => Promise<void>;
};

/**
 * getPlatformProxy 経由で local R2 binding を用意する。
 *
 * @require apps/playback の wrangler 依存と wrangler.jsonc の EPISODES binding がある。
 * @ensure env.EPISODES を bucket として返し、dispose で proxy を解放する。
 * @ensure 起動失敗は throw する（黙って Fake に落とさない）。
 */
export async function createLocalR2Binding(): Promise<LocalR2BindingHandle> {
  const { getPlatformProxy } = await import("wrangler");
  const { fileURLToPath } = await import("node:url");
  const path = await import("node:path");

  const configPath = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    "../../wrangler.jsonc",
  );

  const proxy = await getPlatformProxy({
    configPath,
    persist: false,
    remoteBindings: false,
  });

  const env = proxy.env as { EPISODES?: R2BucketBinding };
  const bucket = env.EPISODES;
  if (bucket === undefined) {
    await proxy.dispose();
    throw new Error("EPISODES binding が無い");
  }

  return {
    bucket,
    dispose: async () => {
      await proxy.dispose();
    },
  };
}
