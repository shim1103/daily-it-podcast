#!/usr/bin/env bash
# name: playback-check-static
# description: playback の静的検査（Biome + tsc --noEmit + dependency-cruiser + wrangler types 一致）を実行する。
# @require リポジトリ内から呼ぶ。playback は npm 依存が install 済み。
# @ensure package.json が無ければ成功。あれば Biome format/lint・tsc --noEmit・層 import 検査・
#   wrangler.jsonc から再生成した worker-configuration.d.ts が commit 済み内容と一致することが
#   exit 0 のときだけ成功する。
# @invariant Unit / Integration を実行しない。generator を触らない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

(
  cd "$root/apps/playback"
  if [[ ! -f package.json ]]; then
    echo "skip: package.json なし（空 package）"
    exit 0
  fi
  echo "format: playback (biome)"
  npm run format:check
  echo "lint: playback (biome)"
  npm run lint
  echo "typecheck: playback (tsc)"
  npm run typecheck
  echo "layers: playback (dependency-cruiser)"
  npm run lint:layers
  echo "wrangler types: playback (worker-configuration.d.ts 再生成一致)"
  npm run types:worker
  git diff --exit-code -- worker-configuration.d.ts || {
    echo "worker-configuration.d.ts が wrangler.jsonc と不一致です。'npm run types:worker' を実行し直して commit してください" >&2
    exit 1
  }
)
