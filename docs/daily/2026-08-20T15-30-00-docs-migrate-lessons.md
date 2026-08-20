---
name: lessons候補71件のshared知識への昇格とscope-split skillの新設
date: 2026-08-20T15:30:00
session_id: none
branch: docs/migrate-lessons
prev: 2026-08-20T14-20-00-feature-playback-runtime-config-boundary.md
---

## 1. Summary

`docs/lessons/index.md` に溜まった知見候補71件を、既存skillへの追記・新規skill新設・削除へ振り分け、indexを空にした。あわせて `create-new-skill` のdraftを `scope-split` workflowとして確定した。

## 2. Changes

1. layer別に仕分け、削除判定（再発可能性・汎化可能性）と原因分解を適用して71件を分類した。適用可能26件、既記載24件、skill化困難13件、新規skill候補8件。
2. 既存13 fileへ追記した。terms系はshared-types、ports-adapters、defensive-design、cross-cutting、naming、naming-and-layout。meta系はdesign-philosophy §4-2とlogging/tasks。platform系はtypescript.md。
3. `2:platform` へ git、agentsecrets、sandbox-permissions の3 skillを新設し、親のlinksへ登録した。
4. Workers の buffer union は cloudflare 配下への新設を取りやめ、原因が TypeScript の型定義にあると判定して typescript.md へ収めた。
5. `3:workflow/scope-split` を新設した。A/B/Cの順序を「実装 → SSoT化してdocs → 残りをIssue」へ揃えるため、当初案のA（Decision/Docs）とB（Immediate Implementation）を入れ替えた。
6. `docs/tasks/todo/create-new-skill.md` を削除した。
7. dotfiles側で `shim check-contracts` を実行し、新設・編集した全fileで違反ゼロを確認した。
8. session外の変更（Go系build cacheのsandbox許可、issue-managerのbranch操作強化）を意図を推測して6 commitへ分割し、pushした。
