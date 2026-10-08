---
name: Gemini TTS 3.8 Flash への適合とクライアント PCM 変換コード全廃、WAV 再生尺算出の実装
date: 2026-10-08T14:29:01
session_id: cursor-generator-tts-3-8-flash-e6cc
branch: cursor/generator-tts-3-8-flash-e6cc
prev: 2026-10-07T16-20-28-fix-generator-retry-by-structured-error.md
---

## 1. Summary

epic #225 の AC 7（Gemini TTS を 3.8 Flash 仕様へ適合し、出力 WAV を直接処理して PCM 変換コードを削除する）を達成した。Gemini TTS endpoint を `gemini-3.8-flash-tts` へ変更し、`user_input` 構造化 JSON によるリクエストと `audio/wav` の直接デコードへ切り替えた。不要となったクライアント側 PCM wrap コード（`pcm_to_wav.go`、テスト、fuzz target、corpus）を完全削除し、WAV header から再生尺を算出する `wavDurationSec` を新規実装した。関連する既存 Decision 4 件を更新し、静的解析・全単体テスト・結合テスト・fuzz テストがすべて正常に通過することを確認した。

## 2. Changes

1. `gemini-3.8-flash-tts` へのリクエスト形式（`user_input` / `speech_metadata` annotation）およびレスポンス（`audio/wav`）へのアダプター適合
2. クライアント側の PCM ヘッダ付与処理・定数・テストコード（`pcm_to_wav.go`、sociable unit test、fuzz test、corpus）の完全削除
3. RIFF/WAVE header をパースして再生尺（秒）を算出する純粋関数 `wavDurationSec` の実装、単体テスト、マジックナンバーの定数化、および fuzz test の作成
4. `docs/decisions/` 内の 4 件の Decision レコード（TTS WAV 直出し、パフォーマンス、E2E 再生尺、CI ファジング）の整合
5. `./scripts/generator/check-static.sh`（0 issues）、`./scripts/generator/test-unit.sh`（coverage 92.1% >= 90%）、`./scripts/generator/test-integration.sh`、および `go test -run=FuzzWAVDuration -fuzz=FuzzWAVDuration -fuzztime=5s` による動作検証のパス

### Commits

- `a2e1ad6`
- `b9283ba`
- `fbed988`
- `c921616`
