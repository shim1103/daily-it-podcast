#!/usr/bin/env bash
# name: generator-segment-texts-to-test-mp3
# description: 完成稿 json から TTS→sec 更新→mp3 し、TEST Drive へ json+mp3 を書く。
# @require リポジトリ内から呼ぶ。Go / ffmpeg が PATH にある。
# @require MANUSCRIPT_JSON / TEST_GEMINI_API_KEY。
# @require TEST_GOOGLE_OAUTH_* / TEST_DRIVE_FOLDER_ID（本 script が正規 Drive env 名へ map）。
# @ensure go run ./hack/segment-texts-to-mp3 の exit code をそのまま返す。
# @invariant 本番 GEMINI / 本番 Drive を読まない。原稿 text は変えず sec だけ更新。細かい verify をしない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

: "${MANUSCRIPT_JSON:?MANUSCRIPT_JSON required}"
: "${TEST_GEMINI_API_KEY:?TEST_GEMINI_API_KEY required}"
: "${TEST_GOOGLE_OAUTH_CLIENT_ID:?TEST_GOOGLE_OAUTH_CLIENT_ID required}"
: "${TEST_GOOGLE_OAUTH_CLIENT_SECRET:?TEST_GOOGLE_OAUTH_CLIENT_SECRET required}"
: "${TEST_GOOGLE_OAUTH_REFRESH_TOKEN:?TEST_GOOGLE_OAUTH_REFRESH_TOKEN required}"
: "${TEST_DRIVE_FOLDER_ID:?TEST_DRIVE_FOLDER_ID required}"

export GOOGLE_OAUTH_CLIENT_ID="${TEST_GOOGLE_OAUTH_CLIENT_ID}"
export GOOGLE_OAUTH_CLIENT_SECRET="${TEST_GOOGLE_OAUTH_CLIENT_SECRET}"
export GOOGLE_OAUTH_REFRESH_TOKEN="${TEST_GOOGLE_OAUTH_REFRESH_TOKEN}"
export DRIVE_FOLDER_ID="${TEST_DRIVE_FOLDER_ID}"

echo "segment-texts-to-test-mp3: manuscript bytes=${#MANUSCRIPT_JSON}"
(
  cd "$root/apps/generator"
  go run ./hack/segment-texts-to-mp3
)
