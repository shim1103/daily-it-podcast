#!/usr/bin/env bash
# name: generator-test-r2-smoke
# description: R2（S3 互換 API）へ test bucket の Put/List/Get 往復ができるかを確かめる
#              dispatch 専用 smoke test を実行する。stem pair 整合・schema 適合は測らない。
# @require リポジトリ内から呼ぶ。Go が PATH にある。apps/generator が存在する。
# @require TEST_R2_ACCOUNT_ID / TEST_R2_BUCKET / TEST_R2_ACCESS_KEY_ID / TEST_R2_SECRET_ACCESS_KEY が
#          env に渡っている（無ければ test 側で Skip = 環境要因、smoke 対象外）。
# @ensure `r2smoke` tag の TestR2Smoke だけ実行する。他 test を巻き込まない。
# @invariant Unit / Integration / System gate を呼ばない。cron を持たない。secret 値を log に出さない。
#            probe object は test 側の t.Cleanup が dispatch 末尾で自前 delete する。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"

echo "r2-smoke: generator (go)"
(
  cd "$root/apps/generator"
  go test -v -tags=r2smoke -run TestR2Smoke -timeout 10m -count=1 ./test/system/
)
