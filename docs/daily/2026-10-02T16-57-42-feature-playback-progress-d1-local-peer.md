---
name: 進捗D1のlocal peer・adapter・Root結線を実装し、epicのApplicationを取り込み、並行制御とcursorの境界(A)を切った
date: 2026-10-02T16:57:42
session_id: d4d2716d-c0b8-40f5-be2c-27fbe4fe1235
branch: feature/playback-progress-d1-local-peer
prev: なし
---

## 1. Summary

Issue #196（進捗D1 local peer本実装・NI・Fake・adapter・Root結線）を実装し、PRへ載せる区切りまで進めた。reviewerのMust fix（r2 modeで進捗D1が未結線でも無言で永続しない代替に落ちる）を、設定不足のErrorへ改めた。epic（#197のApplication）を取り込み、D1 adapterを「merge済みの行を保存するだけ」の新しいPortへ適合させた。そのうえで、競合のCAS・cursor pull・時刻契約・skewの設計をDecisionに記録し、境界（定数・Clock port・cursor契約・Portの版つき書込面・migration）を切った。進捗の永続先をepisodeのmodeから独立した明示設定にし、進捗のin-memory実装を本番treeへ持った。

## 2. Changes

1. 最終gateは緑だった。unit 69 file・548 test（coverageのbranchesは100%）、integration 8 file・41 test、typecheck・lint・layersも緑。pre-pushはintegrationを走らせる。
2. local D1（miniflare）で使い捨てscriptにより実測した（repoには残していない）。1000行の`MAX+1`採番は、索引なしで`rows_read`が1000、索引ありで1。条件付きUPDATEの`meta.changes`は一致で1、不一致で0。`ON CONFLICT DO NOTHING`は新規で1、競合で0。同じ期待値の10並行は勝者が1本。索引ありの書込は`rows_written`が2（表+索引）。本番D1での同じ実測は未実施。
3. docsを調べた。D1は各DBが単一スレッドでクエリを1つずつ処理し、`batch()`はSQL transactionだが送信前に全文が確定している必要がある。Sessions APIを使わなければ全クエリはprimaryで実行される。課金はrows readが百万行/bin/zsh.001、rows writtenが.00で、索引は書込行を1つ増やす。Drizzleのdocsは`@rc`のinstallを案内し、migrationはmigrationごとのdirectoryに`migration.sql`と`snapshot.json`を置く。
4. 取り違えがあった。「Errorのtest fileを消す」指示を、Error classの削除と解釈して`ProgressWriteConflictError`を削除した。revertで復元し、別commitでError専用のtest fileを削除した。
5. epic取り込みの衝突は3 file（Port、Portのtest、Rootのtest）だった。typecheckのhookを通すため、D1 adapterを新Portへ適合させる変更をmerge commitに含めた。
6. 並行して動かしたexecutorは、同じworktreeの別fileを触る単位に限った。commitは呼び出し側が単位ごとに行った。
7. `~/settings/agent-standards`の変更。test名の規約の食い違い（`naming-and-layout` §2.8、`coding-style`、`documentation`）をcommitした（そのrepoの`82cb274`）。PreToolUse hookのblock reasonと`documentation/commit`の本文の分量目安は、非scopeの指示によりcommitせず未commitのまま残した。`AGENTS.md`はhookに直接編集を止められ、置換案を提示した。shimの未commit変更（`coding-style`など）は無変更で残した。
8. GitHub側を更新した。wiki #192のBacklogへ「既存testの英語名化」と、本sessionで増えた未確認事項を追記した。lesson #211へ知見を追記した。Issue #198は、pullを`cursor`版へ一度に切り替える前提に合わせてScope・AC・Contract・Dependenciesを更新した。
9. sandbox起因の事象があった。wrangler/miniflareは`XDG_CONFIG_HOME`・`WRANGLER_REGISTRY_PATH`・`WRANGLER_LOG_PATH`を`$TMPDIR`へ向けると通った。`git push`と`gh`はsandbox外で実行した。
10. 未確認として残した点。本番D1への進捗表の適用手順（本番の`wrangler.jsonc`に`migrations_dir`が無い）、Drizzle生成物をwranglerが読めるか、smokeの進捗をin-memoryにできるか（`createApp()`が永続先のoptionを受けない）。

### Commits

- `98b1475`
- `6b4b33c`
- `6b30eec`
- `b5055c3`
- `87c73c2`
- `f81c36a`
- `c27a2bc`
- `8d48e2f`
- `c6be266`
- `21a0bf2`
- `f1a7b39`
- `9686fac`
- `9d797b1`
- `0db26d4`
- `58ead39`
- `5f77540`
- `b8a6758`
- `54af844`
- `03d97d7`
- `001d431`
- `0940785`
- `27c3296`
- `de3e1e5`
- `1c5a634`
- `9ca0678`
- `a27b904`
- `801bc51`
- `5b4bbc5`
- `7ae9b03`
- `1a318da`
- `aa1d72c`
- `f88d3d9`
- `4cab7aa`
- `6477c26`
- `c0c3c0b`
- `6df25d1`
- `06d9287`
- `8458ccd`
- `50250bd`
- `4d04823`
- `a7d021b`
- `61dae96`
- `8cb0399`
- `0a66806`
- `5519b53`
- `bb737d0`
- `3eebce7`
- `1bab2dc`
- `11939a0`
- `5117cf2`
- `8666b9b`
- `76a9431`
- `e201cb6`
- `3411c84`
- `c3ff841`
- `b6e5ee6`
- `df9007d`
- `dcb978a`
- `9ef7e2d`
- `b667112`
- `94d18b7`
