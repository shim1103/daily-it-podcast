---
name: gdrive Narrow Integration Test 実装と責務分離
date: 2026-08-26T19:34:26
session_id: none
branch: feature/generator-narrow-gate-vendor-gdrive
prev: なし
---

## 1. Summary

`docs/tasks/todo/generator-narrow-gate-vendor-gdrive.md` の達成契約に従い、gdrive Writer の境界 I/O 観測を sociable unit test から Narrow Integration Test へ分離した。Unit は Adapter 内分岐のみを検証する Client Stub 直接実装へ置換し、二重検証を解消した。実装後の監査で Acceptance Criteria の「error message に dummy 値を含めない」が test 化されていないことが判明し、追加で該当 test case を実装させた。Issue file は完了として削除した。

## 2. Changes

1. Narrow Integration Test `apps/generator/test/gdrive_narrow_integration_test.go` を新設した。processenv 実物・httptest TLS server・DialTLSContext redirect で list→create→upload の成功 call sequence と json/wav stem 整合を self-validate する。
2. files.list 500 応答時に `error.Error()` が folder ID・token の実値を含まないことを assert する test case を追加した。
3. `writer_sociable_unit_test.go` から境界 I/O 観測（processenv・httptest・DialTLSContext）を全て除去し、`secrettransport.Client` を直接満たす境界 I/O なしの Stub へ置換した。
4. Integration gate（`scripts/generator/test-integration.sh`）と Unit test の両方が exit 0 であることを確認した。
5. Acceptance Criteria 全項目を manager が現物照合し、Issue file を削除した。
6. `develop` を base に PR #70 を作成した。CI（integration・static-and-unit）は全て pass、merge state は CLEAN。Copilot review は quota 制限で未実施だった。

### Commits

- `5def29b`
- `edf33c2`
- `7284ffa`
