---
name: 文書整理 PR に git history 改変を含めない
date: 2026-09-01T12:58:00
branch: cursor-docs-ssot-develop-cleanup-9567
---

## 1. Decision

obsolete 文書の削除や README / DEPLOY / lane の現状合わせを `develop` へ反映するとき、**filter-branch 等の git history 改変は同梱しない**。file の追加・削除・更新は通常 merge で足りる。

## 2. Reason

1. 先行の文書 3 分業 Decision（`2026-08-15T16-23-08`）と mock 常設却下（`2026-08-25T07-01-00` §1-3）は、**置かない方針**を既に固定している。残存していた `docs/SPEC` 系は方針の物理実施であり、履歴を書き換えなくても merge で除去できる。
2. history 改変を同梱すると、reviewer は「文書が正しいか」と「履歴が壊れないか」の二重監査になり、却下理由が文書内容と無関係に膨らむ。PR #108 はこの同梱が拒否された。
3. 旧実装は `archive/2026-08-15-pre-rewrite` tag で git 上に残る。README から archive 行を除いても、履歴改変なしで現行地図の SSOT を読める。

## 3. Rejected

1. pre-rewrite 根 commit を filter-branch で除去する案 — file 削除と独立した高リスク操作。merge だけで達成できる目的に対し force push を別途要する。
2. 「履歴が汚いから整理 PR と一緒に直す」案 — 文書正本化と履歴整容は変更理由が違う。同 PR に混ぜると却下時に文書修正まで巻き戻る。
3. archive tag ごと消す案 — 凍結記録を失い、rewrite 以前の参照不能になる。今回の目的（現行文書の SSOT 化）に不要。
