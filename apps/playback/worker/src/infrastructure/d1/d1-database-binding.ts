/**
 * Workers D1 binding の読み書き最小面。
 * wrangler 生成の `D1Database` / `D1PreparedStatement` の部分集合に合わせる設計意図であり、型検査では担保していない。
 * `worker-configuration.d.ts` は `--include-runtime false` 生成で `D1Database` が型 scope に無く、
 * 実 `D1Database` から本型への構造的代入はどこでも検査されない。実 binding との整合は local D1 を刺す NI の実行で確かめる。
 * Port（ProgressRepository）は本型を露出しない。D1 adapter だけが使う。
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
