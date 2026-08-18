---
name: Gemini TTS Adapter を PCM→WAV で実装する
date: 2026-08-18T13:21:00
session_id: none
branch: feature/tts-speech-synthesizer
prev: 2026-08-17T18-17-35-feature-tts-speech-synthesizer.md
---

## 1. Summary

Gemini Developer API の TTS 戻り raw PCM を Adapter 内で WAV に wrap し、`SpeechSynthesizer` を Composition から結線できるようにした。Drive 契約の音声拡張子を wav に揃え、Adapter Issue draft を完了扱いで削除した。mp3 encoder 依存は入れない。

## 2. Changes

1. `infrastructure/speech/gemini` に HTTP・retry・PCM decode・WAV wrap と sociable unit を追加し Composition から注入
2. Port / `SpeechAudio` / Drive 契約 / README の戻り形式を WAV に更新
3. 薄い Error method を coverage 除外へ追加（twitterapiio と同型）
4. `generator-lane` の Gemini TTS を完了にし、`gemini-tts-adapter.md` を削除
5. playback worker 着手用の session memo を todo に残した

### Commits

1. `4ed6ca6` — feat(generator): Gemini TTS Adapter を PCM→WAV で実装する
2. `017f0c9` — docs: Drive 音声形式を WAV にし Adapter 完了を記録する
3. `27fa0eb` — docs: playback worker 着手用の session memo を残す
4. 本 commit — docs(log): セッションログ
