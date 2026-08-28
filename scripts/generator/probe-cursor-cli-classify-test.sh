#!/usr/bin/env bash
# name: probe-cursor-cli-classify-test
# description: probe-cursor-cli.sh の純粋関数(classify_failure / build_cursor_args / build_env_prefix)を
#              対象にした最小 TDD runner。
# @require リポジトリ内から呼ぶ。同じ dir に probe-cursor-cli.sh が存在する。bash が使える。
# @ensure classify_failure の全ケース、build_cursor_args の argv、build_env_prefix の各 case が
#         期待値と一致したときだけ exit 0。1件でも不一致なら exit 1。
# @invariant probe 専用 artifact。production tree へ残さない（Issue Phase 4 で削除）。実 CLI を install/実行しない。
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# why: 対象は副作用のない純粋関数のみ。source で関数だけ読み込み、main 実行は走らせない。
# shellcheck source=/dev/null
source "$script_dir/probe-cursor-cli.sh"

test_total=0
test_failed=0

# 関数名・入力・期待・実際を日本語で表示し、不一致なら test_failed を増やす。
assert_classify() {
  test_name="$1"
  install_exit="$2"
  run_exit="$3"
  stderr_byte_count="$4"
  expected="$5"

  test_total=$((test_total + 1))

  actual="$(classify_failure "$install_exit" "$run_exit" "$stderr_byte_count")"

  if [ "$actual" = "$expected" ]; then
    printf '成功: %s\n' "$test_name"
    printf '  関数=classify_failure 入力=(install_exit=%s run_exit=%s stderr_byte_count=%s) 期待=%s 実際=%s\n' \
      "$install_exit" "$run_exit" "$stderr_byte_count" "$expected" "$actual"
  else
    test_failed=$((test_failed + 1))
    printf '失敗: %s\n' "$test_name"
    printf '  関数=classify_failure 入力=(install_exit=%s run_exit=%s stderr_byte_count=%s) 期待=%s 実際=%s\n' \
      "$install_exit" "$run_exit" "$stderr_byte_count" "$expected" "$actual"
  fi
}

# Given: 公式 install 手順が非ゼロ終了した / When: 分類する / Then: installer になる
# 根拠: install が完了しない限り run 系の入力は評価対象にならないため、install_exit を最優先で判定する。
assert_classify "install が失敗したら installer に分類する" 1 0 0 "installer"
assert_classify "install が失敗すれば run_exit がゼロでも installer を優先する" 127 0 4096 "installer"

# Given: install 成功かつ CLI 実行も成功 / When: 分類する / Then: success になる
# 根拠: 成功と失敗を観測結果として区別する（Issue Contract 4）。
assert_classify "install 成功かつ run 成功なら success に分類する" 0 0 0 "success"

# Given: install 成功だが exit 2 (汎用 usage error) で非ゼロ終了 / When: 分類する / Then: argv になる
# 根拠: exit 2 は多くの CLI で未知フラグ・引数不整合全般の汎用 usage error。--sandbox / --mode / --model の
#       どれが起因かは区別不能なため、argv 全体の不整合として機械分類する。model 名解決不能はこれに含まれ、
#       model 起因かどうかは stderr の数値と人間の後追いで判断する。
assert_classify "run が exit 2 なら現行 argv の不整合として argv に分類する" 0 2 128 "argv"

# Given: install 成功だが CLI が stderr を一切出さず非ゼロ終了 / When: 分類する / Then: environment になる
# 根拠: 実行はされたが診断出力なしで落ちるのは HOME/PATH/TMPDIR 等 child environment 欠落での起動失敗パターン。
assert_classify "run が非ゼロかつ stderr 0 バイトなら environment に分類する" 0 1 0 "environment"

# Given: install 成功だが CLI が大量の stderr を出して非ゼロ終了 / When: 分類する / Then: service になる
# 根拠: stderr 本文が大きいのはサーバ側エラー応答本文や stack trace。Cursor 側の一時障害と切り分ける。
assert_classify "run が非ゼロかつ stderr が閾値以上なら service に分類する" 0 1 512 "service"
assert_classify "run が非ゼロかつ stderr が大量なら service に分類する" 0 1 8192 "service"

