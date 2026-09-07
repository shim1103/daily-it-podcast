#!/usr/bin/env bash
# name: generator-geminiapi-smoke-summary
# description: generator-geminiapi-smoke の実行ログから疎通結果（`疎通 OK` 行 / Skip / FAIL）を拾い、
#              $GITHUB_STEP_SUMMARY（未設定なら stdout）へ書く。
# @require GEMINIAPI_SMOKE_LOG が指す path（既定 /tmp/geminiapi-smoke.log）。無くても exit 0。
# @ensure ログ内の `geminiapi 疎通 OK` 行、または `--- SKIP` / `--- FAIL` の該当行を summary へ書く。
#         取れなければその旨を書く。ログが無ければ「ログ無し」。
# @invariant go test を YAML に直書きさせないための切り出し。副作用は summary への追記のみ。secret を出さない。
set -euo pipefail

log="${GEMINIAPI_SMOKE_LOG:-/tmp/geminiapi-smoke.log}"

emit() {
  if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
    cat >> "${GITHUB_STEP_SUMMARY}"
  else
    cat
  fi
}

{
  echo "### geminiapi 実 API 疎通 smoke"
  echo ""
  if [ -f "$log" ]; then
    ok="$(grep -E 'geminiapi 疎通 OK' "$log" | tail -1 || true)"
    if [ -n "${ok}" ]; then
      echo "- ✅ ${ok}"
    elif grep -qE '^\s*--- SKIP' "$log"; then
      echo "- ⏭️ SKIP（TEST_GEMINI_API_KEY 未設定）"
      grep -E 'smoke precondition' "$log" || true
    elif grep -qE '^\s*--- FAIL|^FAIL' "$log"; then
      echo "- ❌ FAIL"
      grep -E 'Write\(\) error|空断片' "$log" || true
    else
      echo "- 結果行が取れなかった（早期失敗の可能性）"
    fi
  else
    echo "- ログ無し"
  fi
} | emit
