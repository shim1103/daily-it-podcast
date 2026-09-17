---
name: OAuth/Drive codebase削除・frontend sortバグ修正・runtime diagram全面見直し
date: 2026-09-17T09:40:00
session_id: none
branch: feature/r2-post-cutover-verify-oauth
prev: なし
---

## 1. Summary

`r2-post-cutover-verify-oauth`Issueの残タスクとして、generator/playbackからGoogle OAuth/Drive実装を削除し、Issue fileを完了削除した。並行してfrontendのepisode一覧date降順sortバグとR2 Adapterのエラー診断性を修正し、runtime構成図をR2移行後の実態（技術的に正確なCloudflare Workers/R2の関係）へ全面的に描き直した。

## 2. Changes

1. generator/playbackからGoogle OAuth/Drive実装一式を削除（executor委譲＋残存コメント修正・in-memory-episode-repositoryのディレクトリ移設）
2. `listEpisodes`にepisode一覧のdate降順sortを追加し、契約（`ListEpisodesResponseSchema`）にも並び順を明記
3. R2 Adapterの応答形式エラー診断を改善。executor実装（`isR2ObjectBodyLike`+`describeR2ObjectBodyLikeMismatch`の2関数・2回呼び出し・`undefined`兼用）へのshim指摘を受け、`checkR2ObjectBodyLike`という単一tagged union関数へ設計をやり直した
4. GHA workflow（`generator-produce-episode.yml`・`generator-system.yml`）から削除済みGoogle Secret/Variable参照を除去
5. `DEPLOY.md`・`DESIGN.md`・`README.md`・lane indexをR2移行完了状態へ更新
6. runtime構成図をshimの技術的指摘に基づき全面見直し。R2をWorkers isolateの外（独立storage service）に配置し、generator経路（Internet越しS3互換API）とplayback経路（Worker binding、no Internet hop）を区別。Access層のTLS終端・routing責務、Vite buildの成果物と実行時の区別、TypeScriptの型共有edgeを復元、5 sourceからGo CLIへの矢印を1本に集約、全ラベルを英語化
7. Issue file（`r2-post-cutover-verify-oauth.md`）を完了削除

### Commits

- `e86c831`
- `cb01f1d`
- `3318cea`
- `46dbfda`
- `aea81ba`
- `31aa097`
- `4b30dc3`
