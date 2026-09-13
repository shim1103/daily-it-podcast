---
name: 音声を mp3 契約へ固定し R2/cache 方針と実装 Issue を分離した
date: 2026-09-13T14:46:55
session_id: none
branch: feature/playback-now-playing-audio-listenability
prev: なし
---

## 1. Summary

playback の load 遅さに対し、保存・配信形式を mp3 に契約固定し（`EncodeWAVToMP3` stub・未結線）、着手順・R2 完全移行・薄い cache を Decision に残した。mp3 実装は Issue 3 本（1=1 PR）と lane の release 単位へ。R2/cache は Issue 化せず方針 Decision + 未決細部を lane D に置いた。地図文書は現行 Drive のまま将来方針への動線のみ。

## 2. Changes

- 誤って「確認問い（？？）」を edit 許可と取り file を触ったが、指摘後に全 revert。その後明示の `execute A/B` / `Execute C,D` / `/pr-completion` で再適用。
- R2/cache を C Issue にする案は shim 訂正で却下。Decision のみ。
- worker Infra が HTTP contracts から拡張子を import すると dependency-cruiser が落ちるため、値は `drive-layout` 正本・Infra は local 定数へ戻した。
- 検証: commit hook 経由で generator static/unit coverage、playback format/lint/tsc/layers 緑。
- PR は create-pr（`gh pr`）で作成予定。

### Commits

- `0b06dd2`
- `20ab0e9`
- `86dbdf9`
- `ac6fbe8`
