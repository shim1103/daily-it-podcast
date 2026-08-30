---
name: 読み上げ Port 境界 stub と TTS Issue draft
date: 2026-08-17T18:17:35
session_id: none
branch: feature/tts-speech-synthesizer
prev: なし
---

## 1. Summary

読み上げの単位を `SpeechSynthesizer` として固定し、Gemini Adapter の空定数と `MaxAttempts`、秘密名を stub した。設計判断を decision に残し、Adapter 実装の Issue draft を todo へ置いた。GitHub Issue 作成と Adapter HTTP 実装はしていない。

## 2. Changes

1. `SpeechSynthesizer` / `SpeechAudio` / Gemini 空定数 / `MaxAttempts` / `GeminiAPIKeyName` を追加
2. `MaxAttempts >= 1` の sociable unit を追加
3. TTS の Port・空定数・retry 上限・課金方針を decision に記録
4. `gemini-tts-adapter.md` を create-issue template で作成し `generator-lane.md` からリンク

### Commits

1. `f177abc` — feat(generator): 読み上げ Port と Gemini 定数の境界 stub を追加する
2. `4c64495` — docs: TTS 設計判断と Adapter Issue draft を残す
3. 本 commit — docs(log): セッションログ
