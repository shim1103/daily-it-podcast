#!/usr/bin/env bash
# name: playback-test-integration
# description: playback の Integration Test を実行する。
# @require リポジトリ内から呼ぶ。Playback は npm 依存が install 済み。
# @ensure package.json が無ければ成功。あれば npm run test:integration が exit 0。
# @invariant Unit 専用 suite を再実行しない。generator を触らない。本番 credential を読まない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "integration: playback (vitest)"
(
  cd "$root/apps/playback"
  if [[ ! -f package.json ]]; then
    echo "skip: package.json なし（空 package）"
    exit 0
  fi
  npm run test:integration
)
