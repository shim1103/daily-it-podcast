---
name: generator Drive 保存 Adapter slim化と PR 準備
date: 2026-08-19T19-16-08
session_id: none
branch: feature-generator-drive-storage-adapter
prev: none
---

## 1. Summary

`docs/tasks/todo/generator-drive-storage-adapter.md` の Issue を実装完了させた。`EpisodeWriter.Write` から manuscript schema 検証と OAuth token refresh の実処理を除去し、infra 内 `TokenSource` interface へ分離した。executor による実装後、code-reviewer による査読で未使用定数（`formMIME`・`TokenURL`）・depguard の過剰許可・`go.mod` の不要 require を検出し、後続 Issue（`generator-google-oauth-adapter.md`・`generator-episode-validation.md`）の Scope 記述と突き合わせて是正した。Issue file は完了に伴い削除した。

## 2. Changes

- Port `EpisodeWriter.Write(ctx, episodeID, manuscript, audio)` の signature は不変。Drive REST（list/create/upload）は既存実装のまま name/MIME/byte のみ扱う設計を維持
- `TokenSource` interface（`Token(ctx) (string, error)`）を `gdrive` package 内に新設。本番実装（OAuth refresh）は別 Issue の scope
- schema 検証専用だった `entities/errors` の `EpisodeIDMismatch`・`InvalidManuscript` を削除
- `episode_writer.go` の `@invariant` にあった「mp3 を書かない」は、`SpeechSynthesizer` Port が mp3 → WAV へ設計変更された後の残骸と判明し削除（shim 指摘により発覚）
- 査読後の是正差分は独立した commit には分けず、実装 commit へ含めた
- build / test / lint は独立検証済み: `go build ./...` pass、generator 全 pkg test pass（coverage 93.3%）、`check-static.sh` 0 issues、playback vitest 87 tests pass

### Commits

- `ff470ec`
- `88dc667`
- `0e26f2e`
