---
name: playback worker の Preview URL を恒久的に無効化する
date: 2026-09-04T23:15:00
branch: feature/playback-e2e-redeploy-master
---

## 1. Decision

`apps/playback/wrangler.jsonc` に `"preview_urls": false` を明示する。`workers.dev` route が有効な間、deploy のたび自動生成される Preview URL を発行させない。

## 2. Reason

1. `workers.dev` route を有効にしたまま `preview_urls` を wrangler config で明示しないと、wrangler は deploy ごとに既定で Preview URL を発行する（`wrangler deploy` 実行時の warning で顕在化）。Preview URL は本番と同じ worker code・同じ Google Drive 読取を持つが、Cloudflare Access のセッション状態や route 設定が本番 hostname と揃っている保証がない別 origin であり、意図せず認証の薄い入口を増やす。
2. daily-it-podcast は個人利用 podcast で、複数 environment（staging preview 等）を運用する要件がない。Preview URL は「今使っている value」が無いまま常時生成される攻撃面の純増でしかない。
3. wrangler の既定挙動（未設定なら有効）に依存すると、将来 wrangler の default が変わった時にも気づかず有効なままになりうる。config に明示することで、この worker の意図（Preview URL は要らない）を設定として固定する。

## 3. Rejected

1. warning を無視して既定のまま運用する — 未使用の別 origin が Cloudflare 側に増え続け、Access の適用漏れがあれば認証を経ずに worker（Drive 読取）へ到達できる経路になる。攻撃面を明示的に閉じられるのに開けたままにする理由がない。
2. `workers.dev` route 自体を無効化し custom domain だけにする — 本 Decision は Preview URL の要否だけを扱う。route 構成全体の見直しは別軸で、現状 `workers.dev` route を使う運用（Access で保護済み）を変える動機が無い。
