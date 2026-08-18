#!/usr/bin/env bash
# name: generator-check-static
# description: generator の静的検査（現行は depguard）を実行する。
# @require リポジトリ内から呼ぶ。golangci-lint が PATH にある。apps/generator が存在する。
# @ensure golangci-lint が exit 0 のときだけ成功する。
# @invariant Unit / Integration を実行しない。playback を触らない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "lint: generator (depguard)"
(
  cd "$root/apps/generator"
  golangci-lint run ./...
)
