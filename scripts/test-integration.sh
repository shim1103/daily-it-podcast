#!/usr/bin/env bash
# name: test-integration
# description: 全系統の Integration Test 入口。片系 script を呼ぶだけ。
# @require リポジトリ内から呼ぶ。片系 test-integration が存在する系統の前提を満たす。
# @ensure 両系統の Integration がすべて pass する。対象 package が空なら成功扱い。
# @invariant toolchain を直接実行しない。Unit 専用 suite を再実行しない。本番 credential を読まない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

"$root/scripts/generator/test-integration.sh"
"$root/scripts/playback/test-integration.sh"
