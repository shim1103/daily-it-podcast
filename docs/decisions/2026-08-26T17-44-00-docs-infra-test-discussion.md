---
name: Generator test の秘密供給は gate 無・local 実物は AgentSecrets test 値・本番 GHA は test に使わない
date: 2026-08-26T17:44:00
branch: docs/infra-test-discussion
---

## 1. Decision

1. Generator の **Integration gate** は secret 値を使わない（dummy / Fake 機構）。
2. Generator の **local 実物** suite（実 AgentSecrets binary / keychain / proxy）が使う値は、OS keychain 上の **test 専用値**とする。本番値を載せない。
3. **本番 GHA secret** は本番（または remote）実行用 processenv 供給に限り、test suite・Integration gate・local 実物 suite へ流用しない。
4. production 実行の置き場が processenv であることは既存 Decision を正とし、本 Decision は **test がどの秘密を持つか**だけを答える。

## 2. Reason

1. 秘密の置き場2軸（local = AgentSecrets、remote/本番 = processenv）は既に固定されている。test 側で「gate にも本番 GHA を載せる」「local 実物に本番値を使う」と混ぜると、軸は残っても検証対象が本番副作用になる。
2. `credential` は Integration を dummy、System / E2E を test 専用、本番 credential の CI/test 流用を禁止する。gate を dummy に揃え、local 実物だけ test 専用 keychain 値を使うのは、その Scope 別方針の project 適用である。
3. shim の keychain に test 用全 key があり本番は別値、という運用前提がある。local 実物は「機構は実・値は test」で self-validate でき、本番永続領域や本番 vendor key を叩く必要が gate / local 実物の必須条件にはならない。

## 3. Rejected

1. Integration gate に CI 用 GHA test secret を今すぐ載せる案 — 実外部を CI Narrow に載せる判断がまだ無い。今載せるのは供給路だけの先行投資になる（YAGNI）。
2. local 実物に本番 keychain 値を使う案 — 本番永続・本番 API を test が触る経路になり、禁止事項と同型の事故面を開く。
3. 本番 GHA secret と CI/test 用 secret を同一値にする案 — 流用禁止を形だけ分けて実体で破る。
