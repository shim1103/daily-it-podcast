---
name: X 取得 Adapter は vendor ごとに置き、x/ は切替 facade にしない
date: 2026-08-17T17:15:01
branch: feature/x-getxapi-adapter
---

## 1. Decision

GetXAPI と TwitterAPI.io の `PostSource` 実装は共通 package にまとめない。

`infrastructure/x/` は外部系 X の Driven Adapter 置き場であり、vendor 差分を吸収する facade ではない。統一 interface の所有者は Application の `PostSource`。vendor 切替の唯一の場所は Composition Root。

次の X vendor の仕様が同じだと予測しない。同じ知識が三度現れてから契約変換と判定の分離を見直す。

## 2. Reason

見た目の page loop は似ていても、endpoint・auth・page 名・media・error 形が違う。共通化は「X vendor はみな同型」という未確認予測を固定する。Port の `@ensure`（オリジナルのみ・`since` 以上）が既に共有判定の正。Adapter は vendor JSON をその契約へ落とす mechanism だけを持つ。`x/` が両 vendor を import して中で選ぶと Composition の結線特権を奪い、1 module の変更理由が二つになる。

## 3. Rejected

- pager / tweet 正規化の共通 package（偽の共通概念、YAGNI）
- `infrastructure/getxapi` と `infrastructure/twitterapiio` への平坦化（仕組みと会社名が直下で混ざる）
- `infrastructure/x` が両 Adapter を選ぶ facade（Composition の仕事の移動、多層 fallback に近い）
