## 1. Summary

このIssueでは、generator の boundary logic に local-only mutation testing を導入する。survived mutant を assertion不足の候補として確認できる状態にし、CI score gate は作らない。

## 2. Context

1. generator は Go `1.26.6` を使う。
2. mutation testing は coverage と異なり、test assertion が小さな implementation change を検出するかを補助的に確認する。

## 3. Canonical Sources

1. `docs/decisions/2026-08-22T17-56-00-chore-generator-ci-test-configuration-hardening.md` — mutation tool、実行場所、survived mutant の扱い
2. `2:platform/go/ci-checks.md` — Go test check の能力境界
3. `testing-strategy/coverage.md` — mutation testing の位置付け

## 4. Scope

### In Scope

1. `mutest` `v0.6.0` を使う generator local mutation entrypoint
2. 原稿 validation を含む Application boundary logic の mutation report
3. survived mutant が仕様 bug を示す場合だけの Unit Test 修正

### Out of Scope

1. hook、GitHub Actions、mutation score threshold
2. 全 generator package の mutation 実行
3. survived mutant ごとの機械的な test 追加

## 5. Contract

local entrypoint は対象 boundary logic の mutation result を出力する。result は review input であり、score の exit status を CI quality gate に使わない。

## 6. Constraints

1. `mutest` は `v0.6.0` に固定する
2. Unit Test と同じく credential と external service を使わない
3. equivalent mutant は test の追加で固定しない

## 7. Acceptance Criteria

1. [ ] AC-1: local entrypoint が対象 package の mutation result を出力する
2. [ ] AC-2: hook と GitHub Actions は mutation entrypoint を呼ばない
3. [ ] AC-3: survived mutant ごとに仕様 bug、equivalent mutant、tool limitation のいずれかを区別できる
4. [ ] AC-4: 仕様 bug の survived mutant は、その仕様を検証する Unit Test で kill される

## 8. Verification

1. mutation local entrypoint が対象 package の result を出力する
2. generator Unit gate が exit 0
3. hook と workflow が mutation entrypoint を参照しない

## 9. Dependencies

なし。

## 10. Risks

mutation tool が timeout または equivalent mutant を出す可能性がある。result を score だけで判定せず、分類してから test を変える。

## 11. Notes

comparison / equality 以外の mutation が必要になった場合は、このIssueを拡張せず別の Decision Record を作る。
