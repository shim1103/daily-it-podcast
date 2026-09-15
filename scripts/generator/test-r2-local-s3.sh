#!/usr/bin/env bash
# name: generator-test-r2-local-s3
# description: wrangler experimental local S3 peer の到達 Verification を実行する。
# @require リポジトリ内から呼ぶ。Go と apps/playback の wrangler（npm install）がある。
# @ensure `-tags r2locals3` の peer 到達 NI が exit 0。
# @invariant test-unit に混ぜない。本 script の go test だけが -tags r2locals3 を持つ（Decision §1-5）。
# @invariant 本番 Adapter を local S3 に刺さない。本番 credential を読まない。起動失敗を黙って skip して緑にしない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

wrangler_bin="$root/apps/playback/node_modules/.bin/wrangler"
if [[ ! -x "$wrangler_bin" && ! -f "$wrangler_bin" ]]; then
  echo "generator r2 local S3: wrangler が無い（apps/playback で npm install）" >&2
  exit 1
fi

echo "integration: generator r2 local S3 peer (wrangler experimental)"
(
  cd "$root/apps/generator"
  go test -tags r2locals3 -count=1 ./test/ -run 'LocalS3'
)
