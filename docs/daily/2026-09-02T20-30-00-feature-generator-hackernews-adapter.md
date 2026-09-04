---
name: HackerNews ItemSource Adapter の List を実装し shim レビュー3ラウンドで取得ロジック・GWT・doc を是正する
date: 2026-09-02T20:30:00
session_id: none
branch: feature-generator-hackernews-adapter
prev: 2026-09-02T15-40-00-feature-hackernews-api-adapter.md
---

## 1. Summary

scope-split A で stub のまま残っていた `infrastructure/hackernews` の `ListItemSource.List` を issue-manager flow（manager audit + executor/reviewer 委譲）で TDD 実装した。`topstories.json` → 個別 item/comment 取得 → `SourceItem` 変換を通し、Sociable Unit と Narrow Integration を緑にした。その後 shim レビューを3ラウンド受け、(1) issue stub が godoc へ書いた「振る舞い（B/C が実装する）」ブロックの削除、(2) 取得ロジックのバグ（先頭 N 件だけを見てその中の window 内を残す形になっており「1日以内があれば N 件取る」意図を満たさない）の修正、(3) Sociable Unit / Narrow の GWT 構造違反（When 直後の早期 Fatal、@then コメントの散文化）の是正を行った。`MaxStoriesScanned` の意味を「走査 id 数の上限」から「時間窓内 story を最大この件数まで結果に含める上限」へ再定義し、同型の lobsters/itmedia task file へ申し送りを残した。pr-completion で 2 commit へ分割し PR #112 を `gh pr` で base `develop` へ作成した。

## 2. Changes

1. issue-manager flow: manager が read-only で plan、executor が TDD 実装 + self-review、reviewer が code-review + /simplify、executor が reviewer 指摘（should-fix 4件: retry 破棄コメント / Replacer 巻き上げ / io.ReadAll err 経路の SU / Narrow の retry 回数 assert）を再対応、manager が AC §7 全 12 項目の緑を audit、issue file 削除。
2. shim レビュー R1: issue stub（scope-split A）が `item_source.go` の `List` godoc へ書いた「振る舞い（B/C が実装する）」箇条書きは実装完了後に削除すべき（実装コードと sociable unit test が SSOT）。hackernews で削除し、lobsters/itmedia task file の §4 に同じ削除を明記。`MaxStoriesScanned` const コメントを「走査上限であり結果件数上限ではない」旨へ一旦明確化。
3. shim レビュー R2 バグ: 旧実装は `topstories.json` の先頭 `MaxStoriesScanned` 件だけ fetch しその中から `time >= since` を残していた。topstories はランキング順（時刻順でない）ため、1日以内の story が上位 N 件の外にあると取りこぼす。実装を「id 列を先頭から走査し `time >= since` の story を集め、`MaxStoriesScanned` 件に達したら break」へ変更。RED（先頭30件が全部 window 外でも後ろの window 内 story を拾う test 等）2本を追加、既存 1本をリネーム。`MaxStoriesScanned` の意味を「結果件数の上限」へ再定義し const コメント・godoc `@ensure`・lobsters 申し送りを整合。
4. shim レビュー R2 回答: `item_source.go` の `net/http` / `io` import は標準ライブラリ型・関数の参照であり DI 対象ではない。`*http.Client` は既に `ListItemSource.client` フィールドとして composition から注入済み。`os` はプロセスのグローバル状態アクセスなので他 Adapter は関数注入で隔離しているが、この Adapter は `os` を import せず `time.Now()` も呼ばない（`since` は引数）。設計どおりで修正不要。
5. shim レビュー R2 GWT（RTFM）: `item_source_sociable_unit_test.go` 20 関数（サブテスト含む）と `hackernews_narrow_integration_test.go` 2 関数で、`// @when` が説明句付き（gwt.md 規則3）、`// @then` が期待結果の散文（規則13/14）だった。When ブロックを対象呼び出し1行へ、Then コメントを構造ラベル（「戻り値と error」等）へ全件是正。
6. 検証: generator は build/vet/gofmt/golangci-lint(depguard) 0 issues、hackernews package SU 20/20、Narrow 2/2、`go test ./...` 全緑（broad integration 3 SKIP は別 task 担当）、generator unit coverage 91.9%。commit 前に playback toolchain（node@22 を PATH 先頭へ置いて npm install）を両系統 install し、pre-commit / pre-push hook を実際に緑で通した。
7. `npm install` が worktree の dir 名を拾って `package-lock.json` の `name` を書き換えた副産物を手で HEAD 値へ戻し、commit 対象から除外。
8. pr-completion: `commit --repo --split`（2 commit）→ `log-session` → `create-pr`。log-session commit（`be33691`）とその push は playback toolchain（node@22 を PATH 先頭へ置いて npm install）を両系統 install したうえで pre-commit / pre-push hook を実際に緑で通した。PR は素の `gh pr create`（`shim gh` wrapper は使わない）で base `develop` へ作成。#112。MERGEABLE、CI（static-and-unit / integration）実行中。AgentReview check は無し。GitHub Issue は無く（scope-split C は task file 起票のみ）、PR は関連 Issue なしで作成。

### Commits

- `9d0a01a` feat(generator): HackerNews ItemSource Adapter の List を実装する
- `5b7f77d` docs(generator): hackernews task を完了とし lobsters/itmedia へ申し送りを残す
- `be33691` docs(log): セッションログ

### PR

- #112 → base `develop`（README の branch 方針）。https://github.com/shim1103/daily-it-podcast/pull/112
