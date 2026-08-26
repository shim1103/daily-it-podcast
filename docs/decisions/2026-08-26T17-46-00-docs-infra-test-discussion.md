---
name: Generator に Consumer-Driven Contract Test を今は導入しない
date: 2026-08-26T17:46:00
branch: docs/infra-test-discussion
---

## 1. Decision

1. Generator に Consumer-Driven Contract Test を **今は導入しない**。
2. 出口契約（`secrettransport` / `commandlaunch`）と vendor Adapter の期待ずれは、既存の Sociable Unit と secret なし Narrow Integration で検知する。
3. Design by Contract（precondition / postcondition / invariant）の検証義務は全 Scope に直交して残る。CDC を入れないことと DbC を捨てることは同義ではない。

## 2. Reason

1. Contract Test は相手を起動せず consumer / provider の interface 期待一致だけを見る分類である。同一 repo・同一 Composition 結線では、期待ずれは既に Unit / Narrow の失敗として現れる。同じ期待表を CDC に写すと再 assert になる。
2. CDC が効く典型は、consumer と provider が別 repo / 別 team で並行進化し、起動コストの高い Narrow を毎回回せない場合である。現状はその前提が無い。
3. 種類として「将来必要かも」だけで suite と命名・runner を増やすのは、今必要な最小構造を超える（YAGNI）。

## 3. Rejected

1. 出口契約ごとに CDC file を先に空導入する案 — A で固定すべきは gate / local 収集境界であり、CDC の空 file は別契約の先行投資になる。
2. Narrow をやめて CDC だけにする案 — 外部境界の実 I/O は Narrow が既定。CDC は相手非起動のため実 I/O 契約の代替にならない。
