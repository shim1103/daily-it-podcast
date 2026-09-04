---
name: ITmedia NEWS ItemSource Adapter の List を実装し pr-completion で PR 作成
date: 2026-09-02T22:35:00
session_id: none
branch: feature-generator-itmedia-adapter
prev: 2026-09-02T20-30-00-feature-generator-hackernews-adapter.md
---

## 1. Summary

scope-split A の stub `infrastructure/itmedia` の `ListItemSource.List` を issue-manager flow（manager audit + executor/reviewer 委譲）で実装した。ITmedia NEWS 速報 RSS 2.0 の GET → `encoding/xml` parse → `pubDate >= since` filter → `SourceItem` 変換を通し、Sociable Unit 8 本（サブテスト含む）と Narrow Integration 2 本を緑にした。pr-completion 前に Narrow の GWT comment を Go 慣習の `@given` / `@when` / `@then` へ是正した。GitHub Issue は無く task file 起票のみ（hackernews と同型）。

## 2. Changes

1. issue-manager flow: executor が `item_source.go` List 実装・SU 6 本 + reviewer 指摘（+0900 test / 4xx no-retry / stub 整理 / Op assert / retry 成功 test）・Narrow 2 本。manager が AC 照合後 issue file 削除。
2. shim 指摘: Go test の GWT tag `@given` は Go 慣習であり project 横断規約ではない。TS / shell は `Given:` のまま。repo 全体統一 task（`test-gwt-comment-tags.md`）は Go `*_test.go` のみ scope に修正したが、本 PR には含めない（別 work）。
3. Narrow Integration の GWT を Sociable Unit と同型（`@when` は List 呼び出し 1 行のみ）へ修正。
4. 検証: `go build` / `go vet` / itmedia SU + Narrow / `golangci-lint` 0 issues。

### Commits

- `d4f556b` feat(generator): ITmedia NEWS ItemSource Adapter の List を実装する
- `018489c` docs(generator): itmedia adapter task を完了する
- `edc0219` docs(log): セッションログ

### PR

- （作成後に追記）
