#!/usr/bin/env bash
# name: generator-produce-episode
# description: Generator 本番 CLI（cmd/generator）を実行する。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @require process environment に config 契約の runtime config が揃っている（本 script は secret を読まない・作らない）。
# @ensure go run ./cmd/generator の exit code をそのまま返す。
# @invariant playback を触らない。Integration / System test 入口を呼ばない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "produce: generator"
(
  cd "$root/apps/generator"
  go run ./cmd/generator
)
