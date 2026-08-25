#!/usr/bin/env bash
# name: generator-test-unit
# description: generator の Unit Coverage（statement）を計測し、除外後に閾値未満なら失敗する。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @ensure 除外後の statement coverage が閾値以上のときだけ exit 0。
# @invariant Integration package（./test/...）を計測しない。playback を触らない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
threshold_pct="${GENERATOR_UNIT_COVER_THRESHOLD:-90}"
profile="$(mktemp)"
filtered="$(mktemp)"
trap 'rm -f "$profile" "$filtered"' EXIT

echo "unit: generator (go + coverage)"
(
  cd "$root/apps/generator"
  go test ./cmd/... ./internal/... -shuffle=on -count=1 -covermode=atomic "-coverprofile=$profile"
)

# why: Composition Root と CLI Driving Adapter（cmd）は結線・入口だけ。error.go / names.go / constants.go は名前では除外しない。
awk -v out="$filtered" '
  NR == 1 { print > out; next }
  $0 ~ /\/internal\/composition\// { next }
  $0 ~ /\/cmd\// { next }
  { print > out }
' "$profile"

# why: coverprofile の import path 解決に module root が要る。
summary="$(
  cd "$root/apps/generator"
  go tool cover -func="$filtered" | awk '/^total:/ { print $(NF) }'
)"
pct="${summary%\%}"
awk -v pct="$pct" -v want="$threshold_pct" 'BEGIN {
  if (pct + 0 < want + 0) {
    printf "generator unit coverage %s%% < %s%% (除外後 statement)\n", pct, want > "/dev/stderr"
    exit 1
  }
  printf "generator unit coverage %s%% >= %s%% (除外後 statement)\n", pct, want
}'
