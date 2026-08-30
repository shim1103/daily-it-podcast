#!/usr/bin/env bash
# name: playback-test-e2e
# description: playback の browser E2E を実行する。
# @require リポジトリ内から呼ぶ。Playback は npm 依存が install 済み。Playwright browser が用意済み。
# @ensure package.json が無ければ成功。あれば npm run test:e2e が exit 0。
# @invariant Integration gate（test-integration.sh）を呼ばない。Unit を再実行しない。generator を触らない。
# @invariant 本入口は必須 Unit / Integration gate 外である。本番 credential 値を script 内に持たない・読まない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "e2e: playback (playwright)"
(
  cd "$root/apps/playback"
  if [[ ! -f package.json ]]; then
    echo "skip: package.json なし（空 package）"
    exit 0
  fi
  npm run test:e2e
)
