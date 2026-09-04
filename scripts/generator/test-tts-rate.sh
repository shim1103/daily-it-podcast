#!/usr/bin/env bash
# name: generator-test-tts-rate
# description: Gemini TTS の PASS 率・所要を計測する dispatch 専用 test を実行する。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @require TEST_GEMINI_API_KEY 相当の値が GEMINI_API_KEY env に渡っている（無ければ test 側で Skip）。
# @ensure `system ratemeasure` tag の TestGeminiTTSRate だけ実行する。他 System test を巻き込まない。
# @invariant Unit / Integration gate を呼ばない。cron gate（test-system.sh）に載せない。secret 値を log に出さない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "tts-rate: generator (go)"
(
  cd "$root/apps/generator"
  go test -v -tags="system ratemeasure" -run TestGeminiTTSRate -timeout 60m -count=1 ./test/system/...
)
