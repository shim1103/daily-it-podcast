/**
 * Workers D1 binding の読み書き最小面。
 * 本番型は wrangler 生成の `D1Database` / `D1PreparedStatement` に合わせる。
 * Port（ProgressRepository）は本型を露出しない。Adapter（C）だけが使う。
 */

/** `first` / `all` / `run` が返す行の受け皿。列名は SQL 側の alias に従う。 */
export type D1Row = Record<string, unknown>;

export type D1PreparedStatementBinding = {
  bind(...values: unknown[]): D1PreparedStatementBinding;
  first<T extends D1Row = D1Row>(): Promise<T | null>;
  all<T extends D1Row = D1Row>(): Promise<{ results: T[] }>;
  run(): Promise<{ success: boolean; meta: { changes: number } }>;
};

/**
 * D1 database binding の最小面（progress adapter が呼ぶ範囲）。
 *
 * @invariant vendor 固有の batch / session 等は、adapter が必要になったら面を広げる（YAGNI）
 */
export type D1DatabaseBinding = {
  prepare(query: string): D1PreparedStatementBinding;
};

/**
 * A 足場用 Stub。読取は空、書込は success・changes 0。
 */
// todo: D1 adapter（C）が入ったら本番結線へ差し替え、本 Stub の呼び出し元を adapter test の Fake に寄せる
export class StubD1Database implements D1DatabaseBinding {
  prepare(_query: string): D1PreparedStatementBinding {
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
    return statement;
  }
}
