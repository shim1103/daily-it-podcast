---
name: generator Application boundary logicへlocal専用mutation testingを導入
date: 2026-08-22T19:30:00
session_id: none
branch: test/generator-mutation-testing
prev: なし
---

## 1. Summary

`docs/decisions/2026-08-22T17-56-00-chore-generator-ci-test-configuration-hardening.md` の決定に従い、generator の `internal/application` に対して `mutest` v0.6.0 による local専用 mutation testing entrypoint を追加した。

## 2. Changes

1. `mutest` の module path は検索で `github.com/fchimpan/mutest` と特定し、sandbox解除の上 `go install github.com/fchimpan/mutest@v0.6.0` した。
2. `internal/application/write_episode.go` の比較・等価演算子 4 箇所（episodeID 空判定、audio 空判定、episodeID 不一致判定、trailing JSON 判定）を mutation 対象とした。`internal/application/fetch_source_items.go` は対象演算子なしのため対象外、`internal/entities/errors/*` は Application boundary logic ではないため対象外と判断した。
3. `scripts/generator/mutation-local.sh` を実行し、4 mutant すべてが既存 Sociable Unit Test で kill されたことを確認した（Survived 0、Score 100.0%）。
4. `code-review` は findings 0 件。`simplify` の Altitude 観点で「survived 時に非 0 exit するのは CI gate 契約と不整合では」という指摘が出たが、`mutest` の source（`run.go` の `Summary` 出力が `os.Exit` 前に完了する実装）を読んで反証し、修正不要と判定した。
5. hook（`scripts/git-hooks/*`）と GitHub Actions（`.github/workflows/*`）が mutation entrypoint を参照しないことを確認した。

### Commits

1. `ee0cfa3`
2. `43c7963`
