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

# Given: PROBE_SANDBOX_MODE 未設定で build_cursor_args を呼ぶ / When: cursor_args を空白連結する / Then: constants.go buildCursorArgs() と同一 argv
# 根拠: constants.go / buildCursorArgs() の argv がズレたら probe が古い argv で回るのを防ぐ。
#       probe の目的は現行 argv の可否観測なので、argv 定義がズレたまま観測すると結果が無意味になる。
#       PROBE_SANDBOX_MODE 未設定は enabled 相当で、--sandbox enabled が --trust の前に来る現状順を厳密維持する。
unset PROBE_SANDBOX_MODE
build_cursor_args
assert_equals "build_cursor_args は PROBE_SANDBOX_MODE 未設定なら constants.go と同一順の argv を組む" \
  "-p --mode ask --output-format json --model composer-2.5 --sandbox enabled --trust" \
  "${cursor_args[*]}"

# Given: PROBE_SANDBOX_MODE=enabled で build_cursor_args を呼ぶ / When: cursor_args を空白連結する / Then: 未設定時と完全一致
# 根拠: enabled は現状の constants.go 完全再現。明示指定でも --sandbox enabled が --trust の前に来る順を維持する。
PROBE_SANDBOX_MODE=enabled build_cursor_args
assert_equals "build_cursor_args は PROBE_SANDBOX_MODE=enabled で constants.go と同一順の argv を組む" \
  "-p --mode ask --output-format json --model composer-2.5 --sandbox enabled --trust" \
  "${cursor_args[*]}"

# Given: PROBE_SANDBOX_MODE=disabled で build_cursor_args を呼ぶ / When: cursor_args を空白連結する / Then: --sandbox disabled を含む argv
# 根拠: sandbox enabled が AppArmor 非対応で GHA 非動作(run 33160392008 で確定)。disabled 化で argv 残りの GHA 可否を観測する。
#       enabled 同様 --sandbox 指定は --trust の前へ挿入し、値だけ disabled へ替える。
PROBE_SANDBOX_MODE=disabled build_cursor_args
assert_equals "build_cursor_args は PROBE_SANDBOX_MODE=disabled で --sandbox disabled を --trust の前に置く" \
  "-p --mode ask --output-format json --model composer-2.5 --sandbox disabled --trust" \
  "${cursor_args[*]}"

# Given: PROBE_SANDBOX_MODE=off で build_cursor_args を呼ぶ / When: cursor_args を空白連結する / Then: --sandbox フラグ自体が無い argv
# 根拠: --sandbox フラグを完全に外した場合の GHA 可否も観測対象。off は基本 argv のみを組む。
PROBE_SANDBOX_MODE=off build_cursor_args
assert_equals "build_cursor_args は PROBE_SANDBOX_MODE=off で --sandbox フラグを付けない" \
  "-p --mode ask --output-format json --model composer-2.5 --trust" \
  "${cursor_args[*]}"

# Given: PROBE_SANDBOX_MODE=bogus (不正値) で build_cursor_args を呼ぶ / When: 戻り値を見る / Then: 非ゼロ(fail-fast)
# 根拠: 取りうる値は enabled / disabled / off の3種のみ。未知値は誤った argv での観測を招くため呼び出し時点で弾く。
if PROBE_SANDBOX_MODE=bogus build_cursor_args 2>/dev/null; then
  sandbox_mode_bogus_result="ゼロ(誤)"
else
  sandbox_mode_bogus_result="非ゼロ(正)"
fi
assert_equals "build_cursor_args は不正な PROBE_SANDBOX_MODE を非ゼロで弾く" \
  "非ゼロ(正)" "$sandbox_mode_bogus_result"
unset PROBE_SANDBOX_MODE

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

# Given: GITHUB_STEP_SUMMARY が実ファイルに設定済み(GitHub Actions 実挙動)で append_summary を呼ぶ /
# When: stdout を拾う / Then: case ブロック先頭行が stdout にも出る
# 根拠: GitHub Actions 上では GITHUB_STEP_SUMMARY が常設のため、summary へしか出さないと job log からプログラム回収できない。
#       gh CLI に Step Summary 取得コマンドは無い。summary へ append しつつ stdout へも必ず出し、
#       `gh run view --log | grep 'probe-cursor-cli case='` で拾える不変を固定する。
append_summary_tmp_summary="$(mktemp "${TMPDIR:-/tmp}/probe-cursor-cli-test.XXXXXX")"
append_summary_stdout="$(GITHUB_STEP_SUMMARY="$append_summary_tmp_summary" append_summary "full" "test" 0 0 "/x" 1 1 2 2 "success" "" "enabled")"
if printf '%s\n' "$append_summary_stdout" | grep -q '^### probe-cursor-cli case=full$'; then
  append_summary_stdout_result="stdout に出力あり(正)"
else
  append_summary_stdout_result="stdout に出力なし(誤)"
fi
assert_equals "append_summary は GITHUB_STEP_SUMMARY 設定時でも case ブロック先頭行を stdout へ出す" \
  "stdout に出力あり(正)" "$append_summary_stdout_result"

