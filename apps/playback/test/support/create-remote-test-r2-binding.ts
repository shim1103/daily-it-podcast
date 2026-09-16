import type { R2BucketBinding } from "../../worker/src/infrastructure/r2/r2-episode-repository.ts";

/**
 * 実 TEST R2 binding（getPlatformProxy の remoteBindings）の注入ハンドル。
 * playback smoke 専用。本番 bucket には触らない（`daily-it-podcast-dev`固定）。
 *
 * 正: docs/decisions/2026-09-16T00-20-08-feature-generator-r2-test-peer-scope.md
 */
export type RemoteTestR2BindingHandle = {
  bucket: R2BucketBinding;
  dispose: () => Promise<void>;
};

/**
 * getPlatformProxy（remoteBindings: true）経由で実 TEST R2 bucket の binding を用意する。
 *
 * @require apps/playback の wrangler 依存、test/support/wrangler.smoke.jsonc の EPISODES binding
 *   （`daily-it-podcast-dev`）、実 Cloudflare 認証（`CLOUDFLARE_API_TOKEN` / account 設定）がある。
 * @ensure env.EPISODES を bucket として返し、dispose で proxy を解放する。
 * @ensure 起動失敗は throw する（黙って Fake に落とさない）。
 */
export async function createRemoteTestR2Binding(): Promise<RemoteTestR2BindingHandle> {
  const { getPlatformProxy } = await import("wrangler");
  const { fileURLToPath } = await import("node:url");
  const path = await import("node:path");

  const configPath = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    "wrangler.smoke.jsonc",
  );

  const proxy = await getPlatformProxy({
    configPath,
    persist: false,
    remoteBindings: true,
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
