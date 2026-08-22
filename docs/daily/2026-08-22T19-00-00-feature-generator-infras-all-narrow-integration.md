---
name: Generator credential runtime の境界契約と実装Issueを分離する
date: 2026-08-22T19:00:00
session_id: none
branch: feature/generator-infras-all-narrow-integration
prev: なし
---

## 1. Summary

Generator の credential runtime を、HTTP transport と command launch の境界へ分けた。A で contract を固定し、B で Composition の選択責務と production process environment runtime を decision として残し、C を 2 つの独立 Issue file へ分けた。

## 2. Changes

1. Composition が secret reference binding、Cursor project、child environment allowlist を所有する contract を追加した。
2. Cursor project identifier を AgentSecrets 側へ一元化し、既存 Cursor Adapter と Unit を同じ定数参照へ更新した。
3. process-env HTTP transport と process-env command launcher を別 Issue として lane へ登録した。
4. Decision に未実測の test fixture を書いた誤りを取り消し、Decision には決定済みの runtime boundary だけを残した。
5. Generator static check と Unit coverage gate を実行し、pass を確認した。

### Commits

- `ef979bb`
- `6c59867`