# Then: GITHUB_STEP_SUMMARY ファイルへも従来どおり append される(UI 目視用の出力を失わない)
if grep -q '^### probe-cursor-cli case=full$' "$append_summary_tmp_summary"; then
  append_summary_file_result="summary に出力あり(正)"
else
  append_summary_file_result="summary に出力なし(誤)"
fi
assert_equals "append_summary は GITHUB_STEP_SUMMARY 設定時にそのファイルへも従来どおり append する" \
  "summary に出力あり(正)" "$append_summary_file_result"
rm -f "$append_summary_tmp_summary"

# Given: 12 個目引数へ sandbox モード文字列を渡して append_summary を呼ぶ /
# When: stdout を拾う / Then: `- sandbox 指定: <モード>` 行が出る
# 根拠: どの sandbox モードで観測したかを記録に残さないと、後から run 結果とモードの対応が取れなくなる。
#       PROBE_SANDBOX_MODE の切替を summary へ 1 行で刻む。判明後 PROBE_SANDBOX_MODE ごと削除する一時措置。
append_summary_sandbox_stdout="$(GITHUB_STEP_SUMMARY=/dev/null append_summary "full" "test" 0 1 "/x" 0 0 2 2 "argv" "" "disabled")"
if printf '%s\n' "$append_summary_sandbox_stdout" | grep -q "^- sandbox 指定: disabled$"; then
  append_summary_sandbox_result="sandbox 行あり(正)"
else
  append_summary_sandbox_result="sandbox 行なし(誤)"
fi
assert_equals "append_summary は 12 個目の sandbox モード文字列を sandbox 指定行として stdout へ出す" \
  "sandbox 行あり(正)" "$append_summary_sandbox_result"

# Given: 12 個目引数を省略して append_summary を呼ぶ / When: stdout を拾う / Then: sandbox 指定行は enabled 相当で出る
# 根拠: 12 個目は省略時に enabled 相当。既存呼び出し互換のためデフォルト値を持たせる。
append_summary_sandbox_default_stdout="$(GITHUB_STEP_SUMMARY=/dev/null append_summary "full" "test" 0 0 "/x" 1 1 2 2 "success")"
if printf '%s\n' "$append_summary_sandbox_default_stdout" | grep -q "^- sandbox 指定: enabled$"; then
  append_summary_sandbox_default_result="enabled 相当で出力(正)"
else
  append_summary_sandbox_default_result="enabled 相当で出力されず(誤)"
fi
assert_equals "append_summary は 12 個目省略時に sandbox 指定行を enabled 相当で出す" \
  "enabled 相当で出力(正)" "$append_summary_sandbox_default_result"

# Given: PROBE_REVEAL_STDERR=1 相当の開示文字列を 11 個目引数として渡して append_summary を呼ぶ /
# When: stdout を拾う / Then: `- stderr 先頭300byte(開示):` 行が出る
# 根拠: 259 byte stderr の失敗理由(キー未読込 / entitlement / trust / その他)を確定するため、
#       開示フラグ ON のときだけ stderr 先頭を 1 回開示する一時分岐を固定する。secret 値・prompt 本文・stdout 本文は非開示のまま。
append_summary_reveal_stdout="$(GITHUB_STEP_SUMMARY=/dev/null append_summary "full" "test" 0 1 "/x" 0 0 259 2 "entitlement" "Error: Authentication required. Please run 'agent login' first" "enabled")"
if printf '%s\n' "$append_summary_reveal_stdout" | grep -q "^- stderr 先頭300byte(開示): "; then
  append_summary_reveal_result="開示行あり(正)"
else
  append_summary_reveal_result="開示行なし(誤)"
fi
assert_equals "append_summary は 11 個目の開示文字列が非空なら stderr 先頭300byte 開示行を stdout へ出す" \
  "開示行あり(正)" "$append_summary_reveal_result"

# Given: 11 個目引数へ空文字を渡して append_summary を呼ぶ(PROBE_REVEAL_STDERR 未設定相当) /
# When: stdout を拾う / Then: 開示行は一切出ない(default 不変・Contract 3 維持)
# 根拠: 開示は opt-in。フラグ未設定時は従来と完全に同じ挙動(stderr 本文非出力)でなければならない。
append_summary_noreveal_stdout="$(GITHUB_STEP_SUMMARY=/dev/null append_summary "full" "test" 0 1 "/x" 0 0 259 2 "entitlement" "" "enabled")"
if printf '%s\n' "$append_summary_noreveal_stdout" | grep -q "stderr 先頭300byte(開示)"; then
  append_summary_noreveal_result="開示行あり(誤)"
else
  append_summary_noreveal_result="開示行なし(正)"
fi
assert_equals "append_summary は 11 個目の開示文字列が空文字なら開示行を出さない" \
  "開示行なし(正)" "$append_summary_noreveal_result"

printf '\n合計 %s 件 / 失敗 %s 件\n' "$test_total" "$test_failed"

if [ "$test_failed" -ne 0 ]; then
  exit 1
fi
