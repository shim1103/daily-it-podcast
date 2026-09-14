---
name: 成果物の正本 storage は将来 R2 へ完全移行する（Drive+OAuth を廃止）
date: 2026-09-13T14:22:55
branch: feature/playback-now-playing-audio-listenability
---

## 1. Decision

1. 最終形では Google Drive + OAuth refresh をやめ、generator 書込と playback 読取の正本を **Cloudflare R2** にする。
2. 着手は **mp3 一本化の後**（順序の正は Decision `2026-09-13T13-41-00`。本 file は順序を再掲しない）。
3. 本 Decision は **方針**だけを固定する。Issue file（C）には落とさない。実装・契約 stub・cutover 手順は、未決細部が埋まったあとの別 scope とする。
4. 現行 runtime の正本は当面 Drive のまま（地図・DEPLOY の latest 値は実装完了まで Drive）。

## 2. Reason

1. playback の音声 GET は現状、request ごとに OAuth + Drive 全文取得が乗り、refresh token 失効が週次 E2E を落とす（`2026-09-07T22-25-00` Rejected #3 の派生）。R2 は Workers 近傍の object storage で、その依存を消せる。
2. generator と playback が同じ Drive 契約を共有しているため、読取だけ R2 にすると正本が割れる。完全移行（両経路）が一本化になる。
3. 方針を Decision に残し Issue 化を急がないのは、binding・公開形態・credential 注入など **未決がまだ実施契約を書けない**ため（未決は lane D）。

## 3. Rejected

1. playback だけ R2、generator は Drive 残し — 正本が二重になり、配置契約と運用が割れる。
2. Drive を残したまま Worker Cache で OAuth 問題を隠す — token 失効と全文 pull の構造が残る。
3. 本 Decision と同時に R2 の C Issue を起こす案 — bucket/binding/proxy 形態が未決のまま Issue 化すると実装者へ再判断が漏れる（scope-split 禁止）。
