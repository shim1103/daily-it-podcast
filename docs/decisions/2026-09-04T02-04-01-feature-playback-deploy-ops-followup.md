---
name: Playback 本番 Worker は observability を常時有効にする
date: 2026-09-04T02:04:01
branch: feature/playback-deploy-ops-followup
---

## 1. Decision

1. Playback 本番 Worker は Workers Logs 用に `observability` を **常時有効** とする。契約値の正本は `apps/playback/wrangler.jsonc`。本 Decision に字段値を写さない。
2. 即時 tail（`wrangler tail`）は切り分け用の補助とし、永続 log の代替にはしない。
3. traces の追加最適化、sampling の手動調整、Logpush 等の外部 sink は今採用しない。
4. 本番への反映は通常の deploy 手順に従う。本運用後続作業の scope では再 deploy を必須にしない（別 Decision）。

## 2. Reason

1. 未設定だと Dashboard に console が残らず、リアルタイム tail だけになる（既存 lesson）。障害後に「さっき何が起きたか」を追えない。
2. 障害時だけ手動 ON にすると、ON 前の区間が永久に欠ける。常時 ON の方が観測の穴が無い。
3. traces / sampling / Logpush は今の個人利用・同一 origin 運用に対して設定面が増えるだけで、一次の切り分け（HTTP 契約 code + Worker log）にはまだ要らない。

## 3. Rejected

1. 障害時だけ `observability` を手動 ON する案 — ON 前の log が残らず、常時 ON より劣る。
2. skill や Decision に Dashboard クリック手順・CLI の版依存 flag 一覧を固定する案 — platform skill の方針（現行 CLI / 公式で都度確認）と衝突する。手順の latest は `DEPLOY.md`。
3. Logpush や外部 sink を今入れる案 — 依存と権限が増え、目的（Worker log を後から読む）に対して過剰。
