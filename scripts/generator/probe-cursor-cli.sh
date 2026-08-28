#!/usr/bin/env bash
# name: probe-cursor-cli
# description: Cursor CLI を GitHub Actions runner で非対話実行し、install 結果・実行 binary の絶対 path・
#              現行 argv の可否・child environment 要件を 1 case 単位で観測する probe。
# @require リポジトリ内から呼ぶ。引数に case 名(full / no-home / no-tmpdir / minimal-path)を 1 個渡す。
#          CURSOR_API_KEY が environment に設定済み。curl と bash が使える runner。
# @ensure 与えられた 1 case だけを実行し、入力条件・install_exit・run_exit・binary 絶対 path・
#         stdout / stderr byte 数・分類結果を stdout(および GITHUB_STEP_SUMMARY が設定されていれば両方)へ metadata だけ append する。
#         run 系が失敗しても script 自体は exit 0 で返す。install 失敗と呼び出し方の不正だけ exit 1。
# @invariant CURSOR_API_KEY の値・prompt 本文・stdout 本文・stderr 本文を stdout / summary / artifact へ一切出さない。
#            constants.go と同一 argv を使い、prompt は無害な 1 文に固定する。
#            probe 専用 artifact。production tree へ残さない(Issue Phase 4 で削除)。
set -euo pipefail

# what: stderr byte 数がこの値以上なら service、未満(かつ 0 超)なら entitlement とみなす境界。
#       サーバ側エラー応答本文・stack trace は数百 byte を超え、認可拒否メッセージは短文に収まる経験則。
readonly STDERR_SERVICE_THRESHOLD=512

# 失敗段階を分類する純粋関数。exit code と install 成否だけで機械的に 1 語へ写す。
# 入力: install_exit run_exit stderr_byte_count
# 出力: installer / success / argv / environment / service / entitlement のいずれか 1 語
# 注記: 出力は run_exit と stderr byte 数からの機械推定であり、暫定・補助ラベル。
#       service / environment / entitlement の境界は実測前の経験則(Issue Risks 1)。
classify_failure() {
  install_exit="$1"
  run_exit="$2"
  stderr_byte_count="$3"

  # why: install が完了しない限り run 系の入力は評価対象にならないため最優先で分離する。
  if [ "$install_exit" -ne 0 ]; then
    printf 'installer\n'
    return 0
  fi

  # why: 成功と失敗を観測結果として区別する(Issue Contract 4)。
  if [ "$run_exit" -eq 0 ]; then
    printf 'success\n'
    return 0
  fi

  # why: exit 2 は多くの CLI で未知フラグ・引数不整合全般の汎用 usage error。
  #      --sandbox / --mode / --model のどれが起因かは区別不能なため、現行 argv 全体が
  #      当該 CLI バージョンで不整合、と機械分類する(model 名解決不能を含む)。
  #      model 起因かどうかは stderr の数値と人間の後追いで判断する。
  if [ "$run_exit" -eq 2 ]; then
    printf 'argv\n'
    return 0
  fi

  # why: 実行はされたが診断出力ゼロで落ちるのは HOME/PATH/TMPDIR 等 child environment 欠落での起動失敗。
  if [ "$stderr_byte_count" -eq 0 ]; then
    printf 'environment\n'
    return 0
  fi

  # why: stderr 本文が閾値以上なのはサーバ側エラー応答本文・stack trace。Cursor 側の一時障害と切り分ける。
  if [ "$stderr_byte_count" -ge "$STDERR_SERVICE_THRESHOLD" ]; then
    printf 'service\n'
    return 0
  fi

  # why: 閾値未満の短い stderr は認可拒否の短文メッセージ(API key 無効 / plan 不足)。
  printf 'entitlement\n'
  return 0
}

