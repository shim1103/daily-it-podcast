---
name: HTTP vendorのSU/NI最新化はCompositionのHTTP Adapter移行の後に行う
date: 2026-08-28T12:49:00+09:00
branch: docs/infra-unit-narrow-integration-latest
---

## 1. Decision

1. GeneratorのHTTP Adapterを`*http.Client`とcapability Configへ移行するComposition接続（M1）を先に完了する。
2. getxapi / oauth / gemini / gdrive の Sociable Unit と Narrow Integration をtarget architectureへ揃える作業は、その移行の後に行う。
3. processenv / `secrettransport` 前提の旧vendor Narrow gate taskは、移行後形のSU/NI taskへ置き換える。旧gateを移行前に実装してから書き直さない。

## 2. Reason

HTTP Adapterの依存が`secrettransport`のままだと、Narrowはprocessenv bindingと独自Request表現を観測対象にしてしまう。target architecture（Decision `2026-08-27T13-56-14`）ではその境界自体が消えるため、移行前にgateを完成させても直後に破棄する。

Sociable UnitとNarrowの責務分離（Adapter内分岐と外部境界I/O）は残るが、配線の正本が変わるとtestの組み立て方も変わる。配線移行を先に固定し、その上で責務分離を仕上げれば、同一observableを二度実装しない。

旧gate taskを残したまま新taskを足すと、どちらが現行契約か判別できず、実装者がprocessenv形へ戻る。置き換えでlatestの正を1つにする。

## 3. Rejected

1. 現行`secrettransport`のままgetxapi/oauth Narrowを先に実装する案。移行で配線が変わるため二重作業になる。
2. M1と4 vendorのSU/NI仕上げを1 Issueに束ねる案。配線移行と責務分離の失敗原因が混ざり、Fault Isolationが崩れる。
3. 旧gate taskを履歴としてtodoに残す案。未完了入口が増え、採用しない契約が実施予定に見える。
