import type {
  D1DatabaseBinding,
  D1PreparedStatementBinding,
} from "../../worker/src/infrastructure/d1/d1-database-binding.ts";

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

function createZeroDatabase(): D1DatabaseBinding {
  const statement: D1PreparedStatementBinding = {
    bind(..._values: unknown[]) {
      return statement;
    },
    async first() {
      return null;
    },
    async all() {
      return { results: [] };
    },
    async run() {
      return { success: true, meta: { changes: 0 } };
    },
  };
  return {
    prepare(_query: string) {
      return statement;
    },
  };
}

/**
 * local D1 binding 入口（A: signature＋stub）。
 *
 * @require apps/playback の wrangler 依存と wrangler.jsonc の EPISODE_PROGRESS binding がある（C 本実装時）。
 * @ensure A stub は getPlatformProxy を起動せず、読取空・書込 success・changes 0 を返す。
 * @ensure 起動失敗を Fake に黙って落とさないこと（本実装の義務）は C。
 */
// todo: C で getPlatformProxy（remoteBindings: false）本実装へ差し替え。本 stub の zero 戻りをやめる
export async function createLocalD1Binding(): Promise<LocalD1BindingHandle> {
  return {
    database: createZeroDatabase(),
    async dispose() {},
  };
}
