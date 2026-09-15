#!/usr/bin/env bash
# name: generator-test-integration
# description: generator の secret なし Narrow / Broad Integration Test を実行する。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @ensure Integration package が空なら成功。空でなければ go test が exit 0。local S3 gate も exit 0。
# @invariant Unit 専用 suite を再実行しない。本番 credential を読まない。
# @invariant build tag 付き System suite と test-system.sh を実行しない。
# @invariant local S3（-tags r2locals3）は本 script の go test へ混ぜず test-r2-local-s3.sh 経由（Decision §1-5）。
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

# why: httptest NI と併存。gate 正 peer は local S3（Decision 2026-09-15T12-02-48）。失敗は skip しない。
"$root/scripts/generator/test-r2-local-s3.sh"
