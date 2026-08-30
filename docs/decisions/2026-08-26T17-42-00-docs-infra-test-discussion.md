---
name: Generator の Integration gate は secret なし Narrow のみとする
date: 2026-08-26T17:42:00
branch: docs/infra-test-discussion
---

## 1. Decision

1. Generator の Integration gate（pre-push と GitHub Actions の Integration workflow、およびそれらが呼ぶ `scripts/generator/test-integration.sh`）が実行してよいのは、**credential / secret 値を使わない Narrow Integration** に限る。
2. 実 AgentSecrets binary・実 OS keychain・実 proxy・外部 service の本番または test 専用 credential を要する suite は、この gate に載せない。
3. gate の収集境界の正本は code（Integration 入口 script と build tag 契約）を参照する。本 Decision に path や tag 文字列を正本として書かない。

## 2. Reason

1. 現行 gate の invariant は本番 credential を読まない。実 keychain や実 vendor を gate に混ぜると、Repeatable な CI と local 解錠依存が同居し、失敗原因が「契約バグ」か「環境欠落」か判別できなくなる（Fault Isolation）。
2. Narrow の既定は外部境界の実 I/O 契約だが、gate 用 Narrow は **Go 側の出口契約**を dummy / Fake 機構で self-validate する層として既に存在する。機構そのもの（実 AgentSecrets）の証明は別入口に分離した方が、Least Privilege（test が持つ権限）と一貫する。
3. production 実行の秘密供給（processenv / GHA）は「本番 job」の話であり、Integration gate の話ではない。両者を同一 script に載せると「CI が緑＝本番経路が実測済み」という誤読が起きる（Least Astonishment）。

## 3. Rejected

1. gate Integration に実 AgentSecrets / 実 keychain を載せる案 — CI runner に local keychain が無く、解錠もできない。無人 CI と local 実物を同一 gate にすると flaky か skip だらけになる。
2. file 名だけで local 実物を「gate 対象外」とみなす案 — `go test ./test/...` は package 内の test file を名で除外しない。除外の正本にならない。
3. 本番 GHA secret を Integration gate へ流用する案 — `credential` の禁止事項（本番 credential の CI/test 流用）に反する。
