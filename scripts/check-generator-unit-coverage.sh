#!/usr/bin/env bash
# name: check-generator-unit-coverage
# description: generator の Unit Coverage（statement）を計測し、除外後に閾値未満なら失敗する。
# @require リポジトリ root から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @ensure 除外後の statement coverage が閾値以上のときだけ exit 0。
# @invariant Integration package（./test/...）を計測しない。playback を触らない。
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
threshold_pct="${GENERATOR_UNIT_COVER_THRESHOLD:-90}"
profile="$(mktemp)"
filtered="$(mktemp)"
trap 'rm -f "$profile" "$filtered"' EXIT

(
  cd "$root/apps/generator"
  go test ./cmd/... ./internal/... -covermode=atomic "-coverprofile=$profile"
)

# why: Composition Root 配線と Error の薄い method は gate 対象外（decision 参照）。
# coverprofile 1 行目は mode。以降は file:span statements count。
awk -v out="$filtered" '
  NR == 1 { print > out; next }
  $0 ~ /\/internal\/composition\// { next }
  $0 ~ /\/infrastructure\/x\/twitterapiio\/error\.go:/ { next }
  $0 ~ /\/infrastructure\/speech\/gemini\/error\.go:/ { next }
  { print > out }
' "$profile"

# why: coverprofile の import path 解決に module root が要る。
summary="$(
  cd "$root/apps/generator"
  go tool cover -func="$filtered" | awk '/^total:/ { print $(NF) }'
)"
# summary は "82.8%" 形式。
pct="${summary%\%}"
awk -v pct="$pct" -v want="$threshold_pct" 'BEGIN {
  if (pct + 0 < want + 0) {
    printf "generator unit coverage %s%% < %s%% (除外後 statement)\n", pct, want > "/dev/stderr"
    exit 1
  }
  printf "generator unit coverage %s%% >= %s%% (除外後 statement)\n", pct, want
}'
