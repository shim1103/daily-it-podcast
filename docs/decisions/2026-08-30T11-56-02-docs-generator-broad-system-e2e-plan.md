---
name: Generator System Test の file 名分類語は system とする
date: 2026-08-30T11:56:02
branch: docs/generator-broad-system-e2e-plan
---

## 1. Decision

1. Generator の System Test file 名に出す分類語は **`system`** とする（例: `*_system_test.go`）。
2. Generator では `system_e2e` や単独の `e2e` を System 分類語に使わない。
3. 収集除外の正本は build tag 契約であり、file 名だけを除外契約にしない。tag 文字列と入口 script の正本は code を参照する。

## 2. Reason

1. testing-strategy の汎用例は `system_e2e` だが、本 repo の Generator は CLI 入口→出口の System であり、playback のブラウザ E2E と語が衝突すると Scope 判別が壊れる。
2. `e2e` 一語は「実境界に届く」以上の情報を持たず、levels の System / E2E 併記を file 名へ落とすとどちらを指すか読み手が再判断する。`system` 一語に固定すれば Generator 側の分類が一意になる。
3. file 名は分類ラベル、gate 除外は build tag、という役割分担は先行 Decision（`docs/decisions/2026-08-26T17-43-00-docs-infra-test-discussion.md`）と同じ。名前を変えても除外の正本を file 名へ移さない。

## 3. Rejected

1. `system_e2e` を Generator でも使う案 — skill 例との字面一致だけを優先し、playback ブラウザ E2E との混線を残す。
2. `e2e` だけを使う案 — 実境界到達の有無以外が名前から読めず、System と狭義 E2E の区別が file 名で消える。
3. file 名に `system` を含めず dir や script 名だけで補う案 — naming-and-layout の「file 名自体へ分類を含める」に反する。
