#!/usr/bin/env bash
# name: playback-deploy
# description: playback を本番 Cloudflare Workers へ deploy する。
# @require リポジトリ内から呼ぶ。Playback は npm 依存が install 済み。CLOUDFLARE_API_TOKEN が env にある。
# @ensure npm run deploy（build + wrangler deploy）が exit 0。
# @invariant secret 値を log に出さない。Variable / Secret（wrangler.jsonc の PlaybackEnv 4 key）の値は
#   ここでは変更しない（Dashboard または `wrangler secret put` の別手順）。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "deploy: playback (wrangler)"
(
  cd "$root/apps/playback"
  npm run deploy
)
