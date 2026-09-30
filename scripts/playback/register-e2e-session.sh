#!/usr/bin/env bash
# name: playback-register-e2e-session
# description: 普段の browser で Access に入場した後の CF_Authorization（JWT）から Playwright の
#              storageState を組み立て、GHA Secret PLAYWRIGHT_STORAGE_STATE_JSON へ登録する。
# @require リポジトリ内から呼ぶ。gh がログイン済みで、この repo の Secret を書ける。Node（apps/playback/.nvmrc）が PATH にある。
# @require 普段の browser の DevTools（Application の Cookies）から、本番 origin の CF_Authorization の値をコピー済み。
# @ensure JWT の exp を期限とする storageState を、標準入力経由で Secret へ登録する。表示するのは期限だけ。
# @invariant cookie の値を argv・file・log に出さない（標準入力と変数だけで渡す）。
#            既に失効した値・JWT でない値は登録しない（Secret は変更されない）。
set -euo pipefail

command -v gh >/dev/null || { echo "gh が見つからない" >&2; exit 1; }

root="$(git rev-parse --show-toplevel)"

origin="${PLAYWRIGHT_BASE_URL:-}"
if [[ -z "$origin" ]]; then
  read -r -p "本番の origin（https://...）: " origin
fi
read -r -s -p "CF_Authorization の値（貼り付け。画面には出ません）: " token
echo >&2

state_json="$(
  cd "$root/apps/playback"
  printf '%s' "$token" | node --experimental-strip-types --no-warnings test/support/access-session-cli.ts build "$origin"
)" || { echo "storageState を組み立てられなかった。Secret は変更していない" >&2; exit 1; }

printf '%s' "$state_json" | gh secret set PLAYWRIGHT_STORAGE_STATE_JSON
echo "登録した。確認: gh workflow run playback-e2e.yml --ref master" >&2
