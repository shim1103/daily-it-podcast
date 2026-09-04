#!/usr/bin/env bash
# name: generator-system-summary
# description: generator-system の -json 出力（jsonl）から top-level test の pass/fail/skip を数え、
#              $GITHUB_STEP_SUMMARY（未設定なら stdout）へ PASS 率を書く。
# @require SYSTEM_JSONL が指す path（既定 /tmp/generator-system.jsonl）。無くても exit 0。
# @ensure top-level test（Test 名に "/" を含まない = サブテストでない）の pass/fail/skip を数えて summary へ書く。fail=0 でも exit 0。
# @ensure fail>0 のとき FAIL した top-level test 名一覧も書く。jsonl が無ければ「jsonl 無し」行を書く。
# @invariant go test を YAML に直書きさせないための切り出し。副作用は summary への追記のみ。secret を出さない。
set -euo pipefail

jsonl="${SYSTEM_JSONL:-/tmp/generator-system.jsonl}"

emit() {
  if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
    cat >> "${GITHUB_STEP_SUMMARY}"
  else
    cat
  fi
}

{
  echo "### Generator System"
  echo ""
  if [ -f "$jsonl" ]; then
    # why: top-level test（=/= を含まない Test 名）の pass/fail/skip を数える。
    # why: set -o pipefail 下で grep 0 件が exit 1 になるので、|| true で 0 件を許す。
    pass=$(grep -E '"Action":"pass"' "$jsonl" | grep -E '"Test":"[^"/]+"' | wc -l | tr -d ' ') || true
    fail=$(grep -E '"Action":"fail"' "$jsonl" | grep -E '"Test":"[^"/]+"' | wc -l | tr -d ' ') || true
    skip=$(grep -E '"Action":"skip"' "$jsonl" | grep -E '"Test":"[^"/]+"' | wc -l | tr -d ' ') || true
    pass=${pass:-0}
    fail=${fail:-0}
    skip=${skip:-0}
    total=$((pass + fail + skip))
    echo "- top-level test: pass ${pass} / fail ${fail} / skip ${skip}（total ${total}）"
    if [ "${fail}" -gt 0 ]; then
      echo ""
      echo "FAIL:"
      grep -E '"Action":"fail"' "$jsonl" | grep -oE '"Test":"[^"/]+"' | sort -u || true
    fi
  else
    echo "- jsonl 無し（System package なし、または早期失敗）"
  fi
} | emit
