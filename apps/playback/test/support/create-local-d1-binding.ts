import type { D1DatabaseBinding } from "../../worker/src/infrastructure/d1/d1-database-binding.ts";
import { EPISODE_PROGRESS_D1_BINDING } from "../../worker/src/infrastructure/d1/progress-d1-constants.ts";

/**
 * local D1 binding（getPlatformProxy）の注入ハンドル。
 * 本番 Worker env ではない。adapter が刺す入口。
 *
 * 正: docs/decisions/2026-09-16T00-20-08-feature-generator-r2-test-peer-scope.md
 */
export type LocalD1BindingHandle = {
  database: D1DatabaseBinding;
  dispose: () => Promise<void>;
};

/**
 * getPlatformProxy 経由で local D1 binding を用意する。
 *
 * @require apps/playback の wrangler 依存と wrangler.jsonc の EPISODE_PROGRESS binding がある。
 * @ensure env.EPISODE_PROGRESS を database として返し、dispose で proxy を解放する。表は未作成の空 D1 である。
 * @ensure 起動失敗は throw する（黙って Fake に落とさない）。binding が無い時は proxy を解放してから throw する。
 */
export async function createLocalD1Binding(): Promise<LocalD1BindingHandle> {
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

  const env = proxy.env as { [EPISODE_PROGRESS_D1_BINDING]?: D1DatabaseBinding };
  const database = env[EPISODE_PROGRESS_D1_BINDING];
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
