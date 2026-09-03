---
name: Generator System gate を TEST key 固定・1 回通しへ縮小し、PASS 率計測を dispatch 専用 test へ分離
date: 2026-09-03T14:45:00
session_id: none
branch: feature/generator-system-e2e-produce-episode
prev: 2026-09-02T21-37-48-feature-generator-system-e2e-produce-episode.md
---

## 1. Summary

前 session で用途別に分割した System suite を、shim の指示に沿ってさらに縮小した。`generator-system.yml` の credential を Cursor / Gemini 含め全て `TEST_*` へ固定し、`workflow_dispatch` の実行対象パターン切替と inline Cursor smoke を撤去。cron 週次 + dispatch は `speech_synthesis` / `cursorcli_draft`(1 回版) を 1 回ずつ通すだけとし、N 回ループも `-count=N` も入れない。Drive を fake する `gemini_excluded_full` と full の `produce_episode`、その専用 helper `drive_observe.go` を削除し `//go:build system && full` を畳んだ。PASS 率・所要の計測は `//go:build system && ratemeasure` の dispatch 専用 test（`tts_rate` / `draft_rate`）+ 専用 workflow に残し、「1 回通しが落ちた後の切り分け」の位置づけにした。計測 test は `TEST_GEMINI_API_KEY` / `TEST_CURSOR_API_KEY` を直読みし本番 key が計測へ流れる経路を塞ぐ。付随して Gemini TTS の待機系パラメータ（callGap / retryBackoffBase / retryBackoffMax）を field 化して注入可能にし、`build.ComposeBriefWithTemplate` を分離、Cursor CLI install と workflow の集計 shell を script へ切り出した。判断は Decision `2026-09-03T14-45-00` / `14-46-00` / `14-47-00` / `16-30-00`。lessons へ 6 件追記。

## 2. Changes

- 実装は executor へ 2 回委譲（1 回目: tuning 注入 / ComposeBriefWithTemplate / install script / rate test + workflow / TEST key 化。2 回目: full 系 3 file 削除 / rate test の env 名 TEST_ 直読み / 集計 shell の script 化）。各回 manager が `go build` / `go test ./internal/...` / `go vet -tags=system` / `-tags="system ratemeasure"` / `go list` の tag files / shell test を再実行して監査、いずれも採用。
- `go list -tags=system` の TestGoFiles = `cursorcli_draft` + `speech_synthesis` の 2 本。`-tags="system ratemeasure"` で `tts_rate` + `draft_rate` が加わる。`system && full` / `drive_observe` / `geminiExcluded` は grep 0 件。
- Decision `2026-09-03T16-30-00` は先行 3 本（`14-45-00` / `14-46-00` / `14-47-00`）の supersede 範囲を §4 に明示。先行 file 本文は不触。
- `docs/tasks/todo/generator-system-rate-workflows.md`（前段で作った実装 Issue 下書き）は方針転換で削除。`generator-system-pass-rate.md` は rate 計測を「1 回通し失敗後の切り分け」へ位置づけ直し。引き継ぎ `generator-system-e2e-produce-episode.md` と `generator-lane.md` を現状へ同期。
- commit は `--repo --split`（skill_options.json）で repo 内全変更を意味単位 7 分割。type は `constants.toml` 準拠（workflow 系は `ci` が無いため `chore`）。commit 4 で pre-commit が 2 回 Bash timeout（システム負荷で hook が 5 分超）→ timeout 10 分へ延長して成立。hook 自体は正常。
- Domain 定数（`entities/constants` の `Draft*` / `CharsPerSecond` / 秒数 / `ModelID`）、`gemini.MaxAttempts`、`config` パッケージは全変更を通して不変。本番 workflow（`generator-produce-episode.yml`）は不触。
- `TEST_CURSOR_API_KEY` / `TEST_GEMINI_API_KEY` の repo Secret 登録は人手作業として残置。未登録だと TEST key 化後の実 API 部分と新 2 workflow は Skip / 失敗する。

### Commits

- `d43cb4b`
- `d8e4657`
- `610676e`
- `9f30681`
- `1818435`
- `ce2502f`
- `f187111`
