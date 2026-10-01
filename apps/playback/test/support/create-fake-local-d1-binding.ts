import type { LocalD1BindingHandle } from "./create-local-d1-binding.ts";
import type {
  D1DatabaseBinding,
  D1PreparedStatementBinding,
  D1Row,
} from "../../worker/src/infrastructure/d1/d1-database-binding.ts";

/** prepare から bind までで記録した 1 回分の呼び出し。 */
export type D1Call = { sql: string; values: unknown[] };

type D1RunResult = Awaited<ReturnType<D1PreparedStatementBinding["run"]>>;

/** 読取・書込ごとの応答の差し込み口。省略した操作は既定（空応答）に倒れる。throw / reject で失敗も差し込める。 */
export type FakeLocalD1Responses = {
  first?: (call: D1Call) => Promise<D1Row | null>;
  all?: (call: D1Call) => Promise<D1Row[]>;
  run?: (call: D1Call) => Promise<D1RunResult>;
};

/** database と dispose に加え、prepare に渡った呼び出しを記録して返す handle。 */
export type FakeLocalD1BindingHandle = LocalD1BindingHandle & { calls: D1Call[] };

/**
 * SU 用 Fake local D1 binding。実 getPlatformProxy / wrangler を起動しない。
 * D1 binding の SU 用 double はこの 1 つに集約し、test file ごとに手書きしない。
 *
 * 正: docs/decisions/2026-09-16T00-20-08-feature-generator-r2-test-peer-scope.md
 *
 * @ensure responses を渡さない時、読取は常に空（first は null・all は results 空）、書込は success・changes 0 を返す。no-op dispose を返す。
 * @ensure responses で差し込んだ操作は、その戻り値または失敗を返す。差し込まない操作は空応答のまま。
 * @ensure calls に prepare ごとの SQL と、bind された値（bind しなければ空配列）を呼び出し順で記録する。
 * @ensure bind は statement 自身を返す。
 */
export async function createFakeLocalD1Binding(
  responses: FakeLocalD1Responses = {},
): Promise<FakeLocalD1BindingHandle> {
  const calls: D1Call[] = [];
  const database: D1DatabaseBinding = {
    prepare(sql) {
      const call: D1Call = { sql, values: [] };
      calls.push(call);
      const statement: D1PreparedStatementBinding = {
        bind(...values) {
          call.values = values;
          return statement;
        },
        async first<T extends D1Row = D1Row>() {
          return ((await responses.first?.(call)) ?? null) as T | null;
        },
        async all<T extends D1Row = D1Row>() {
          return { results: ((await responses.all?.(call)) ?? []) as T[] };
        },
        async run() {
          return (await responses.run?.(call)) ?? { success: true, meta: { changes: 0 } };
        },
      };
      return statement;
    },
  };
  return { database, calls, async dispose() {} };
}
