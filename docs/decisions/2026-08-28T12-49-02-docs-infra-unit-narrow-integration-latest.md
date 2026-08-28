---
name: M1は既存gateを最小配線でgreenにしSU/NI責務分離は後続taskへ残す
date: 2026-08-28T12:49:02+09:00
branch: docs/infra-unit-narrow-integration-latest
---

## 1. Decision

1. CompositionのHTTP Adapter移行Issue（M1）は、移行後の依存形で既存Unit / Integration gateがpassする状態までを完了条件に含める。
2. Sociable Unitから境界I/O観測を除きNarrowへ移す責務分離の仕上げは、vendorごとの後続SU/NI taskが所有する。
3. M1はCursor / `commandlaunch`経路を変更しない。

## 2. Reason

Adapter constructorと`secrettransport`削除は、既存のgemini / gdrive Narrowや混在Sociable Unitをコンパイル不能または契約不一致にする。配線だけ変えてtestを赤のまま完了扱いにすると、次sessionが「移行欠陥」と「未分割の観測」を同時に背負う。

一方、M1に4 vendor分の責務分離ACまで入れると、配線移行の失敗とtest分割の失敗が同じIssue赤になる。最小配線更新でgateを緑にし、分離の仕上げをvendor taskへ分けると、どちらが壊れたか名前で判別できる。

Cursor経路はHTTP移行と直交し、child environment再設計待ちである。M1へ混ぜると完了条件がprobe結果に依存する。

## 3. Rejected

1. M1をproduction codeだけに限り、test赤を後続へ残す案。gateが壊れた状態を完了と誤読する。
2. M1に全vendorのSU/NI責務分離を含める案。Issueが大きくなり、配線と分割のFault Isolationが失われる。
3. M1でCursor launcherも同時移行する案。未実測のchild env判断をHTTP移行へ結合する。
