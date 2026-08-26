#!/usr/bin/env bash
# name: generator-test-integration-local
# description: generator の local-real Integration Test（local_real build tag）を実行する。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @ensure Integration package が空なら成功。空でなければ go test -tags=local_real が exit 0。
# @invariant git hook や GitHub Actions から呼ばない。本番 credential を読まない。local_real tagged suite の local 専用入口である。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "integration-local: generator (go -tags=local_real)"
(
  cd "$root/apps/generator"
  # why: `|| true` で go list 失敗を握りつぶすと、構文・module 不整合が緑になる。
  # why: 空集合は go list が exit 0・stdout 空なので、空だけ skip する。
  packages="$(go list -tags=local_real ./test/...)"
  if [ -z "$packages" ]; then
    echo "generator: local-real Integration package なし（skip）"
  else
    go test -tags=local_real ./test/...
  fi
)
