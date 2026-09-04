#!/usr/bin/env bash
# name: generator-test-system
# description: generator の System Test を実行する。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @ensure system build tag 付き suite を実行する。対象が空なら成功。
# @ensure PASS 率集計用に -json 出力を /tmp/generator-system.jsonl へ tee する（-v と両立）。
# @invariant Integration gate（test-integration.sh）を呼ばない。Unit を再実行しない。playback を触らない。
# @invariant 本入口は credential 付き実 operation 用である。通常の local 開発・gate では使わない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "system: generator (go)"
(
  cd "$root/apps/generator"
  packages="$(go list -tags=system ./test/system/...)"
  if [ -z "$packages" ]; then
    echo "generator: System package なし（skip）"
  else
    set -o pipefail
    go test -json -v -tags=system -timeout 40m -count=1 ./test/system/... | tee /tmp/generator-system.jsonl
    exit "${PIPESTATUS[0]}"
  fi
)
