---
name: generator の静的入口へ lint / formatter を追加する
date: 2026-08-18T18:40:00
session_id: none
branch: chore/generator-static-lint-format
prev: なし
---

## 1. Summary

`apps/generator` の静的入口が depguard のみだった状態に、curated subset（errcheck・govet）と formatter（gofmt、check のみ）を追加した。既存 code の指摘 6 件（`Body.Close` 未処理 3 件、gofmt 崩れ 3 件）はローカル修正のみで解消した。`check-static.sh` / `test-unit.sh` の Acceptance Criteria は両方 exit 0 を確認した。

## 2. Changes

1. `apps/generator/.golangci.yml` に `errcheck` / `govet` を `linters.enable` へ、`gofmt` を `formatters.enable` へ追加した。既存 `depguard` の rules は無変更
2. `scripts/generator/check-static.sh` の description / echo 文言を実態に合わせて更新した（実行ロジックは無変更）
3. `res.Body.Close()` の error 未処理 3 件を `defer func() { _ = res.Body.Close() }()` へ修正した
4. gofmt 崩れ 3 件（struct field / const alignment、doc comment 改行）を整形した
5. `docs/tasks/todo/pr-b-generator-static-lint-format.md` を完了として削除した

### Commits

- `f042864` chore(generator): golangci-lintへerrcheck/govet/gofmtを追加する
- `9014bc6` docs(tasks): generator静的lint/format issueを完了として削除する
