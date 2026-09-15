#!/usr/bin/env bash
# name: playback-test-local-r2-binding
# description: getPlatformProxy による local R2 binding infra の Verification を実行する。
# @require リポジトリ内から呼ぶ。apps/playback の npm 依存（wrangler）がある。
# @ensure createLocalR2Binding の Narrow Integration が exit 0。
# @invariant test-unit に混ぜない。本番 remote binding を使わない。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "integration: playback local R2 binding (getPlatformProxy)"
(
  cd "$root/apps/playback"
  npx vitest run --project integration test/integration/local_r2_binding.narrow_integration.test.ts
)
