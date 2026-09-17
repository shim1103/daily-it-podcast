---
name: workflow ymlのcredential命名統一とTTS composition結線の完成
date: 2026-09-17T10:55:31
session_id: none
branch: feature/generator-workflow-fallback-wiring
prev: なし
---

## 1. Summary

Decision 2026-09-16T00-39-21で確定したcredential役割区分（値が本番と異なるものだけTEST_接頭辞を持つ）へ、`generator-system.yml`・`generator-draft-rate.yml`のsecret/variable参照名を統一した。あわせて`generator-system.yml`に`SYSTEM_TEST_TOPIC_COUNT`のdispatch入力を追加した。

またTTS composition結線（`internal/composition/gemini.go`）をPrimary（free）/Spare（paid）の2要素へ分割し、`produce_episode.go`のspeech合成layerへ`SPARE_GEMINI_API_KEY`のfinal-fallbackとして結線した（T2で「2つ目のfallback source追加は別task」として先送りされていた分）。

このPRはfeature/generator-textwriter-adapter-fallback branchで進めていた一連の作業を、develop向けstacked PR（#178 TextWriter fallbackをbaseにした4番目のPR）へ分割する過程で切り出したもの。作業中、`generator-tts-rate.yml`・`generator-draft-rate.yml`のsystem test実装（`tts_rate_system_test.go`・`draft_rate_system_test.go`）が意図的にfallback UseCase非経由の単体API計測のままであることを確認し、これらのworkflowへ`SPARE_GEMINI_API_KEY`を追加しない判断をした（§3 Deviations相当）。

## 2. Changes

1. `generator-system.yml`のGoogle OAuth/Drive TEST_credential参照（developで既に撤去済み）と`TEST_CURSOR_API_KEY`/`TEST_SPARE_GEMINI_API_KEY`/`TEST_R2_ACCOUNT_ID`を、Decision準拠の`CURSOR_API_KEY`/`SPARE_GEMINI_API_KEY`/`R2_ACCOUNT_ID`（値が本番と同一のため統合）へ置換
2. `generator-system.yml`に`workflow_dispatch.inputs.topic_count`を追加し、`SYSTEM_TEST_TOPIC_COUNT`環境変数として渡す
3. `generator-draft-rate.yml`の`TEST_CURSOR_API_KEY`を`CURSOR_API_KEY`へ統一（`draft_rate_system_test.go`が実際に読む環境変数名に合わせた）
4. `internal/composition/gemini.go`のTTS結線を`newGeminiSpeechSynthesizerPrimary`（free/TierFree）・`newGeminiSpeechSynthesizerSpare`（paid/TierPaid）の2関数へ分割
5. `internal/composition/produce_episode.go`のspeech合成layerへSpareを2つ目のfallback sourceとして追加
6. `generator-tts-rate.yml`はCursor credentialを持たず`SPARE_GEMINI_API_KEY`もtest実装が使わないため、変更対象外と判断（現状維持）

### Commits

- `ac37c28`
- `5a27b3a`
