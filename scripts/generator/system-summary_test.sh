#!/usr/bin/env bash
# name: generator-system-summary-test
# description: system-summary.sh が fake jsonl から top-level test の pass/fail/skip を数えて
#              summary へ書くことを確認する素 bash test。
# @require リポジトリ内から呼ぶ。bash 3.2+。
# @ensure pass 2 / fail 1 / skip 1 の行を持つ fake jsonl を食わせると、出力に "pass 2" "fail 1" "skip 1" が出る。
# @ensure jsonl が無い場合は "jsonl 無し" 系の行が出る。
# @ensure fail=0 / skip=0 の pass-only jsonl でも exit 0 かつ "pass N" / "fail 0" / "skip 0" を書く。
# @invariant 実ネットワークへ出ない。GITHUB_STEP_SUMMARY を一時 file に差し替える。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
script="${root}/scripts/generator/system-summary.sh"

if [ ! -x "$script" ]; then
  echo "FAIL: ${script} が実行可能でない" >&2
  exit 1
fi

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

# --- case 1: fake jsonl（top-level pass 2 / fail 1 / skip 1、サブテスト行はノイズ） ---
jsonl="${workdir}/generator-system.jsonl"
cat > "$jsonl" <<'JSONL'
{"Action":"pass","Test":"TestAlpha"}
{"Action":"pass","Test":"TestBravo"}
{"Action":"fail","Test":"TestCharlie"}
{"Action":"skip","Test":"TestDelta"}
{"Action":"pass","Test":"TestAlpha/sub_case"}
{"Action":"fail","Test":"TestCharlie/sub_case"}
{"Action":"pass","Test":""}
JSONL

summary="${workdir}/step_summary_1"
: > "$summary"

SYSTEM_JSONL="$jsonl" GITHUB_STEP_SUMMARY="$summary" bash "$script"

out="$(cat "$summary")"
echo "--- case1 output ---"
echo "$out"
echo "--------------------"

if ! grep -q "pass 2" <<<"$out"; then
  echo "FAIL: case1 出力に 'pass 2' が無い" >&2
  exit 1
fi
if ! grep -q "fail 1" <<<"$out"; then
  echo "FAIL: case1 出力に 'fail 1' が無い" >&2
  exit 1
fi
if ! grep -q "skip 1" <<<"$out"; then
  echo "FAIL: case1 出力に 'skip 1' が無い" >&2
  exit 1
fi
if ! grep -q "TestCharlie" <<<"$out"; then
  echo "FAIL: case1 出力に FAIL した 'TestCharlie' が無い" >&2
  exit 1
fi

# --- case 2: jsonl 無し ---
summary2="${workdir}/step_summary_2"
: > "$summary2"

SYSTEM_JSONL="${workdir}/does-not-exist.jsonl" GITHUB_STEP_SUMMARY="$summary2" bash "$script"

out2="$(cat "$summary2")"
echo "--- case2 output ---"
echo "$out2"
echo "--------------------"

if ! grep -q "jsonl 無し" <<<"$out2"; then
  echo "FAIL: case2 出力に 'jsonl 無し' が無い" >&2
  exit 1
fi

# --- case 3: GITHUB_STEP_SUMMARY 未設定なら stdout へ ---
out3="$(SYSTEM_JSONL="$jsonl" bash "$script")"
echo "--- case3 output ---"
echo "$out3"
echo "--------------------"
if ! grep -q "pass 2" <<<"$out3"; then
  echo "FAIL: case3（STEP_SUMMARY 未設定）出力に 'pass 2' が無い" >&2
  exit 1
fi

# --- case 4: fail=0 / skip=0（実 e2e 1 本 PASS だけ）でも exit 0 ---
# why: set -o pipefail 下で grep 0 件が exit 1 になり、workflow summary step が赤になる事故を防ぐ。
jsonl4="${workdir}/pass_only.jsonl"
cat > "$jsonl4" <<'JSONL'
{"Action":"pass","Test":"TestProduceEpisodeSystem_runsEndToEndOnce_whenAllCredentialsPresent"}
{"Action":"pass","Package":"github.com/shim1103/daily-it-podcast/apps/generator/test/system"}
JSONL

summary4="${workdir}/step_summary_4"
: > "$summary4"
set +e
SYSTEM_JSONL="$jsonl4" GITHUB_STEP_SUMMARY="$summary4" bash "$script"
status4=$?
set -e
out4="$(cat "$summary4")"
echo "--- case4 output (exit ${status4}) ---"
echo "$out4"
echo "--------------------"
if [ "$status4" -ne 0 ]; then
  echo "FAIL: case4（fail=0）で script が exit ${status4}（want 0）" >&2
  exit 1
fi
if ! grep -q "pass 1" <<<"$out4"; then
  echo "FAIL: case4 出力に 'pass 1' が無い" >&2
  exit 1
fi
if ! grep -q "fail 0" <<<"$out4"; then
  echo "FAIL: case4 出力に 'fail 0' が無い" >&2
  exit 1
fi
if ! grep -q "skip 0" <<<"$out4"; then
  echo "FAIL: case4 出力に 'skip 0' が無い" >&2
  exit 1
fi

echo "PASS: system-summary.sh が fake jsonl の pass/fail/skip を集計した"
