#!/usr/bin/env bash
# name: playback-test-unit
# description: playback の Unit Test を実行する。
# @require リポジトリ内から呼ぶ。Playback は npm 依存が install 済み。
# @ensure package.json が無ければ成功。あれば npm run test:unit が exit 0。
# @invariant Integration 以上を実行しない。generator を触らない。本番 credential を読まない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "unit: playback (vitest)"
(
  cd "$root/apps/playback"
  if [[ ! -f package.json ]]; then
    echo "skip: package.json なし（空 package）"
    exit 0
  fi
  npm run test:unit
)
