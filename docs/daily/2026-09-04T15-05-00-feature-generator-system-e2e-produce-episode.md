---
name: origin/develop を 2 度取り込み、System suite を e2e 1 回通し + rate 計測 2 本へ組み替え、cursorapi draft rate を実 API dispatch で緑化
date: 2026-09-04T15:05:00
session_id: none
branch: feature/generator-system-e2e-produce-episode
prev: 2026-09-04T12-40-00-feature-generator-system-e2e-produce-episode.md
---

## 1. Summary

`origin/develop` を feature branch へ 2 度取り込み（`df3eb1f` / `b8cdf74`）、その過程で System 用 test を組み替えた。1 度目の取り込みは Cursor CLI → Cloud Agents HTTP API 移行（develop 側 Decision `2026-09-03T17-03-33`）と自分の TTS 側変更が衝突する大きな merge で、`x` / `cursorcli` / `commandlaunch` / `processenv` package の削除は develop に従い、gemini synthesizer は自分の `SynthesizeAll` / rate 計測 field / package 分割を採る一方、`geminiHTTPTimeout` 定数を `composition/runtime.go` から gemini infra（`constants.go` の `httpCallTimeout`、constructor の shallow-copy client で適用）へ移し、`geminiHTTPClient()` を廃して `sharedHTTPClientWithoutTimeout()` 1 本へ集約、`delivery/format.go` の error 分類を `cursorcli` / `processenv` / `getxapi` から `cursorapi` / `hackernews` / `lobsters` / `itmedia` へ差し替えた。cursorcli 前提の system test（`cursorcli_draft_system_test.go` / `draft_rate_system_test.go`）と `generator-draft-rate.yml` 系は最初「削除」で通したが、shim 指摘で cursorapi 前提へ書き直し。さらに shim 方針で System suite を「TTS 単体到達 + Cursor 疎通」から **「e2e 1 回通し 1 本（`TestProduceEpisodeSystem`）+ rate 計測 2 本（`TestGeminiTTSRate` / `TestCursorAPIDraftRate`）」** へ再編し、`speech_synthesis_system_test.go` と `cursorapi_draft_system_test.go` を削除、共有 helper を `system_shared_test.go` へ切り出した。運用方針は `DEPLOY.md` §5 に Generator System / TTS rate / Cursor draft rate の 3 節として集約し、`generator-system-pass-rate.md` 台帳は削除（SSoT は DEPLOY）。`generator-draft-rate.yml` は default branch（`master`）に無いと `workflow_dispatch` が 404 になるため、workflow + script 3 file だけを `master` へ先行配置する PR #125 を出して merge、実 API dispatch で 401 → shim が `TEST_CURSOR_API_KEY` 更新 → 3/3 PASS（run 33840526373）。lessons へ 10 件追記。

## 2. Changes

- **1 度目の merge（`df3eb1f`）**: conflict file は `.github/workflows/generator-system.yml` / `composition/produce_episode.go` / `runtime.go` / `gemini/synthesizer*.go` / `test/integration_support_test.go` / `docs/lessons/index.md` / `docs/tasks/todo/generator-lane.md` ほか。`delivery/format*.go` は develop 側で package ごと消えていたが feature branch にしか無いため merge が残し、deleted package への import が壊れる → error 分類を新 Adapter 群へ書き替えて解消。全 build tag（`system` / `system ratemeasure`）で compile 確認、generator/playback の static・unit・integration・race 全緑で `df3eb1f` を commit。
- **cursorapi draft rate の作り直し（`65b9a1d`）**: 削除前の `draft_rate_system_test.go` を `git show df3eb1f^:` から復元し、`agent` binary / PATH 前提を捨てて `cursorapi.NewTextWriter(&http.Client{}, apiKey)` + `TEST_CURSOR_API_KEY` 直読みへ。環境要因の分母除外は旧 `*cursorcli.Error` `Op=="run"` を `*cursorapi.Error` `Op=="do"`（API へ到達不可）へ読み替え。prompt variant A/B（`testdata/brief_prompt_variant_a.txt`）は復活。
- **System suite 再編（`4fb8abc` / `07022d0` + 2 度目 merge 後の追加）**: shim の「system test 全体は 1 回だけ、PASS 率ではない」指示で `TestProduceEpisodeSystem`（`//go:build system`、`NewProduceEpisodeFromEnv` → `Run` を 1 度、`no_source_items` は PASS 扱い）を新設。`speech_synthesis_system_test.go`（TTS 単体）と `cursorapi_draft_system_test.go`（Cursor 疎通、作った直後）を削除。`seedSourceItems` / `draftTotalRunes` を `system_shared_test.go` へ移動。
- **PR #125（別 branch `drafts/generator-draft-rate-wf`、merge commit `2c0efd6`）**: `master` は `generator-system.yml` / `generator-tts-rate.yml` は持つが rate 系 script を 1 つも持たない状態。`generator-draft-rate.yml` + `test-draft-rate.sh` + `draft-rate-summary.sh` の 3 file を `master` 起点の一時 worktree で作り commit（`9dda649`）→ PR #125 → shim が merge。test 本体は feature branch 側に残し、dispatch 時 `--ref feature/...` で feature の checkout から job を回す。
- **実 API dispatch**: 1 回目（run 33839989187）は `cursorapi: create_status: create status 401` で 3/3 FAIL、0.4s で即死。Cursor docs を調べ、Cloud Agents REST は Dashboard の `crsr_` key 1 種で CLI 用の別枠は無いこと、401 は key の値（expired / malformed / 前後空白混入）が原因と確認。shim が `TEST_CURSOR_API_KEY` を更新 → 2 回目（run 33840526373）は 3/3 PASS（全体 3362 / 3458 / 3723 文字、topic 5 件、63.9〜107.1s/回）。default variant で 1 回が下限 +2 文字とマージンが薄い。
- **2 度目の merge（`b8cdf74`）**: develop が `generator-lane.md` を lean index へ大幅トリム、`.agentsecrets/project.json` を削除、playback web を大きく更新。conflict は `generator-lane.md` のみ。develop の構造を base に自分の済み項目（3 Adapter / cursorapi 移行 / e2e + rate test 配置）を再挿入。`DEPLOY.md` の「default branch（`develop`）」表記は dispatch 404 の実証に反するため `master` へ訂正。playback は `npm ci` 後 static/unit/integration 全緑。
- **docs DRY（`6edd300`）**: 引き継ぎ `generator-system-e2e-produce-episode.md` の「feature → develop PR」項目を除去（本 PR そのもの）、「rate 計測 2 本未 dispatch」を `generator-draft-rate` 3/3 PASS 済みで TTS 側のみへ縮小、尺マージンの「実績 3473」を run 33840526373 の新実測へ差し替え。
- `generator-produce-episode.yml`（本番 workflow）は不触。gemini / cursorapi package の公開 API と Domain 定数は全変更を通して不変。

### Commits

- `df3eb1f`
- `267b07a`
- `4fb8abc`
- `65b9a1d`
- `07022d0`
- `b8cdf74`
- `6edd300`
- `9dda649`（別 branch `drafts/generator-draft-rate-wf` → PR #125 merge commit `2c0efd6`）
