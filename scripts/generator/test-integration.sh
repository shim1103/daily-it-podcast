#!/usr/bin/env bash
# name: generator-test-integration
# description: generator の Integration Test を実行する。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @ensure Integration package が空なら成功。空でなければ go test が exit 0。
# @invariant Unit 専用 suite を再実行しない。playback を触らない。本番 credential を読まない。local_real build tag を渡さない。local-real suite を実行しない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "integration: generator (go)"
(
  cd "$root/apps/generator"
  # why: `|| true` で go list 失敗を握りつぶすと、構文・module 不整合が緑になる。
  # why: 空集合は go list が exit 0・stdout 空なので、空だけ skip する。
  packages="$(go list ./test/...)"
  if [ -z "$packages" ]; then
    echo "generator: Integration package なし（skip）"
  else
    go test ./test/...
  fi
)
