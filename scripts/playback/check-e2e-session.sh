#!/usr/bin/env bash
# name: playback-check-e2e-session
# description: env PLAYWRIGHT_STORAGE_STATE_JSON の Access session の期限を確認する。失効していれば失敗させる。
# @require リポジトリ内から呼ぶ。Node（apps/playback/.nvmrc）が PATH にある。
# @ensure env が未設定・session が有効・失効間近のときは exit 0（期限と残り日数を表示）。
#         失効・JSON として読めない・Access cookie が無いときは exit 1。
# @invariant cookie の値を log に出さない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

(
  cd "$root/apps/playback"
  node --experimental-strip-types --no-warnings test/support/access-session-cli.ts check
)