# case 名から child environment の env コマンド接頭辞を配列 probe_env_prefix へ組み立て、
# あわせて入力条件の説明文を probe_env_desc へ設定する。未知 case は return 1 で弾く。
# full: HOME/PATH/TMPDIR を全指定 / no-home: HOME を外す / no-tmpdir: TMPDIR を外す /
# minimal-path: PATH を /usr/bin:/bin のみへ絞る
#
# why not env -i: runner の生環境を丸ごとクリアすると Cursor CLI が runner 依存の変数を
#   要求する場合に観測不能になる。本 probe は HOME/PATH/TMPDIR を落とす/絞る 2 値観測に限定し、
#   最小 child environment の厳密な特定は実測後の後続 scope とする(Issue Notes 1)。
#   このため full と no-home の差は HOME だけにはならず、親環境由来の変数は両 case へ継承される。
build_env_prefix() {
  case_name="$1"
  probe_home="$2"
  probe_path="$3"
  probe_tmpdir="$4"

  case "$case_name" in
    full)
      probe_env_prefix=(env "HOME=$probe_home" "PATH=$probe_path" "TMPDIR=$probe_tmpdir")
      probe_env_desc="HOME/PATH/TMPDIR 全指定"
      ;;
    no-home)
      probe_env_prefix=(env -u HOME "PATH=$probe_path" "TMPDIR=$probe_tmpdir")
      probe_env_desc="HOME 未指定(env -u HOME)"
      ;;
    no-tmpdir)
      probe_env_prefix=(env -u TMPDIR "HOME=$probe_home" "PATH=$probe_path")
      probe_env_desc="TMPDIR 未指定(env -u TMPDIR)"
      ;;
    minimal-path)
      probe_env_prefix=(env "HOME=$probe_home" "PATH=/usr/bin:/bin" "TMPDIR=$probe_tmpdir")
      probe_env_desc="PATH=/usr/bin:/bin のみ"
      ;;
    *)
      printf 'error: 未知の case=%s\n' "$case_name" >&2
      return 1
      ;;
  esac
  return 0
}

# 入力条件と観測結果を stdout へ必ず出し、あわせて GITHUB_STEP_SUMMARY が設定されていればそこへも append する。
# why: GitHub Actions 上では GITHUB_STEP_SUMMARY が常設で、summary(UI 目視専用)へしか出さないと job log から
#      プログラム回収できない(gh CLI に Step Summary 取得コマンドは無い)。stdout へ出すことで
#      `gh run view --log | grep 'probe-cursor-cli case='` で機械回収できる。
# stdout 本文・stderr 本文・prompt 本文・secret 値は引数に取らない(Issue Contract 3 / @invariant)。
append_summary() {
  summary_line_case="$1"
  summary_line_env_desc="$2"
  summary_line_install_exit="$3"
  summary_line_run_exit="$4"
  summary_line_binary_path="$5"
  summary_line_stdout_bytes="$6"
  summary_line_stdout_lines="$7"
  summary_line_stderr_bytes="$8"
  summary_line_stderr_lines="$9"
  summary_line_classification="${10}"

  {
    printf '### probe-cursor-cli case=%s\n' "$summary_line_case"
    printf -- '- 入力条件: %s\n' "$summary_line_env_desc"
    printf -- '- install_exit: %s\n' "$summary_line_install_exit"
    printf -- '- run_exit: %s\n' "$summary_line_run_exit"
    printf -- '- binary 絶対 path: %s\n' "$summary_line_binary_path"
    printf -- '- stdout byte 数: %s\n' "$summary_line_stdout_bytes"
    printf -- '- stdout 行数: %s\n' "$summary_line_stdout_lines"
    printf -- '- stderr byte 数: %s\n' "$summary_line_stderr_bytes"
    printf -- '- stderr 行数: %s\n' "$summary_line_stderr_lines"
    printf -- '- 分類結果(暫定): %s\n' "$summary_line_classification"
    printf -- '- 注記: 分類は run_exit と stderr byte 数からの機械推定。service/environment/entitlement の境界は実測前の暫定値。\n'
  } | tee -a "${GITHUB_STEP_SUMMARY:-/dev/null}"
}

# 公式 install 手順を実行し、解決した binary の絶対 path を返す。
# install の exit code は global 変数 install_exit へ、binary 絶対 path は resolved_binary へ入れる。
run_install() {
  installer_file="$(mktemp)" || { printf 'error: mktemp 失敗(installer_file)\n' >&2; return 1; }

  # why: curl と bash の exit を分離する。pipe だと install_exit が右辺 bash の exit のみ拾い、
  #      curl 失敗(network / 404)を握り潰す(Issue Contract 4 / Risks 1)。
  set +e
  curl -fsS https://cursor.com/install -o "$installer_file"
  curl_exit=$?
  set -e

  if [ "$curl_exit" -ne 0 ]; then
    # why: curl 失敗時は installer を実行しないまま、curl の非ゼロ値を install_exit とする。
    install_exit="$curl_exit"
    rm -f "$installer_file"
    resolved_binary=""
    return 0
  fi

  set +e
  bash "$installer_file" > /dev/null 2>&1
  install_exit=$?
  set -e
  rm -f "$installer_file"

  # why: 公式手順は cursor-agent を置くが constants.go の BinaryName は "agent"。両名を command -v で探索する。
  resolved_binary=""
  candidate_path="$(command -v cursor-agent 2>/dev/null || true)"
  if [ -z "$candidate_path" ]; then
    candidate_path="$(command -v agent 2>/dev/null || true)"
  fi
  if [ -n "$candidate_path" ]; then
    resolved_binary="$candidate_path"
  fi
}

