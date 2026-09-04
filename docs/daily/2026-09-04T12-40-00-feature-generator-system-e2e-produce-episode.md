---
name: TTS retry 予算を Adapter へ集約し port を SynthesizeAll 化、gemini package を責務分割、rate 計測を本番相当尺 + TTS_DOUBLE 化
date: 2026-09-04T12:40:00
session_id: none
branch: feature/generator-system-e2e-produce-episode
prev: 2026-09-03T14-45-00-feature-generator-system-e2e-produce-episode.md
---

## 1. Summary

前 session の設計相談（`AskPlan`）の続き。shim の 3 論点のうち論点 1（Node20 deprecation）は task file 起票のみ、論点 2・3 を executor へ委譲して実装した。`port.SpeechSynthesizer` を `Synthesize(text)` から `SynthesizeAll(texts []string)` へ変え、retry 予算・callGap・RPD quota を Gemini Adapter が「1 度の呼び出し = 1 episode 分の TTS 呼び出し群」で束ねる二段構え（内側 `MaxAttempts` / 外側 `SynthesizeBudget=15`）にした。`Timeline` / `ConcatWAV` は application に残置（案 C = Adapter が原稿束ね構造を知る層違反、として却下）。Gemini が返す極小 PCM（`minPCMBytes` 未満）を `decode_pcm` 相当の retryable にし、Adapter が「非空・最小尺の WAV」を contract 保証、計測 test 側の非空 WAV / 尺 > 0 の再確認と `ProduceEpisode` の `dur <= 0` 防御（到達不能）を削除。判断は Decision `2026-09-04T12-40-00`。あわせて shim 方針で gemini package（`synthesizer.go` 398 行に 5 責務同居）を `synthesizer.go` / `constructor.go` / `retry.go` / `backoff.go` / `transport.go` へ move only 分割し、各 file に対応 test file 1 本・共有 fake は `fake_test.go` へ集約（RUN 数 44 一致で挙動不変を担保）。`tts_rate_system_test.go` は計測対象を「先頭 1 束（~21s）」から本番 topic 束へ引き上げ、`TTS_DOUBLE`（空/min/tgt/max、既定 max、未知値は `t.Fatalf`）で尺帯を選べるようにし、既定値の SSOT を `generator-tts-rate.yml` の `inputs.default` へ一本化した。lessons へ 8 件追記。

## 2. Changes

- 論点 2・3 の実装は executor へ 1 回委譲。manager 相当の監査として `go build` / `go test ./...`（broad integration 込み 105s）/ `go vet`（`system` / `system ratemeasure`）/ `go test -race` / `golangci-lint`（build tag 付き含む）を再実行、全緑を確認して採用。gemini package 分割・rate 計測の尺調整・`TTS_DOUBLE` 化は main で直接実装（TDD、move only は RUN 数一致で担保）。
- commit は 9 本。`8ab057e`（論点 2+3。executor が `docs/lessons/index.md` へ知見 3 件を指示外で同梱、内容が論点そのものなので採用）、`7fef8a1`（Node24 bump task 起票）、`bd5bb9b`（gemini package 分割）、`d6c70f4`→`cb2605c`（rate 計測を Min → Max 尺へ往復。shim 指摘「Max まで確かめろ」で Max に確定）、`fb8a3c9`→`5789a41`（`TTS_DOUBLE` 追加 → env パーサと「パーサの test」3 本を「過剰・暗黙 fallback」指摘で全廃し switch 直書き + fail-fast へ）、`5c25b5c`（precondition 下限を検算せず 8→7 割へ緩めていたのを検算のうえ 8 割へ戻し）。
- 実 dispatch は shim 指示で `generator-tts-rate`（TTS 単体）のみ。run 33780794360（`double` 前・先頭 1 束 ~21s × 10）PASS 10/10・平均 20.9s、run 33828810131（本番 topic 束 Max ~141s / 986 rune × 10）PASS 10/10・平均 71.8s。両 run とも 429 / retry / MAX_TOKENS 0。長尺特有の失敗は 146s 圏内では観測されず。所要は 30〜104s と 3 倍以上ぶれる（synthesize 自体が非線形）。`generator-system` の e2e 前通し（`ProduceEpisode.Run` 全経路）は port 変更後まだ dispatch していない（引き継ぎ file 「次」3 へ）。
- gemini package の公開 API（`NewSpeechSynthesizer*` / `Tuning` / 定数 / `Error`）と Domain 定数（`entities/constants` の `Draft*` / `CharsPerSecond`）は全変更を通して不変。`composition/` は port 型を渡すだけなので無変更。`generator-produce-episode.yml`（本番 workflow）は不触。
- pre-commit hook が `bd5bb9b` で 1 度 Bash 2 分 timeout（playback vitest 途中、システム負荷）→ HEAD 未更新を確認し timeout 7 分へ延長して成立。hook 自体は正常（既知の lessons どおり）。
- 引き継ぎ `generator-system-e2e-produce-episode.md` の実装済節・memo・次を現状（`SynthesizeAll` / 二段予算 / package 分割 / `TTS_DOUBLE`）へ同期。

### Commits

- `8ab057e`
- `7fef8a1`
- `bd5bb9b`
- `d6c70f4`
- `cb2605c`
- `fb8a3c9`
- `5789a41`
- `5c25b5c`
