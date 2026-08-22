#!/usr/bin/env bash
# name: generator-check-static
# description: generator の静的検査と build（depguard / errcheck / govet / gofmt / go build）を実行する。
# @require リポジトリ内から呼ぶ。golangci-lint が PATH にある。apps/generator が存在する。
# @ensure golangci-lint と go build が exit 0 のときだけ成功する。
# @invariant Unit / Integration を実行しない。playback を触らない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "static: generator (depguard / errcheck / govet / gofmt / build)"
(
  cd "$root/apps/generator"
  golangci-lint run ./...
  go build ./...
)
