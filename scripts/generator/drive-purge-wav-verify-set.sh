#!/usr/bin/env bash
# name: generator-drive-purge-wav-verify-set
# description: Drive folder 直下の .wav を削除し、完成ペア（json+mp3）のみ・wav 非残存を確認する。
# @require リポジトリ内から呼ぶ。引数は prod または test。Go が PATH にある。
# @require process environment に GOOGLE_OAUTH_CLIENT_ID / GOOGLE_OAUTH_CLIENT_SECRET /
#          GOOGLE_OAUTH_REFRESH_TOKEN / DRIVE_FOLDER_ID がある（本 script は secret を作らない）。
# @ensure go run ./hack/drive-wav-purge の exit code をそのまま返す。
# @invariant Cursor / Gemini 等の他 credential を読まない。playback を触らない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
target="${1:-}"

if [[ "$target" != "prod" && "$target" != "test" ]]; then
  echo "usage: $0 prod|test" >&2
  exit 1
fi

echo "drive-purge-wav-verify-set: target=${target}"
(
  cd "$root/apps/generator"
  go run ./hack/drive-wav-purge
)
