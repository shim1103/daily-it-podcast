#!/usr/bin/env bash
# name: test-unit
# description: 全系統の Unit Test 入口。composer 契約のあと片系 unit を呼ぶ。
# @require リポジトリ内から呼ぶ。片系 test-unit が存在する系統の前提を満たす。
# @ensure composer 契約を実行し、続けて両系統の Unit がすべて pass する。
# @invariant toolchain を直接実行しない。Integration 以上を実行しない。本番 credential を読まない。
set -euo pipefail

root="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"

root="$(git rev-parse --show-toplevel)"

"$root/scripts/test-gate-composer-sociable-unit.shell"
"$root/scripts/generator/test-unit.sh"
"$root/scripts/playback/test-unit.sh"
