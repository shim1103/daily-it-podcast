#!/usr/bin/env bash
# name: generator-tts-rate-summary
# description: generator-tts-rate の実行ログから `PASS率 N/M` 行を拾い、
#              $GITHUB_STEP_SUMMARY（未設定なら stdout）へ書く。
# @require TTS_RATE_LOG が指す path（既定 /tmp/tts-rate.log）。無くても exit 0。
#          表示用に RUNS / CALL_GAP / RETRY_BACKOFF_BASE / RETRY_BACKOFF_MAX / PASS_THRESHOLD / DOUBLE を env で受ける（任意）。
# @ensure ログ内の最後の `PASS率 [0-9]+/[0-9]+` 行を summary へ書く。取れなければその旨を書く。ログが無ければ「ログ無し」。
# @invariant go test を YAML に直書きさせないための切り出し。副作用は summary への追記のみ。secret を出さない。
set -euo pipefail

log="${TTS_RATE_LOG:-/tmp/tts-rate.log}"

emit() {
  if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
    cat >> "${GITHUB_STEP_SUMMARY}"
  else
    cat
  fi
}

{
  echo "### Gemini TTS rate"
  echo ""
  echo "- double=${DOUBLE:-?} runs=${RUNS:-?} callGap=${CALL_GAP:-?} backoffBase=${RETRY_BACKOFF_BASE:-?} backoffMax=${RETRY_BACKOFF_MAX:-?} threshold=${PASS_THRESHOLD:-?}"
  if [ -f "$log" ]; then
    line="$(grep -E 'PASS率 [0-9]+/[0-9]+' "$log" | tail -1 || true)"
    if [ -n "${line}" ]; then
      echo "- ${line}"
    else
      echo "- PASS 率行が取れなかった（Skip または早期失敗）"
    fi
  else
    echo "- ログ無し"
  fi
} | emit
