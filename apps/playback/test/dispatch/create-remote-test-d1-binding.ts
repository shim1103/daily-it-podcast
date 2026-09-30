import { getPlatformProxy } from "wrangler";
import type { D1DatabaseBinding } from "../../worker/src/infrastructure/d1/d1-database-binding.ts";
import { EPISODE_PROGRESS_D1_BINDING } from "../../worker/src/infrastructure/d1/progress-d1-constants.ts";
import { smokeConfigPath } from "./smoke-config.ts";

/**
 * 実 TEST D1 binding の注入ハンドル。dispatch 疎通専用。
 *
 * 正: docs/decisions/2026-09-30T12-32-02-feature-playback-progress-d1-smoke.md
 */
export type RemoteTestD1BindingHandle = {
  database: D1DatabaseBinding;
  dispose: () => Promise<void>;
};

type RemoteTestEnv = Record<string, D1DatabaseBinding | undefined>;

/**
 * getPlatformProxy（remoteBindings: true）経由で TEST D1 の binding を用意する。
 *
 * @require test/support/wrangler.smoke.jsonc に EPISODE_PROGRESS binding（TEST D1）がある。
 * @require 実 Cloudflare 認証（`CLOUDFLARE_API_TOKEN` / `CLOUDFLARE_ACCOUNT_ID`）が process env にある。
 * @ensure env.EPISODE_PROGRESS を database として返し、dispose で proxy を解放する。
 * @invariant TEST 固定。本番 D1 には触れず、切替引数を持たない。
 */
export async function createRemoteTestD1Binding(): Promise<RemoteTestD1BindingHandle> {
  const proxy = await getPlatformProxy({
    configPath: smokeConfigPath,
    persist: false,
    remoteBindings: true,
  });

  // why: proxy.env は Record<string, unknown>。adapter が呼ぶ最小面 D1DatabaseBinding へ絞る
  const database = (proxy.env as RemoteTestEnv)[EPISODE_PROGRESS_D1_BINDING];
  if (database === undefined) {
    await proxy.dispose();
    throw new Error(`${EPISODE_PROGRESS_D1_BINDING} binding が無い`);
  }

  return {
    database,
    dispose: async () => {
      await proxy.dispose();
    },
  };
}