# constants.go / text_writer.go buildCursorArgs() と同一順の argv。secret は載せない。
build_cursor_args() {
  cursor_args=(-p --mode ask --output-format json --model composer-2.5 --sandbox enabled --trust)
}

main() {
  if [ "$#" -ne 1 ]; then
    printf 'usage: %s <case: full|no-home|no-tmpdir|minimal-path>\n' "$(basename "$0")" >&2
    return 1
  fi
  case_name="$1"

  # why: probe は repo root を cwd とした非対話実行を観測する(Issue Acceptance 2)。
  #      呼び出し位置に依存せず cwd を repo root へ固定する。
  root="$(git rev-parse --show-toplevel)"
  cd "$root"
  probe_home="${HOME:-/root}"
  probe_path="${PATH}"
  probe_tmpdir="${TMPDIR:-/tmp}"

  # why: case の妥当性検証を build_env_prefix へ一本化する(S-1)。未知 case はここで exit 1。
  #      install より前に弾くことで、不正 case で公式 install 手順(network 副作用)を走らせない。
  if ! build_env_prefix "$case_name" "$probe_home" "$probe_path" "$probe_tmpdir"; then
    printf 'error: env prefix 構築失敗 case=%s\n' "$case_name" >&2
    return 1
  fi

  # why: CURSOR_API_KEY 未設定のまま空実行すると entitlement 誤分類になる(Issue @require / Constraints 4)。
  #      値は表示せず precondition で弾く。
  : "${CURSOR_API_KEY:?CURSOR_API_KEY が未設定。Secrets 登録を確認せよ}"
  # why: secret を argv へ載せず、子 process へは環境変数の継承だけで届ける(Issue Contract 3 / Constraints 4 / @invariant)。
  #      env コマンドは -i 未使用のため親環境を継承する。argv には CURSOR_API_KEY を一切書かない。
  export CURSOR_API_KEY

  run_install
  # why: install 失敗は run を試す意味がないため、分類だけ summary へ残して exit 1 で返す。
  if [ "$install_exit" -ne 0 ]; then
    classification="$(classify_failure "$install_exit" 0 0)"
    append_summary "$case_name" "install 段階で失敗" "$install_exit" "-" "-" "-" "-" "-" "-" "$classification"
    return 1
  fi

  if [ -z "$resolved_binary" ]; then
    # why: install は exit 0 でも binary を解決できなければ installer 段階の失敗として扱う。
    classification="$(classify_failure 1 0 0)"
    append_summary "$case_name" "install 後に binary 未解決" "$install_exit" "-" "(未解決)" "-" "-" "-" "-" "$classification"
    return 1
  fi

  build_cursor_args

  stdout_file="$(mktemp)" || { printf 'error: mktemp 失敗(stdout_file)\n' >&2; return 1; }
  stderr_file="$(mktemp)" || { printf 'error: mktemp 失敗(stderr_file)\n' >&2; rm -f "$stdout_file"; return 1; }

  # why: prompt はヒアドキュメント固定で stdin 投入する。log へ echo しない(Issue Constraints 4)。
  #      stdout も stderr と同様 tmpfile へ受け、byte / 行数の数値だけ取る。本文は読まない(Issue Risks 3)。
  set +e
  "${probe_env_prefix[@]}" \
    "$resolved_binary" "${cursor_args[@]}" \
    <<'PROBE_PROMPT' \
    2> "$stderr_file" \
    > "$stdout_file"
1たす1を数字だけで答えよ
PROBE_PROMPT
  run_exit=$?
  set -e

  # why: stdout / stderr とも本文を読まず、byte 数と行数の数値だけ保持する(Issue Contract 3 / Risks 3)。
  stdout_bytes="$(wc -c < "$stdout_file" | tr -d ' ')"
  stdout_lines="$(wc -l < "$stdout_file" | tr -d ' ')"
  stderr_bytes="$(wc -c < "$stderr_file" | tr -d ' ')"
  stderr_lines="$(wc -l < "$stderr_file" | tr -d ' ')"
  rm -f "$stdout_file" "$stderr_file"

  classification="$(classify_failure "$install_exit" "$run_exit" "$stderr_bytes")"

  append_summary "$case_name" "$probe_env_desc" "$install_exit" "$run_exit" \
    "$resolved_binary" "$stdout_bytes" "$stdout_lines" "$stderr_bytes" "$stderr_lines" "$classification"

  # why: 1 case の run 失敗で matrix の残りを巻き込まないため exit 0 で返す。分類が成果物(Issue Contract 2 / 4)。
  return 0
}

# why: test から source される時は main を実行しない。直接実行時だけ main へ引数を渡す。
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
  main "$@"
fi
