#!/usr/bin/env bash
# name: generator-report-condition-coverage
# description: generator Unit package の Boolean condition coverage を report する。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator と contracts が存在する。
# @ensure gobco v1.3.4 の condition coverage report を出力する。
# @invariant threshold を持たない。statement coverage gate を置換しない。hook と GitHub Actions から呼ばない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
workspace="$(mktemp -d)"
trap 'rm -rf "$workspace"' EXIT

echo "condition coverage: generator (gobco v1.3.4; informational)"
cp -R "$root/apps/generator" "$workspace/generator"
(
  cd "$workspace/generator"
  go mod edit -replace github.com/shim1103/daily-it-podcast/contracts="$root/contracts"
  packages="$(go list -f '{{.Dir}}' ./cmd/... ./internal/...)"
  while IFS= read -r package_dir; do
    [ -z "$package_dir" ] && continue
    go run github.com/rillig/gobco@v1.3.4 "$package_dir" |
      REPORT_SOURCE_PREFIX="$workspace/generator" \
        REPORT_DESTINATION_PREFIX="$root/apps/generator" \
        perl -pe 's/\Q$ENV{REPORT_SOURCE_PREFIX}\E/$ENV{REPORT_DESTINATION_PREFIX}/g'
  done <<< "$packages"
)
