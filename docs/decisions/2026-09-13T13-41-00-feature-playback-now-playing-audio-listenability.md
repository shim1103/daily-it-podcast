---
name: 音声 UX 改善の着手順は mp3 一本化が先。R2 と cache は後続
date: 2026-09-13T13:41:00
branch: feature/playback-now-playing-audio-listenability
---

## 1. Decision

1. 音声の遅さへの着手順は **(1) mp3 一本化 → (2) Google Drive 廃止と R2 完全移行 → (3) cache 化（薄い HTTP / CF edge）** とする。
2. Drive を残したままの **厚い Worker Cache API** や、R2 前の本格 cache 設計は先行しない。
3. R2・cache の実施契約（Issue 化）は本 Decision の範囲外。ここでは **順序だけ**を固定する。

## 2. Reason

1. 初回 play の体感はまず browser が吸う byte 量で決まる。mp3 は Drive のままでも初回から効く。
2. 現状 Worker は request ごとに Drive から音声全文を読む。R2 はその待ちと OAuth を消すが、WAV のままでは browser 転送が大きいまま残る。mp3 後に R2 へ載せる方が object も小さい。
3. cache は主に 2 回目以降の得。最終形が R2 なら Drive 時代の厚い cache は捨て工事になりやすい（YAGNI）。

## 3. Rejected

1. R2 を mp3 より先にやる案 — 転送サイズが残る。OAuth 除去は価値があるが、UX 影響度優先なら圧縮が先。
2. Drive 向け Worker Cache を最初に厚く作る案 — R2 移行で大半が無効化される。
3. mp3・R2・cache を 1 判断に束ねる案 — 境界が混ざり、完了判定と PR が肥大する。
