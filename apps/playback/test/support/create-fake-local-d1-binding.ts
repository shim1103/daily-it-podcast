import type { LocalD1BindingHandle } from "./create-local-d1-binding.ts";
import type { D1PreparedStatementBinding } from "../../worker/src/infrastructure/d1/d1-database-binding.ts";

/**
 * SU 用 Fake local D1 binding。実 getPlatformProxy / wrangler を起動しない。
 *
 * 正: docs/decisions/2026-09-16T00-20-08-feature-generator-r2-test-peer-scope.md
 *
 * @ensure 読取は常に空（first は null・all は results 空）、書込は success・changes 0 の database と、no-op dispose を返す。
 * @ensure bind は statement 自身を返す。
 */
export async function createFakeLocalD1Binding(): Promise<LocalD1BindingHandle> {
  const statement: D1PreparedStatementBinding = {
    bind() {
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
    database: {
      prepare() {
        return statement;
      },
    },
    async dispose() {},
  };
}
