---
name: Generator の Broad Integration と System は当面 Issue 化せず lane に残す
date: 2026-08-26T17:47:00
branch: docs/infra-test-discussion
---

## 1. Decision

1. Generator の **Broad Integration** と **System / E2E** は、当面 **実装 Issue（C）にしない**。lane（D）に残す。
2. 今 Issue 化してよい Integration 以上は、secret なし Narrow（gate）と、実 AgentSecrets 出口の local Narrow（command / HTTP）に限る。それらの Issue 本文は本 Decision の後続で作る。
3. Broad / System を Issue 化する条件は、Verification（何が最終 postcondition か、どの実物が必須か）が決められるようになってからとする。未実測のまま Issue 化しない。

## 2. Reason

1. Broad は Narrow では検出できない配線・状態伝播だけに使う。ProduceEpisode 本体が未完の時点では、Broad の Verification 境界が定まらず、Issue 化すると実装者に再判断を強いる。
2. System は入口から出口の最終結果に限定する。最終結果の定義が UseCase 未完で欠けるなら、D に残すのが scope-split の判定（実施 scope・Verification を確定できない）に合う。
3. Narrow 出口の local 実物と gate 用 Narrow は、1 境界と Verification が既に言える。Broad / System と同じ「Integration 以上」袋に入れないことで、Issue 単位の独立性を保つ。

## 3. Rejected

1. Broad / System を今まとめて1 Issue にする案 — Verification 未定のまま C の代用になり、D 逃げになる。
2. vendor 実 API Narrow を System に全部畳む案 — 1 vendor 境界の実 I/O は Narrow になり得る。System 必須ではない。ただし今は Verification 未整理のため、それ自体も D / 後続判断とする。
3. Broad を gate に空 file で先置きする案 — 振る舞い契約も Verification も無い空 suite は A（収集境界）の対象外であり、混乱だけが増える。
