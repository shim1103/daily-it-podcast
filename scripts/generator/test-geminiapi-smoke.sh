#!/usr/bin/env bash
# name: generator-test-geminiapi-smoke
# description: Gemini generateContent 原稿 Adapter（geminiapi.TextWriter）が実 API と 1 回疎通するかを
#              確かめる dispatch 専用 smoke test を実行する。PASS 率も原稿品質も測らない。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @require TEST_GEMINI_API_KEY 相当の値が TEST_GEMINI_API_KEY env に渡っている（無ければ test 側で Skip）。
#          geminiapi package を含む ref（fallback 実装済み branch）で走らせる（master 単独では compile fail）。
# @ensure `system ratemeasure` tag の TestGeminiAPISmoke だけ実行する。他 System test を巻き込まない。
# @invariant Unit / Integration gate を呼ばない。cron gate（test-system.sh）に載せない。secret 値を log に出さない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "geminiapi-smoke: generator (go)"
(
  cd "$root/apps/generator"
  go test -v -tags="system ratemeasure" -run TestGeminiAPISmoke -timeout 10m -count=1 ./test/system/...
)
