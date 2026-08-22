## 1. Summary

このIssueでは、generator の Boolean condition 実行を local report で確認できるようにする。既存 statement coverage gate は維持し、condition report は threshold を持たない補助指標とする。

## 2. Context

1. generator Unit は statement coverage 90% を gate にしている。
2. Go 標準 coverage は Boolean condition の true / false 実行を報告しない。

## 3. Canonical Sources

1. `docs/decisions/2026-08-22T17-55-00-chore-generator-ci-test-configuration-hardening.md` — coverage metric と実行場所の決定
2. `scripts/generator/test-unit.sh` — generator Unit coverage の既存入口
3. `2:platform/go/ci-checks.md` — Go coverage の能力境界

## 4. Scope

### In Scope

1. `gobco` `v1.3.4` による generator condition coverage report の local 入口
2. report を読むための local usage documentation

### Out of Scope

1. statement coverage 90% threshold の変更
2. hook、GitHub Actions、coverage threshold への condition report 追加
3. condition report の数値だけを満たす test 追加

## 5. Contract

local entrypoint は generator Unit package の condition coverage report を出力する。report が未達でも exit status を品質 gate に使わない。

## 6. Constraints

1. `gobco` は `v1.3.4` に固定する
2. report は statement coverage の置換ではない
3. target 外構文が report に現れないことを完全 coverage と解釈しない

## 7. Acceptance Criteria

1. [ ] AC-1: local entrypoint が generator condition coverage report を出力する
2. [ ] AC-2: statement coverage 90% gate の対象、閾値、exit status は変わらない
3. [ ] AC-3: hook と GitHub Actions は condition report を呼ばない
4. [ ] AC-4: report の対象外構文と hard gate 非採用を local usage documentation が示す

## 8. Verification

1. condition coverage local entrypoint が report を出力する
2. `./scripts/generator/test-unit.sh` が exit 0
3. hook と workflow が condition coverage entrypoint を参照しない

## 9. Dependencies

なし。

## 10. Risks

condition report の数値を完全 branch coverage と誤認すると test の品質を誤判定する。Decision Record の能力境界と usage documentation を確認する。

## 11. Notes

condition report で見つけた不足は、仕様の正常・異常・境界契約として test を追加できる場合だけ修正する。
