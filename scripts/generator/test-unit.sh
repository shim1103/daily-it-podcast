#!/usr/bin/env bash
# name: generator-test-unit
# description: generator の Unit Coverage（statement）を計測し、除外後に閾値未満なら失敗する。除外は Composition Root・CLI Driving Adapter・build tag 付き local 実物 suite・Broad Integration 以上。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @ensure 除外後の statement coverage が閾値以上のときだけ exit 0。
# @invariant secret なし Narrow Integration（./test/... の build tag なし）は production code（./cmd/... ./internal/...）のカバー計測に含める。build tag 付き local 実物 suite と Broad Integration 以上は含めない。playback を触らない。本番 credential を読まない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
threshold_pct="${GENERATOR_UNIT_COVER_THRESHOLD:-90}"
profile="$(mktemp "${TMPDIR:-/tmp}/generator-coverage-profile.XXXXXX")"
filtered="$(mktemp "${TMPDIR:-/tmp}/generator-coverage-filtered.XXXXXX")"
trap 'rm -f "$profile" "$filtered"' EXIT

echo "unit: generator (go + coverage)"
(
  cd "$root/apps/generator"
  # why: -coverpkg で instrument 対象を production package（./cmd/... ./internal/...）へ固定し、
  #      実行対象に ./test/...（secret なし Narrow。build tag なしなので local 実物 suite は既定 build 外）を足す。
  #      実プロセス境界を持つ処理（processenv.Launch / buildChildEnv 等）は SU では埋まらず Narrow が全経路を通すため。
  go test ./cmd/... ./internal/... ./test/... -shuffle=on -count=1 -covermode=atomic -coverpkg=./cmd/...,./internal/... "-coverprofile=$profile"
)

# why: Composition Root と CLI Driving Adapter（cmd）は結線・入口だけ。error.go / names.go / constants.go は名前では除外しない。
# why: -coverpkg で ./test/... は instrument 対象外なので profile に test package 行は出ないが、防御で落とす（出なければ no-op）。
awk -v out="$filtered" '
  NR == 1 { print > out; next }
  $0 ~ /\/internal\/composition\// { next }
  $0 ~ /\/cmd\// { next }
  $0 ~ /\/apps\/generator\/test\// { next }
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
