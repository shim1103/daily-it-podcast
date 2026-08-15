#!/usr/bin/env bash
# name: test-unit
# description: 全言語の Unit Test を実行する入口。
# @require リポジトリ root から呼ぶ。Go が PATH にある。Playback は npm 依存が install 済み。
# @ensure generator の Unit と playback の Unit がすべて pass する。対象 package が空なら成功扱い。
# @invariant Integration 以上を実行しない。本番 credential を読まない。
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"

echo "unit: generator (go)"
(
  cd "$root/apps/generator"
  go test ./cmd/... ./internal/...
)

echo "unit: playback (vitest)"
(
  cd "$root/apps/playback"
  if [[ ! -f package.json ]]; then
    echo "skip: package.json なし（空 package）"
    exit 0
  fi
  npm run test:unit
)
