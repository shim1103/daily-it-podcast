---
name: Cursor CLI Narrowはchild environment再設計まで後回しにする
date: 2026-08-28T12:49:01+09:00
branch: docs/infra-unit-narrow-integration-latest
---

## 1. Decision

1. Cursor CLI AdapterのsecretなしNarrow（既存`generator-narrow-gate-vendor-cursorcli`）は、HTTP側のComposition移行およびvendor SU/NI最新化より後に実施する。
2. 実施タイミングは、GitHub Actions capability probeの結果を踏まえたchild environment再設計の判断が固まってからとする。
3. 本DecisionはCursor Narrow task fileの本文を変更せず、実施順だけを固定する。

## 2. Reason

Cursor経路の未実測はbinary path、argv、最小child environmentであり、HTTP Adapterの`secrettransport`廃止とは別軸である。HTTP側を先に閉じないと、全体の依存図が「全vendor同時」に見え、並列できない作業まで待ち行列に入る。

child environmentのallowlistや起動契約が再設計されると、Narrowが観測すべきI/O契約も変わる。probe前にNarrowを完成させると、再設計でobservableが無効になる。

task本文を今書き換えないのは、契約内容そのものは変わっておらず、変えるのは着手順だけだからである。着手順はDecisionとlaneが持てば足りる。

## 3. Rejected

1. Cursor NarrowをHTTP SU/NIと同じバッチで実装する案。child env未確定のまま境界I/Oを固定し、再設計で書き直す。
2. Cursor Narrow taskをtodoから削除する案。後回しであり破棄ではない。入口を消すと再設計後に契約を再発明する。
3. probe完了前にchild env再設計Issueを切る案。Verificationが未実測のままC化する。
