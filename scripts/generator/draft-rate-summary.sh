#!/usr/bin/env bash
# name: generator-draft-rate-summary
# description: generator-draft-rate の実行ログから `draft PASS率` 行と `run N/M:` 行を拾い、
#              $GITHUB_STEP_SUMMARY（未設定なら stdout）へ書く。
# @require DRAFT_RATE_LOG が指す path（既定 /tmp/draft-rate.log）。無くても exit 0。
#          表示用に RUNS / PASS_THRESHOLD / PROMPT_VARIANT を env で受ける（任意）。
# @ensure ログ内の最後の `draft PASS率 [0-9]+/[0-9]+` 行と、全 `run [0-9]+/[0-9]+:` 行を summary へ書く。
#         取れなければその旨を書く。ログが無ければ「ログ無し」。
# @invariant go test を YAML に直書きさせないための切り出し。副作用は summary への追記のみ。secret を出さない。
set -euo pipefail

log="${DRAFT_RATE_LOG:-/tmp/draft-rate.log}"

emit() {
  if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
    cat >> "${GITHUB_STEP_SUMMARY}"
  else
    cat
  fi
}

{
  echo "### CursorCLI draft rate"
  echo ""
  echo "- runs=${RUNS:-?} threshold=${PASS_THRESHOLD:-?} variant=${PROMPT_VARIANT:-?}"
  if [ -f "$log" ]; then
    summary="$(grep -E 'draft PASS率 [0-9]+/[0-9]+' "$log" | tail -1 || true)"
    if [ -n "${summary}" ]; then
      echo "- ${summary}"
    else
      echo "- PASS 率行が取れなかった（Skip または早期失敗）"
    fi
    echo ""
    echo "各回:"
    grep -E 'run [0-9]+/[0-9]+:' "$log" || echo "（run 行なし）"
  else
    echo "- ログ無し"
  fi
} | emit