# Given: install 成功だが CLI が短い stderr を出して非ゼロ終了 / When: 分類する / Then: entitlement になる
# 根拠: 短い stderr は認可拒否の短文メッセージ(API key 無効 / plan 不足)。service と切り分ける。
assert_classify "run が非ゼロかつ stderr が短いなら entitlement に分類する" 0 1 64 "entitlement"
assert_classify "run が非ゼロかつ stderr が閾値直下なら entitlement に分類する" 0 1 511 "entitlement"

# 関数名・期待・実際を日本語で表示し、不一致なら test_failed を増やす。
# 配列を空白連結した文字列で完全一致を見る。
assert_equals() {
  test_name="$1"
  expected="$2"
  actual="$3"

  test_total=$((test_total + 1))

  if [ "$actual" = "$expected" ]; then
    printf '成功: %s\n' "$test_name"
    printf '  期待=[%s] 実際=[%s]\n' "$expected" "$actual"
  else
    test_failed=$((test_failed + 1))
    printf '失敗: %s\n' "$test_name"
    printf '  期待=[%s] 実際=[%s]\n' "$expected" "$actual"
  fi
}

# Given: build_cursor_args を呼ぶ / When: cursor_args を空白連結する / Then: constants.go buildCursorArgs() と同一 argv
# 根拠: constants.go / buildCursorArgs() の argv がズレたら probe が古い argv で回るのを防ぐ。
#       probe の目的は現行 argv の可否観測なので、argv 定義がズレたまま観測すると結果が無意味になる。
build_cursor_args
assert_equals "build_cursor_args は constants.go と同一順の argv を組む" \
  "-p --mode ask --output-format json --model composer-2.5 --sandbox enabled --trust" \
  "${cursor_args[*]}"

# Given: build_env_prefix を各 case で呼ぶ / When: probe_env_prefix を空白連結する / Then: case ごとの env 接頭辞
# 根拠: no-home で -u HOME が入り HOME= が入らないこと、minimal-path で PATH が /usr/bin:/bin へ絞られることを固定する。
#       env 接頭辞の組み立てがズレると child environment の 2 値観測が別条件の観測になってしまう。
build_env_prefix "full" "/home/probe" "/opt/bin:/usr/bin" "/tmp/probe"
assert_equals "build_env_prefix full は HOME/PATH/TMPDIR を全指定する" \
  "env HOME=/home/probe PATH=/opt/bin:/usr/bin TMPDIR=/tmp/probe" \
  "${probe_env_prefix[*]}"
assert_equals "build_env_prefix full の env_desc は全指定" \
  "HOME/PATH/TMPDIR 全指定" "$probe_env_desc"

build_env_prefix "no-home" "/home/probe" "/opt/bin:/usr/bin" "/tmp/probe"
assert_equals "build_env_prefix no-home は -u HOME を入れ HOME= を入れない" \
  "env -u HOME PATH=/opt/bin:/usr/bin TMPDIR=/tmp/probe" \
  "${probe_env_prefix[*]}"

build_env_prefix "no-tmpdir" "/home/probe" "/opt/bin:/usr/bin" "/tmp/probe"
assert_equals "build_env_prefix no-tmpdir は -u TMPDIR を入れ TMPDIR= を入れない" \
  "env -u TMPDIR HOME=/home/probe PATH=/opt/bin:/usr/bin" \
  "${probe_env_prefix[*]}"

build_env_prefix "minimal-path" "/home/probe" "/opt/bin:/usr/bin" "/tmp/probe"
assert_equals "build_env_prefix minimal-path は PATH を /usr/bin:/bin へ絞る" \
  "env HOME=/home/probe PATH=/usr/bin:/bin TMPDIR=/tmp/probe" \
  "${probe_env_prefix[*]}"

# Given: build_env_prefix を未知 case で呼ぶ / When: 戻り値を見る / Then: 非ゼロ(fail-fast)
# 根拠: S-1 で case 検証を build_env_prefix へ一本化したため、未知 case をここで弾けることを固定する。
if build_env_prefix "unknown-case" "/home/probe" "/opt/bin:/usr/bin" "/tmp/probe" 2>/dev/null; then
  unknown_case_result="ゼロ(誤)"
else
  unknown_case_result="非ゼロ(正)"
fi
assert_equals "build_env_prefix は未知 case を非ゼロで弾く" "非ゼロ(正)" "$unknown_case_result"

printf '\n合計 %s 件 / 失敗 %s 件\n' "$test_total" "$test_failed"

if [ "$test_failed" -ne 0 ]; then
  exit 1
fi
