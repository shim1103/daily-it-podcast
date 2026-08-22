#!/usr/bin/env bash
# name: generator-fuzz-pcm-to-wav
# description: pcmToWAV の Go stdlib fuzz target を bounded local 実行する。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @ensure -fuzztime 内で fuzzing が完了し、failure なしのときだけ exit 0。
# @invariant hook / PR CI / scheduled CI から呼ばれない。fuzztime は無制限にしない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
fuzztime="${GENERATOR_FUZZ_PCM_TO_WAV_TIME:-20s}"

echo "fuzz: generator pcmToWAV (fuzztime=$fuzztime)"
(
  cd "$root/apps/generator"
  go test -run='^$' -fuzz='^FuzzPCMToWAV$' "-fuzztime=$fuzztime" ./internal/infrastructure/speech/gemini/
)
