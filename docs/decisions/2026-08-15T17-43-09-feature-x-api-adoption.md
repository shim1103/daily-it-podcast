---
name: X 取得の後回し（trends・Reply/Repost/引用・profile cache）
date: 2026-08-15T17:43:09
branch: feature/x-api-adoption
---

## 1. Decision

次は実装しない。必要になった session で改めて決める。

- trends（地域・カテゴリ問わず）
- 監視 user が他 user の post へ Reply した post
- 監視 user が他 user の post を Repost した post
- 監視 user が他 user の post を引用 Repost した post
- 監視 user の post へ他 user が Reply / Repost / 引用 Repost した post
- 人物 profile の取得と repo 内 cache（複数月に1回程度の更新。agent が人物を知る context 用。定数は人物 id のみ）

## 2. Reason

1query で完結するオリジナル投稿の一覧取得が現状の要件。上記は別 query・別責務・別更新周期が要り YAGNI。

## 3. Rejected

- 初回から Reply/Repost/引用・trends・profile 同期を Port に含める案（ISP / YAGNI 違反）
