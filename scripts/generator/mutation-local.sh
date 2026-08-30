#!/usr/bin/env bash
# name: generator-mutation-local
# description: generator の Application boundary logic に対して mutest による local mutation testing を実行する。
# @require リポジトリ内から呼ぶ。mutest v0.6.0 が PATH にある。apps/generator が存在する。
# @ensure internal/application package の mutation result（killed / survived）を標準出力へ出す。exit status は review 判断の入力であり、CI quality gate には使わない。
# @invariant hook と GitHub Actions から呼ばない。internal/application 以外の package を対象にしない。playback を触らない。本番 credential を読まない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "mutation (local only): generator internal/application"
(
  cd "$root/apps/generator"
  mutest -v ./internal/application/...
)
