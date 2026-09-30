#!/usr/bin/env bash
# name: playback-test-d1-smoke
# description: 進捗 D1（TEST）へ migration を適用し、binding 面（prepare / bind / run / first）で
#              書く・読む・消す・消えたことを読むが通り、binding が local の模擬でなく remote へ届く
#              ことを確かめる dispatch 専用の疎通 test を実行する。あわせて R2 の EPISODES binding も
#              同じ判定にかける。merge 意味・HTTP 応答形は測らない。
# @require リポジトリ内から呼ぶ。apps/playback の npm 依存（wrangler・vitest）が install 済み。
# @require CLOUDFLARE_API_TOKEN（D1 の編集権限）と CLOUDFLARE_ACCOUNT_ID が env に渡っている。
# @ensure 未適用の migration を TEST D1 へ適用し、test/dispatch の Narrow Integration が exit 0。
# @invariant Unit / Integration gate を呼ばない。cron を持たない。secret 値を log に出さない。
#            本番 D1・本番 bucket に触れない（共有 test/support/wrangler.smoke.jsonc の TEST 資源だけを使う）。
#            probe の行・object は test 側が各 case の後に自前で delete する。
set -euo pipefail

: "${CLOUDFLARE_API_TOKEN:?CLOUDFLARE_API_TOKEN が未設定}"
: "${CLOUDFLARE_ACCOUNT_ID:?CLOUDFLARE_ACCOUNT_ID が未設定}"

root="$(git rev-parse --show-toplevel)"

echo "d1-smoke: playback (認証の診断。token が見ている account と権限)"
(
  cd "$root/apps/playback"
  # why: 権限不足と account 不一致を log から切り分ける診断。account 名はメールを含みうるので伏せる。診断の失敗で疎通を止めない
  npx wrangler whoami 2>&1 | sed -E 's/[[:alnum:]._+-]+@[[:alnum:].-]+/<email>/g' || true
  # why: 読取 API（一覧）が通るか・TEST D1 がこの account に見えるかで、書込権限不足と所属違いを切り分ける
  npx wrangler d1 list --json 2>&1 | sed -E 's/[[:alnum:]._+-]+@[[:alnum:].-]+/<email>/g' || true
)

echo "d1-smoke: playback (migration を TEST D1 へ適用)"
(
  cd "$root/apps/playback"
  npx wrangler d1 migrations apply EPISODE_PROGRESS --remote --config test/support/wrangler.smoke.jsonc
)

echo "d1-smoke: playback (疎通 test)"
(
  cd "$root/apps/playback"
  npx vitest run --config test/dispatch/d1-smoke.vitest.config.mjs
)
