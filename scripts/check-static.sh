#!/usr/bin/env bash
# name: check-static
# description: 全系統の静的検査入口。片系 script を呼ぶだけ。
# @require リポジトリ内から呼ぶ。片系 check-static が存在する系統の前提を満たす。
# @ensure 呼び先がすべて exit 0。
# @invariant toolchain を直接実行しない。Unit / Integration を実行しない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

"$root/scripts/generator/check-static.sh"
"$root/scripts/playback/check-static.sh"
