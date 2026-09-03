#!/usr/bin/env bash
# name: generator-test-draft-rate
# description: CursorCLI が prompt どおりの draft を返す率を計測する dispatch 専用 test を実行する。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @require TEST_CURSOR_API_KEY 相当の値が CURSOR_API_KEY env に渡っている。PATH 上に `agent` がある（無ければ test 側で Skip）。
# @ensure `system ratemeasure` tag の TestCursorCLIDraftRate だけ実行する。他 System test を巻き込まない。
# @invariant Unit / Integration gate を呼ばない。cron gate（test-system.sh）に載せない。secret 値を log に出さない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "draft-rate: generator (go)"
(
  cd "$root/apps/generator"
  go test -v -tags="system ratemeasure" -run TestCursorCLIDraftRate -timeout 60m -count=1 ./test/system/...
)
