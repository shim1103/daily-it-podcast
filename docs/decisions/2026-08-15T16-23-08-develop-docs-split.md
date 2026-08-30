---
name: 文書は README・DESIGN・contracts の3分業
date: 2026-08-15T16:23:08
branch: develop
---

## 1. Decision

README は地図・使い方・受け入れ・秘密の名前のみ。DESIGN は層・依存・所有・test 配置の規則のみ（skill へ委譲）。Drive の表現は contracts。PROPOSAL / SPEC / dir 単位 README は置かない。

## 2. Reason

shim 主導のため PROPOSAL の受け入れ条件文書は不要。小規模 repo では dir README 乱立が腐敗しやすい。規則は薄く、詳細は architecture / testing-strategy skill を正とする。

## 3. Rejected

- SPEC を常設する案（使い方が肥大したら分離でよい）
- DESIGN に Drive 契約やパス百科を併記する案（DRY 崩壊）
