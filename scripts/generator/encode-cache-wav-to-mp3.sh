#!/usr/bin/env bash
# name: generator-encode-cache-wav-to-mp3
# description: .cache/<prod|test> の wav を本番 Encoder（ffmpeg Adapter）へ渡して mp3 にする。
# @require リポジトリ内から呼ぶ。引数は prod または test。Go / ffmpeg が PATH にある。
# @require CACHE_ROOT 未設定時は repo 根の .cache。
# @ensure hack/encode-cache-wav-to-mp3 の exit code を返す。Drive / secret を読まない。
# @invariant bitrate は Encoder 本体に閉じる（本 script は argv を持たない）。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
target="${1:-}"

if [[ "$target" != "prod" && "$target" != "test" ]]; then
  echo "usage: $0 prod|test" >&2
  exit 1
fi

echo "encode-cache-wav-to-mp3: target=${target}"
(
  cd "$root/apps/generator"
  CACHE_ROOT="${CACHE_ROOT:-}" go run ./hack/encode-cache-wav-to-mp3 "$target"
)
