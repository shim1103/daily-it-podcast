---
name: 定期生成と再生 UI の実行基盤を分け、無人 cron に対話 CLI を載せない
date: 2026-08-25T06:58:00
branch: cursor/docs-early-design-decisions-3502
---

## 1. Decision

1. Generator の定期実行（cron）と Playback UI のホスティングは別基盤とする。現在の正は Generator = GitHub Actions cron、Playback = Cloudflare（Workers + 静的 UI）。初期に検討した Vercel 配置は、現 Playback 選定（`2026-08-18T11-12-00`）に置き換わった。
2. 無人の定期実行に Claude Code 等の**対話** CLI を常駐・起動しない。無人経路は非対話の API / CLI（差し替え可能な Port 背後）とする。対話 CLI は初期構築・局所作業に限り関与してよい。
3. UI から生成を起動しない（README の運用）。deploy 前は cron が動かない状態を許容する。

## 2. Reason

1. cron と UI は寿命・秘密・スケール・失敗モードが違う。1 基盤に同居させると直交性が落ちる。
2. 対話 CLI は承認・TTY・session 前提で、無人 cron の標準経路ではない。
3. 生成と再生を UI で結ぶと、2 系統分離（`2026-08-15T16-23-06`）が崩れる。

## 3. Rejected

1. 対話 Claude Code を GHA cron の本体にする案
2. UI ホスティング上で cron 相当の生成を回す案
3. 再生 UI から Generator を直接起動する案
