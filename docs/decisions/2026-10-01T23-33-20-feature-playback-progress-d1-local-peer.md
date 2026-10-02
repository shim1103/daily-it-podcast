---
name: D1のschema正本はDrizzleのTS schemaに寄せ、ORMはDrizzleとし、migration SQLは生成された履歴として残す
date: 2026-10-01T23:33:20
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. D1 の schema の正本を TS（Drizzle の schema）にする。ORM は Drizzle とする。
2. migration SQL は `drizzle-kit generate` が生成する履歴として commit する。適用は `wrangler d1 migrations apply` で行う。D1 は適用済みの SQL を台帳 `d1_migrations` に file 名で記録する（`2026-09-30T13-30-00-feature-playback-progress-d1-smoke.md`）。
3. 既存の `0001_episode_progress.sql` を baseline として引き継ぐ扱いは不要とする。`drizzle-kit generate` が出す初期 migration を正本にして作り直す。実 instance へ未適用の間の migration の扱いは `2026-10-02T15-44-40-feature-playback-progress-d1-local-peer.md` が持つ。
4. 未確認（実施 Issue で確かめる）:
   1. Drizzle の版は 0.x または 1.0 rc の可能性があり、版の固定が要る（二次情報）。
   2. SQLite / D1 で `onConflictDoUpdate` と `.returning()` を併用できるか。

## 2. Reason

1. 表名・列名が、migration SQL・`progress-d1-constants.ts`・adapter の SQL の3箇所にある。同じ知識の重複で、列を変えるたびに3箇所を揃える必要がある（`design-philosophy.md` §2-2 DRY）。
2. 生 SQL の組立には型が無い。typo や列名のずれは実行時まで分からない。
3. 手書きの query 組立や、3箇所を突き合わせる guard を保守し続けるのは、既存の道具が持つ機能の再発明になる。
4. `design-philosophy.md` §4-2 は標準形式を優先する。migration を生成された SQL として残せば、標準形式（SQL）を捨てないので、この原則に反しない。正本が SQL から TS へ移るだけで、適用する対象は SQL のままである。
5. Drizzle は、TS が正本になる。D1 binding を `drizzle(env.DB)` の形で直接受けられ、build に足す生成 step が無い。

## 3. Rejected

1. **生 SQL を継続する案** — 上の Reason 1〜3 が残る。3箇所の重複、型の無い query、手書きの guard の保守。
2. **Prisma を使う案** — 取得時点の Cloudflare docs では、D1 の driver adapter が Preview、Prisma Migrate for D1 が Early Access だった。専用の DSL を持ち、`prisma generate` が要る。専用 DSL を増やす点は §4-2 とも合わない。この確認は本記録の作成時に再確認していない。
3. **`.sql` から型付き TS を生成する案（sqlc 系・TypeSQL）** — community の plugin に依存し、保守状況が未確認である。TS が正本にならない。
