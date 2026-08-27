---
name: production sourceをGetXAPIへ一本化しTwitterAPI.ioの現行artifactを除く
date: 2026-08-27T13:56:15+09:00
branch: docs/env-secret-management-reconsider
---

## 1. Decision

GetXAPIをGenerator唯一のproduction sourceとする。未使用のTwitterAPI.io実装、test、Composition、保留task、latest runtime inventoryは現行artifactから除く。一方、過去のDecision Recordとdaily recordは変更も削除もせず、その時点の履歴として残す。

## 2. Reason

production CompositionがGetXAPIだけを選ぶ状態でTwitterAPI.io実装を残すと、保守対象のsourceが二つあるように見える。使われないAdapter、test、credential名、Compositionを更新し続けてもproduction経路の信頼性は上がらず、source契約を変更するたびに無効な選択肢まで確認する負担が増える。

保留taskとruntime inventoryは現在または将来の作業・運用を示すlatest artifactである。採用しないsourceのtaskやcredentialを残すと、実施予定または注入必須と誤読されるため、GetXAPIへ一本化した現在の方針と一致させる必要がある。

Decision Recordとdaily recordはlatest実装のinventoryではなく、当時の判断とsession事実を保存する履歴である。過去にTwitterAPI.ioを検討・実装した記録を消すと、なぜ当時その選択をしたか、いつ前提が変わったかを追跡できない。現行artifactの整理と履歴の保存は寿命が異なる。

## 3. Rejected

1. TwitterAPI.io実装をfallbackとして残す案。fallbackのproduction要件と切替条件がなく、未使用codeを正当化できない。
2. TwitterAPI.io codeだけを残し、inventoryとtaskだけを消す案。発見可能な実装が第二のsupported sourceに見え、保守境界が曖昧なまま残る。
3. 過去のDecision Recordとdaily recordも削除する案。latest policyとの見た目は揃うが、当時の事実と判断の追跡可能性を失う。
4. 過去のDecision本文をGetXAPI前提へ書き換える案。過去時点の判断を現在の結論で上書きし、記録のidentityと時系列を壊す。
