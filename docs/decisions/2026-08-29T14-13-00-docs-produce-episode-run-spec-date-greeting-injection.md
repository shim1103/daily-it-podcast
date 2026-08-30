---
name: 表示用 date と OpeningGreeting 文案は TTS 前に ProduceEpisode が確定する
date: 2026-08-29T14:13:00
branch: docs/produce-episode-run-spec
---

## 1. Decision

1. 原稿 JSON の `date` は CLI `now` を **`entities/constants.DisplayTimeZone`（Asia/Tokyo）で暦日化**した `YYYY-MM-DD` とする。cron は edge 時刻を避ける前提とし、midnight 特例は設けない。
2. **OpeningGreeting の TTS 用文案**は、定数 template と上記 `date` から **`ProduceEpisode` が TTS 前に組み立て**る。Cursor brief / TextWriter 出力には **挨拶文を含めない**。
3. `ClosingFarewell` の最終文言は未決でもよい。置き場は `entities/constants/episode_greetings.go` とする。
4. 固定挨拶 WAV を Drive から読む方式は採用しない。

## 2. Reason

1. 挨拶に暦日を読み上げる要件は、実行日ごとに文案が変わる。ファイル固定 WAV では満たせない。
2. date 決定を Playback や UI に委ねると、Generator と Reader の直交性が崩れる（`2026-08-15T16-23-07`）。
3. 挨拶を Cursor に書かせると、TextWriter 境界が台本方針を知り、定型変更のたび vendor I/O 契約が揺れる。

## 3. Rejected

1. OpeningGreeting を空のまま TTS しない案 — 完成原稿の opening field と音声 opening segment の方針がぶれる。
2. `date` を UTC 暦日で書く案 — 表示契約（JST `YYYY-MM-DD`）と不一致。
3. Greeting 文案を `WriteEpisode` が組み立てる案 — Gate が Builder 方針を持つ。
