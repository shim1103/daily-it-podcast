---
name: PR #215 の merge 後に、依存の脆弱性を解消し、Drizzle の A（schema・生成 migration・config 結線）を実装し、Decision の矛盾を解消した
date: 2026-10-02T19:25:54
session_id: d4d2716d-c0b8-40f5-be2c-27fbe4fe1235
branch: feature/playback-progress
prev: 2026-10-02T16-57-42-feature-playback-progress-d1-local-peer.md
---

## 1. Summary

PR #215（進捗 D1 local peer）が epic へ merge された後の区切り。epic branch へ戻り、並行制御・cursor・skew・版つき永続・Drizzle の実装 Issue（C）を起票し、`npm audit` の指摘を 0 件にし、Drizzle の A（境界）を `schema.ts` まで広げ、進捗まわりの Decision の矛盾を解消した。実装 Issue は #217（skew）・#218（版つき永続）・#219（CAS）・#220（cursor）と、Drizzle の #222（binding の型付けの方針）・#223（adapter の query の書き換え）。

## 2. Changes

1. `npm audit` は、Drizzle の追加前（HEAD の lockfile）でも 10 件（moderate 4・high 6）あった。Drizzle 由来ではない。`npm audit fix`（`--force` なし）で 3 件（vitest 系の moderate）に減り、vitest 4.1.11 への更新で 0 件になった。
2. vitest 4 への更新は、設定の直しだけでは緑にならなかった。`-- @preserve` なしの `v8 ignore` hint では coverage の branches が 99.29%（280/282）になり、付けると 100%（276/276）になった。hint を持つ製品 source は 5 箇所（コメントのみ）。一度は更新を戻して原因を確かめ、再適用した。
3. npm 10.9.8 と 10.9.9 は vitest 4 系の peer 解決で `npm install` が失敗した。npm 11 では成功し、npm 11 で作った lockfile は npm 10.9.8 の `npm ci` で入った。CI の audit gate は無い（`.github/workflows/` 9 本と `scripts/`、Go 側に脆弱性検査が無い）。
4. bundle の実測。`@preserve` を足した後の worker は +334 bytes（コメント行のみ）、web の bundle は bit 同一。Drizzle の schema 経由で worker の Total Upload（非圧縮）が約 +47 KiB（総量約 1010 KiB）、gzip が約 +10 KiB。上限は非圧縮 64 MiB で gzip の上限は無く、startup は 1 秒（公式 docs）。startup 時間は未実測。
5. Drizzle の A は、当初 `schema.ts` を C に回していたが、境界契約（A）の範囲に導入ライブラリの宣言・config 結線・schema・定数が入るため、A に含めた。手書きの 0001 SQL を生成物（`drizzle-kit generate`）に置き換え、本番と smoke の両 wrangler config に `migrations_dir`／`migrations_pattern` を宣言した。pattern の相対 path は config file の dir が基準だと実測した。
6. `@cloudflare/workers-types` を tsconfig に足すと web の `playback-rpc-client.ts` で typecheck が 5 件落ちたため、不採用にして戻した。drizzle-kit の RC は列側の `primaryKey()` で NOT NULL を落としたため、表側で宣言した。
7. Issue の label を taxonomy に揃えた（#193〜#202 の 10 件。epic 2 件は `epic` を新設して付与）。旧 label の定義（`type: feat` など）は repo に残してある。
8. 本 session の gate は緑だった。unit 69 file・547 test（branches 100%）、integration 8 file・41 test（2 回連続）、typecheck・lint・layers。e2e と smoke は実行していない。
9. 実装 Issue の起票は、skew・版つき永続・CAS・cursor を #217〜#220、Drizzle を #222・#223 とした。wiki #192 は、解消した行の削除、件数上限を Backlog へ移動、Drizzle の行を #222・#223 へ結ぶ更新を行った。lesson #211 へ 8 行を追記した。
10. 2 つの subagent が session の利用上限（HTTP 429）で途中終了した。どちらも `SendMessage` で再開して完了した。
11. 未確認として残した点。remote D1（本番・TEST）への適用、複数 migration の昇順適用と台帳の記録名、drizzle-kit が RC の間の layout の安定性、worker bundle の増加を許容するか、worker の startup 時間。
12. `~/settings/agent-standards` の変更（hook の block reason、`documentation` の `wiki`・`issue`・`commit`、workflow の `wiki`）は、非 scope の指示により commit していない。`AGENTS.md` の test 名の行の置換は、hook に止められたので shim の適用待ち。

### Commits

- `5be4502`
- `8d458c1`
- `18c1f27`
- `42f4c39`
