#!/usr/bin/env bash
# name: generator-test-system
# description: generator の System Test を実行する。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @require Cursor CLI の `agent` が PATH で解決できる（GHA では workflow が公式 install する）。
# @ensure system build tag 付き suite を実行する。対象が空なら成功。
# @ensure SYSTEM_TEST_RUN が非空ならその名前パターンの test だけ実行する。
# @invariant Integration gate（test-integration.sh）を呼ばない。Unit を再実行しない。playback を触らない。
# @invariant 本入口は credential 付き実 operation 用である。通常の local 開発・gate では使わない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "system: generator (go)"
(
  cd "$root/apps/generator"
  # why: `|| true` で go list 失敗を握りつぶすと、構文・module 不整合が緑になる。
  # why: 空集合は go list が exit 0・stdout 空なので、空だけ skip する。
  # why: System は別 package。同一 package へ tag だけ足すと Narrow / Broad も -tags=system で再実行される。
  # why: Cursor + Gemini + Drive の実経路は default 10m を超えうる。
  packages="$(go list -tags=system ./test/system/...)"
  if [ -z "$packages" ]; then
    echo "generator: System package なし（skip）"
  else
    # why: SYSTEM_TEST_RUN が指定されたらその test だけ実行する。定時は未設定で全実行。
    run_args=()
    if [ -n "${SYSTEM_TEST_RUN:-}" ]; then
      run_args=(-run "$SYSTEM_TEST_RUN")
    fi
    # why: 各 System test の skip/実行と t.Logf（区間別所要）を CI ログへ出す。
    # why: set -u 下の bash 3.2 は空配列 "${run_args[@]}" を unbound 扱いする。+ 展開で回避する。
    go test -v -tags=system -timeout 40m -count=1 ${run_args[@]+"${run_args[@]}"} ./test/system/...
  fi
)
