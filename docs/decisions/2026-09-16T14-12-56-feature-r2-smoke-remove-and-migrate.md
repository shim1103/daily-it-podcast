---
name: 列6（r2-smoke-migrate-cutover）を完了削除し残タスクを列7（r2-post-cutover-verify-oauth）へ統合する
date: 2026-09-16T14:12:56
branch: feature/r2-smoke-remove-and-migrate
---

## 1. Decision

先行 Decision `2026-09-14T12-49-26`（Issue 分割・列 6/7 を別 Issue とする）を、着手順（mp3 → R2 → cache）自体は変更せず、**Issue 分割の単位だけ**部分的に supersede する。

列6（`r2-smoke-migrate-cutover`）の Scope①②③（結線切替・疎通確認）は develop へ merge 済み・PASS 済みで完了したため、Issue file を削除する。残タスク（prod/test bucket への人手移行済みデータの確認、本番同着 deploy、`DEPLOY.md` 更新）は、列7（`r2-post-cutover-verify-oauth`）へ統合する。以後、旧列6・7の残る作業は 1 Issue（`r2-post-cutover-verify-oauth`）が持つ。

## 2. Reason

1. 列6のScope①②③（generator/playback結線の完全置換、疎通dispatchのPASS確認）は既に検証可能な状態で完了しており、Issue としての未完了達成契約が残っていない。完了したIssueをrelease専用の空Issueとして残すと`logging`規則（「Issueはchangelogではない」）に反する。
2. 残るタスク（人手移行データの確認・本番同着deploy・DEPLOY.md更新）は、列7が元々持つ達成契約（System/E2E緑化・OAuth削除）と同じ「切替が本当に完了したか」を検証する文脈を共有する。deploy・確認・System/E2E緑化は一連の検証作業であり、Issue単位を分けると「列6完了」を宣言する基準と「列7着手可能」の基準が同じ事実（deploy済みR2が正本）を指すのに2つのIssueへ分散する。
3. 先行Decision（`12-49-26`）が列6・7を分けた理由（疎通とSystem/E2Eの責務分離、OAuth削除を切替検証後に送る）は変わらず有効——分けたのは「dispatch専用の疎通」と「正式gateのSystem/E2E」という**検証手段の違い**であり、その責務分離はIssue本文内のScope区分（本Issueの旧Scope①〜④と⑤〜⑧）で表現すれば足りる。Issue file自体を2つに割る必要は無かった。

## 3. Rejected

1. **列6を空のまま残し「完了」ラベルだけ付ける案** — 達成契約が空になったIssueをrepoに残す意味が無い。削除して統合する方が進捗indexの読み手にとって単純（KISS）。
2. **列7を新規に作り直す案** — 既存の列7 Contract/Constraints（OAuth削除は不可逆、System/E2E緑化が前提）は今回の統合でも有効なため、既存fileへ追記する方がSSOTの一貫性を保つ。
