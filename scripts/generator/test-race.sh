#!/usr/bin/env bash
# name: generator-test-race
# description: generator の Unit package を race detector 付きで実行する。
# @require リポジトリ内から呼ぶ。race detector 対応 Go と apps/generator が存在する。
# @ensure Generator Unit package の go test -race が exit 0 のときだけ成功する。
# @invariant Integration package（./test/...）を実行しない。playback を触らない。本番 credential を読まない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "race: generator (go)"
(
  cd "$root/apps/generator"
  go test -race ./cmd/... ./internal/...
)
