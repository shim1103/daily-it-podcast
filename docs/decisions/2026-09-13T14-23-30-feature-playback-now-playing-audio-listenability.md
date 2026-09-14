---
name: 音声配信の cache は R2 後の薄い HTTP/CF edge とし、Drive 時代の厚い Worker cache は作らない
date: 2026-09-13T14:23:30
branch: feature/playback-now-playing-audio-listenability
---

## 1. Decision

1. 音声配信の cache 化は **R2 完全移行の後**に行う（順序の正は `2026-09-13T13-41-00`）。
2. 採る形は **薄い**配信 cache（成功応答の `Cache-Control` と Cloudflare edge）。初手の本命を Worker Cache API にしない。
3. Drive を正本のまま残した期間に、Drive 全文取得回避のための **厚い Worker Cache** を先行実装しない。
4. 本 Decision は **方針**だけを固定する。Issue file（C）には落とさない。具体の header 値・edge 設定・Access 下での browser cache 実効は未決（lane D）。

## 2. Reason

1. append-only・ほぼ update なしの音声は長期 cache に向くが、得は主に 2 回目以降。初回 byte 量は mp3、origin 往復は R2 が先に効く。
2. 最終形が R2 + CF edge なら、Drive 時代の厚い Worker cache は捨て工事になりやすい（YAGNI）。
3. Access 配下では browser HTTP cache が効きにくいことがあり、browser だけに賭けない。edge / `Cache-Control` を薄い本線にする。
4. header 秒数や proxy 形態が未決のため、今は Issue 化しない。

## 3. Rejected

1. Drive 時代に Worker Cache API を厚く先作りする案 — R2 後に無効化されやすい。
2. cache を R2 / mp3 より先にやる案 — 初回 UX の主因（サイズ・Drive pull）が残る。
3. 本 Decision と同時に cache の C Issue を起こす案 — `Cache-Control` 値と Access 実測が無いと Acceptance が書けない。
