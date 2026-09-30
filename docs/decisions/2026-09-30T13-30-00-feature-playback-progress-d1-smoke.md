---
name: 進捗D1の migration file は恒久の置き場に残し、疎通 dispatch 内で migrations apply により TEST D1 へ適用する
date: 2026-09-30T13:30:00
branch: feature/playback-progress-d1-smoke
---

## 1. Decision

1. TEST D1 に進捗表が無いときは、疎通 dispatch の中で `wrangler d1 migrations apply` により、repo の migration file を適用してから疎通する。表の DDL を dispatch 側（yml・script・probe）へ複製しない。
2. migration file は疎通の一過性 artifact に含めず、PASS 後も削除しない。置き場は `apps/playback/worker/migrations/` のままとし、`migrations_dir` を共有 config（`test/support/wrangler.smoke.jsonc`）に宣言する。
3. 疎通の形（binding 面での往復と独立な remote 読み戻し）は、別 Decision（`2026-09-30T12-32-02-feature-playback-progress-d1-smoke`）が持つ。

## 2. Reason

1. 疎通の目的は生 I/O 経路の確認である（R2 疎通の Decision と同じ）。表が無いことは経路の故障ではなく、schema が未適用という環境状態である。表なしで落として手で適用し再 dispatch する運用は、一過性 artifact の目的を二度手間にする。
2. DDL の正本は migration file 1本である。dispatch 側が DDL を持つと第二の SSOT になり、列変更時に片方だけ更新される。
3. `migrations apply` は、適用済みの migration を D1 側の台帳（`d1_migrations`）へ file 名で記録する。`execute --file` で直接流すと台帳に載らず、後で D1 infra が同じ migration を適用したときに表が既にあって衝突する。
4. 適用後に file を消すと、台帳と file が乖離する。本番・新規の環境・local の D1 が、同じ file を必要とする。適用が済んでも、file は schema の履歴として残る。
5. wrangler の既定の置き場は config の隣の `./migrations` で、`worker/migrations` は既定外である。ただし `migrations_dir` の宣言で足りる。dir を移すのは、参照の書き換えだけを増やす。

## 3. Rejected

1. **表を Dashboard 等で手作業適用してから dispatch する案** — 適用内容が repo の migration file と乖離しうり、適用の記録も repo に残らない。
2. **probe が `CREATE TABLE IF NOT EXISTS` を自前で持つ案** — DDL が migration file と probe の2箇所になり、列契約が二重管理になる。
3. **`wrangler d1 execute --file` で migration file を直接流す案** — 適用履歴が台帳に載らず、後続の `migrations apply` と衝突する。
4. **migration file を疎通 dispatch の artifact（`test/dispatch/`）に置き、PASS 後に一緒に削除する案** — 台帳と file が乖離し、本番・新規環境・local の D1 が schema の正本を失う。
5. **file を既定の `./migrations` へ移す案** — `migrations_dir` の宣言で足りるのに、参照の書き換えだけを増やす。
