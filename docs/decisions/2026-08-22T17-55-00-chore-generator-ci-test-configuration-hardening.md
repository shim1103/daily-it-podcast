---
name: generator coverage は statement gate を維持し condition coverage は local report に限定する
date: 2026-08-22T17:55:00
branch: chore/generator-ci-test-configuration-hardening
---

## 1. Decision

1. generator Unit の statement coverage 90% gate を維持する
2. `gobco` `v1.3.4` は Boolean condition coverage を読む local report として使う
3. `gobco` の report は CI、hook、threshold gate に置かない

## 2. Reason

1. Go 標準 coverage は statement / block の到達を測る。既存 gate はこの能力と一致する
2. `gobco` は condition 実行を補助的に可視化できるが、未使用 function と `select` を完全には測れず、完全 branch coverage ではない
3. 対象漏れのある metric を hard gate にすると、数値が品質契約より強くなり、無意味な test を誘発する

## 3. Rejected

1. statement coverage を branch coverage と呼ぶ案（計測能力と名称が一致しない）
2. `gobco` の数値へ threshold を置く案（condition metric の対象外を品質 gate から漏らす）
3. coverage gate を撤去して condition report だけにする案（未到達 code の既存検出を失う）
