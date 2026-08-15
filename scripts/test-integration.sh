#!/usr/bin/env bash
# name: test-integration
# description: 全言語の Integration Test を実行する入口。
# @require リポジトリ root から呼ぶ。Go が PATH にある。Playback は npm 依存が install 済み。
# @ensure generator の Integration と playback の Integration がすべて pass する。対象 package が空なら成功扱い。
# @invariant Unit 専用 suite を再実行しない。本番 credential を読まない。
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"

echo "integration: generator (go)"
(
  cd "$root/apps/generator"
  packages="$(go list ./test/... 2>/dev/null || true)"
  if [ -z "$packages" ]; then
    echo "generator: Integration package なし（skip）"
  else
    go test ./test/...
  fi
)

echo "integration: playback (vitest)"
(
  cd "$root/apps/playback"
  npm run test:integration
)
