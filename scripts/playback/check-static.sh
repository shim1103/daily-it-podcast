#!/usr/bin/env bash
# name: playback-check-static
# description: playback の静的検査（Biome + tsc --noEmit）を実行する。
# @require リポジトリ内から呼ぶ。playback は npm 依存が install 済み。
# @ensure package.json が無ければ成功。あれば Biome format/lint と tsc --noEmit が exit 0 のときだけ成功する。
# @invariant Unit / Integration を実行しない。generator を触らない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

(
  cd "$root/apps/playback"
  if [[ ! -f package.json ]]; then
    echo "skip: package.json なし（空 package）"
    exit 0
  fi
  echo "format: playback (biome)"
  npm run format:check
  echo "lint: playback (biome)"
  npm run lint
  echo "typecheck: playback (tsc)"
  npm run typecheck
)
