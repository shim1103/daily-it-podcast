---
name: playback web API Clientの契約固定とtest分類記法の統一
date: 2026-08-20T14:18:45
session_id: none
branch: playback-web-api-client
prev: 2026-08-19T18-52-00-tmp-branch.md
---

## 1. Summary

`playback-web-api-client` のIssue draftを canonical source と照合して再計画し、決定した契約・定数・interfaceをcodeへ固定した。あわせてTS test fileの分類記法をGo側へ揃え、残る応答処理の実装をIssueへ切り出した。

## 2. Changes

1. session開始時、作業treeに残っていた差分が `origin/develop` と完全一致することを確認し、上流へ戻す逆差分と判定して破棄した
2. Issue draftの `listEpisodes(baseUrl)`・`ArrayBuffer | Uint8Array`・AC-4 の3点が canonical source と食い違うことを特定した
3. 型・写像表・factoryの実装を executor へ委譲し、manager として実物と独立検証で査読した
4. 査読で網羅性分岐の戻り値が宣言型を破ることを実測で確認し、TDDで修正した
5. shimの指摘を受け、契約codeの素通しを写像表による1対1変換へ変更した
6. 同じく指摘を受け、未知分類のfallbackを `unavailable` から throw へ変更した
7. 契約enumの拡張時に写像表の追加がcompileで強制されることを、契約へ仮のcodeを足して実測確認した
8. 網羅性検査の `never` 代入を外すと分岐欠落がcompileを通ることを実測し、代入を残す判断をした
9. 既存TS test file 16件を分類名付きへ改名し、runnerのunit収集条件も分類名で絞った
10. `DESIGN.md` §5の表が同節1の規約と矛盾していたため、TS記法を修正した
11. `playback-runtime-config-boundary.md` がClient実装へ踏み込んでいた範囲を、baseUrl注入元へ絞った
12. 応答処理の実装Issueを `create-issue` templateの11章構成で書き直した
13. Unit `107 passed`、typecheck、lint、formatを実行

### Commits

1. `9fa7f35` — test file分類記法の統一
2. `03d98b1` — API Clientの型・error写像・factory
