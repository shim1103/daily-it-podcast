---
name: playback Unit coverage gate の導入と分岐 pass 化
date: 2026-08-25T23:28:28
session_id: none
branch: chore/playback-unit-coverage
prev: なし
---

## 1. Summary

`docs/tasks/todo/playback-lane.md` の Issue 化待ちだった PR-H を進め、playback の Vitest Unit に branch coverage gate（`@vitest/coverage-v8`）を導入し、閾値未達だった分岐を unit test 追加と到達不能分岐の `v8 ignore` 化で解消した。設計判断は Decision Record（同日時刻）に固定し、gate 定義を `DESIGN.md` へ反映した。

## 2. Changes

1. `@vitest/coverage-v8` を導入する過程で、`test.projects` 配下の個別 project へ coverage 設定を書いても無視される（root top-level `test.coverage` でしか有効にならない）Vitest の仕様上の制約を発見した。最初の設定では threshold 判定自体が起きておらず `exit=0` が誤った成功だったことを、ERROR ログの不在から気づいて設定を root へ移した。
2. `executor` agent へ rename・test 追加・pass 化を委譲し、完了報告を受けた後、報告にあった2箇所の到達不能分岐（`map-internal-error.ts` の exhaustiveness check、`match-playback-route.ts` の `?? ""`）を自分で検証し、`tsc --noEmit` で型上の到達可能性を確認した上で対応（削除／`v8 ignore`）した。
3. threshold を pass させる過程で、global threshold（100%）が個別 glob threshold（90%）該当 file も合算した全体値で判定される Vitest の仕様（vitest-dev/vitest issue #6165）に遭遇し、個別 glob 対象 file（`google-drive-episode-repository.ts` 等）を unit test 追加で実質 100% まで引き上げることで解消した。
4. 全 commit で pre-commit hook（static + generator/playback unit + coverage）・pre-push hook（integration）が実行され、いずれも pass した。

### Commits

1. `478052b`
2. `3fba50f`
3. `751ca3d`
