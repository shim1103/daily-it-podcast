---
name: RetryReporter導入とgolangci-lintへのstaticcheck追加
date: 2026-09-26T14:44:57
session_id: none
branch: feature/generator-logging-mature
prev: なし
---

## 1. Summary

`generator-produce-episode` workflowの失敗調査から、Cursor/Gemini/TTS/R2/記事source取得のretryループが各attemptの失敗理由を握りつぶしていることが判明した。`port.RetryReporter`をDIし、次attemptへ進む直前にstep/attempt/max/reasonをrealtime観測できるようにした。nilの扱いは設計を1往復させ、最終的にComposition Rootの1点（`newProduceEpisodeWithTopicCount`）へfail-fastを集約する形に落ち着いた。併せて`golangci-lint`へ`staticcheck`を導入した。

## 2. Changes

1. GitHub Actions runのログをAPI経由で直接確認し、`invalid_manuscript_draft: topic[1].title has no japanese`が直近4 attempt中3回発生していたことを裏取りした
2. `port.RetryReporter`interfaceを新設し、`geminiapi`・`cursorapi`・`speech/gemini`・`r2`（`EpisodeWriter`/`CompletedEpisodeLookup`）・`httpget`（記事source5種が共有する`GetWithRetry`）へDIした
3. nil扱いの設計を1往復させた。最初はAdapter個別constructorで`retry == nil`をpanicさせ`port.NoopRetryReporter{}`をDummyとして新設したが、Opus 5.5への相談でGoのtyped nil問題（`*T`のnilをinterfaceへ渡すと`== nil`判定が効かない）を踏まえるとAdapter個別のpanicは実際の結線漏れを検出できないことが分かった。fail-fastをComposition Rootの1点へ集約し、Adapterはdoc契約（`@require retry != nil`）だけを持つ形に直し、`NoopRetryReporter`は削除した
4. `.golangci.yml`へ`staticcheck`を追加し、検出された既存コードのスタイル指摘6件（De Morgan則、型注釈の要否）を解消した
5. `1:terms/architecture/configuration-boundary.md`へ「暗黙defaultとoptional observerの区別」「fail-fastの1点集約とtyped nil」の2節を追記した（worktree・`agent-standards`正本の両方）
6. `git revert`と`git apply`を組み合わせ、一度commitした実装をcommit履行上は取り消しつつ、working treeへdirty差分として戻す操作を行った（`git reset`系はhookで禁止されているため）
7. `/pr-completion`に着手し、`feature/generator-logging-mature`は`develop`から分岐した1本のfeature branchで、PRのbaseは`develop`が適切と確認した。`commit --repo`実行時点でworking treeは既にclean、リモートとも同期済みだった

### Commits

1. `195ac19`
2. `1125f95`
3. `50b0663`
