---
name: D1のschema正本はDrizzleのTS schemaに寄せ、ORMはDrizzle（RC版をexact固定）とし、migration SQLは生成された履歴として残す
date: 2026-10-01T23:33:20
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. D1 の schema の正本を TS（Drizzle の schema）にする。ORM は Drizzle とする。
2. `drizzle-orm` と `drizzle-kit` は、RC を exact で固定する。範囲指定（`^` など）で自動追従せず、追従は意図して行う。実際の版は `apps/playback/package.json` が正本である（確認時点は `1.0.0-rc.4`）。
3. migration は `drizzle-kit generate` が生成する履歴（`<時刻>_<名前>/migration.sql` と `snapshot.json` の組）として commit する。適用は `wrangler d1 migrations apply` で行い、この layout を読ませるには `migrations_pattern` の指定が要る。D1 は適用済みの migration を台帳 `d1_migrations` に記録する（`2026-09-30T13-30-00-feature-playback-progress-d1-smoke.md`）。`migrations_dir` と `migrations_pattern` は、本番の `wrangler.jsonc` と smoke 用 config の D1 entry に宣言する。
4. test が migration を流す helper は残し、文の分割を `;` から、生成 SQL の区切り `--> statement-breakpoint` に置き換える。migration は名前の昇順に適用する。
5. 既存の手書き `0001_episode_progress.sql` を baseline として引き継ぐ扱いは不要とする。`drizzle-kit generate` が出す初期 migration を正本にして作り直す（手書きの file は削除した）。実 instance へ未適用の間の migration の扱いは `2026-10-02T15-44-40-feature-playback-progress-d1-local-peer.md` が持つ。
6. 進捗表の DDL の正本は `apps/playback/worker/src/infrastructure/d1/schema.ts`（Drizzle schema）である。表名・列名の定数（`progress-d1-constants.ts`）はそこから導き、literal を複製しない。migration の生成は `apps/playback/package.json` の `db:generate` script が持つ。
7. 段階の分け方は、A（schema・定数の導出・生成 migration・config 結線・test の適用 helper）と C（adapter の query の Drizzle への書き換え）とする。手書きの `D1DatabaseBinding` 型は C まで残す。`@cloudflare/workers-types` は入れない（Rejected 6）。
8. 未確認・未判断（実施 Issue で確かめる）:
   1. 本番・TEST の remote D1 への適用。ここまでの確認は local D1（miniflare）で行った。
   2. 複数の migration を並べた時の昇順の適用と、台帳での名前の一意性。台帳には `<dir 名>/migration.sql` で記録されることは、1 つの migration で確認した。
   3. drizzle-kit が RC の間の layout の安定性。
   4. Worker bundle の増加を許容するか（未判断）。schema の import 後に `wrangler deploy --dry-run` で測ると、非圧縮で約 +47 KiB、gzip で約 +10 KiB だった。許容の判断は Decision にしていない。

## 2. Reason

1. 表名・列名が、migration SQL・`progress-d1-constants.ts`・adapter の SQL の3箇所にある。同じ知識の重複で、列を変えるたびに3箇所を揃える必要がある（`design-philosophy.md` §2-2 DRY）。
2. 生 SQL の組立には型が無い。typo や列名のずれは実行時まで分からない。
3. 手書きの query 組立や、3箇所を突き合わせる guard を保守し続けるのは、既存の道具が持つ機能の再発明になる。
4. `design-philosophy.md` §4-2 は標準形式を優先する。migration を生成された SQL として残せば、標準形式（SQL）を捨てないので、この原則に反しない。正本が SQL から TS へ移るだけで、適用する対象は SQL のままである。
5. Drizzle は、TS が正本になる。D1 binding を `drizzle(env.DB)` の形で直接受けられ、build に足す生成 step が無い。
6. 使い捨ての spike（local D1、repo には残していない）で次を確認した。
   1. 条件付き更新（`update().set().where()` の `meta.changes` が一致 1・不一致 0）、`insert … onConflictDoNothing()`（新規 1・競合 0）、`batch()` が local D1 で成立した。
   2. `.returning()` は update と `onConflictDoUpdate` のどちらとも併用できた。Reason の根拠だった「SQLite / D1 で併用できるか」の未確認は、local では解けた。
   3. wrangler は既定ではこの layout を読めず、`migrations_pattern` を指定すると適用でき、台帳にも記録された。wrangler 自身が警告でこの指定を案内する。
   4. `migrations_pattern` の相対 path は、config file のある dir が基準（cwd ではない）で、`migrations_dir/` で始まる必要がある。本番 config は `apps/playback` 基準、smoke 用 config は自身の位置（`test/support/`）基準で書く。台帳には `<dir 名>/migration.sql` で記録される。
7. `@cloudflare/workers-types` を入れない理由は、型の衝突である。入れて `tsconfig.json` の `types` に加えると、web の `playback-rpc-client.ts` で typecheck が 5 件落ちた。workers-types が DOM の `Response` を上書きし、Hono の `ClientResponse` が代入できなくなる。手書きの `D1DatabaseBinding` が C まで残るのは、この不採用の結果である。Drizzle の query の書き換えも、adapter の書き換えと一緒に C に置く。
8. 版を exact で固定するのは、Drizzle の D1 docs が `@rc` の install を案内していて RC を使う前提になり、かつ RC の間は API と migration の layout が変わりうるからである。固定すれば変更は意図した更新の時だけに起き、追従の確認（上の spike と同じ確認）をその時に行える。

## 3. Rejected

1. **生 SQL を継続する案** — 上の Reason 1〜3 が残る。3箇所の重複、型の無い query、手書きの guard の保守。
2. **Prisma を使う案** — 取得時点の Cloudflare docs では、D1 の driver adapter が Preview、Prisma Migrate for D1 が Early Access だった。専用の DSL を持ち、`prisma generate` が要る。専用 DSL を増やす点は §4-2 とも合わない。この確認は本記録の作成時に再確認していない。
3. **`.sql` から型付き TS を生成する案（sqlc 系・TypeSQL）** — community の plugin に依存し、保守状況が未確認である。TS が正本にならない。
4. **stable の 0.x を使う案** — docs の案内が RC であり、migration の layout も RC と異なる可能性がある。ただし 0.x の layout は確認しておらず、異なると断定はできない。
5. **範囲指定（`^`）で RC に自動追従する案** — RC の間は API と layout が変わりうる。install のたびに挙動が動き、変更の原因が追えなくなる。
6. **`@cloudflare/workers-types` を入れ、手書きの binding 型を置き換える案** — 上の Reason 7 のとおり、web の typecheck が 5 件落ちる（DOM の `Response` の上書きによる Hono の `ClientResponse` の代入不可）。衝突を避けて入れ直す方法は、確かめていない。
