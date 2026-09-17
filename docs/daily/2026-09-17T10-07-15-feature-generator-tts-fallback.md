---
name: TTSのmodel切り替えfallbackと部分成功保持を実装
date: 2026-09-17T10:07:15
session_id: none
branch: feature/generator-tts-fallback
prev: なし
---

## 1. Summary

TTSのGemini呼び出しに、TextWriterと同型のmodel切り替えfallback機構（`application/speech`合成layer）を新設した。RPD=10という無料枠を1エピソードで焼き切ると途中まで合成できたsegmentごと失われていた問題に対し、`port.SpeechSynthesizer.SynthesizeAll`の契約を「失敗時も部分成功を返す」へ変更し、fallback発火時に無駄なquota消費が起きない構造にした。あわせて`fetchPCM`のGemini枯渇判定を、response bodyの事後文字列走査からHTTPレスポンス直読みの一元判定へ置き換えた。

このPRは元のfeature/generator-textwriter-adapter-fallback branchで進めていた一連の作業を、develop向けstacked PR（#176 credential decisionをbaseにした2番目のPR）へ分割する過程で切り出したもの。

## 2. Changes

1. `application/speech`パッケージを新設し、TextWriterの`manuscript.TextWriter`と同型の合成layer（`sources []port.SpeechSynthesizer`、単一fallbackループ）を実装
2. `port.SpeechSynthesizer.SynthesizeAll`の契約を変更し、失敗時もそれまでに合成できた`[]models.SpeechAudio`を返す（部分成功保持）ようにした
3. `speech/gemini`の`fetchPCM`が、HTTPレスポンス（401/403、response bodyのquota_exceeded相当）を見た時点で`port.ErrSourceExhausted`を直接wrapする設計へ変更し、旧`wrapIfSourceExhausted`（事後の文字列走査）を廃止
4. `Tier`（`TierFree`/`TierPaid`）を`SpeechSynthesizer`のconstructor引数として明示化
5. `callGap`（20秒→6秒）・`SynthesizeBudget`（15→10）をAI Studio実測値（RPM=10, RPD=10）に基づき見直し
6. `composition/gemini.go`・`composition/produce_episode.go`をTier引数・合成layer経由の結線へ追従
7. TTS fallback設計のdecisionを追加
8. PR #177としてdevelop向けstacked PR（base: feature/generator-credential-fallback-decision）を作成、push・CI確認まで完了

### Commits

- `2eb5213`
- `83aeae0`
