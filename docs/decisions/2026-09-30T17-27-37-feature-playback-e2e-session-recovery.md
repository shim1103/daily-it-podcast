---
name: playback の運用 CLI は Node の独立区画 cli/ に置き、console を呼ぶ箇所は既存の logger.ts の 1 点に保つ
date: 2026-09-30T17:27:37
branch: feature/playback-e2e-session-recovery
---

## 1. Decision

1. e2e の Access session を扱う運用 CLI は `apps/playback/cli/` に置き、`test/support/` には置かない。層は、`storage-state.ts`（Entities。IO を持たない）、`command.ts`（Delivery Mechanism。IO は注入される）、`main.ts`（結線。分岐を持たない）とする。
2. `cli/` は coverage の分母に含める。coverage の除外は増やさない。
3. `console` を直接呼ぶ file は、既存の `worker/src/routes/logger.ts` の 1 点のままにする。biome の `noConsole` の例外も、この 1 file だけである。CLI の出力は、`main.ts` がこの file の `writeLine` と `writeErrorLine` を注入する。
4. `cli/` は `web`・`worker`・`contracts` を import しない。例外は出力の実体（`logger.ts`）の 1 点だけで、その import は `main.ts` に限る。逆向き（`web`・`worker`・`contracts` から `cli/`）も禁止する。dependency-cruiser で検査する。

## 2. Reason

1. `test/support/` は test 専用の double と helper の置き場である。運用に使う production の code を置くと、責務が2つになり（SRP）、coverage の分母からも外れる。CLI は test を支えるのではなく、運用の手段として動く。
2. `DESIGN.md` §1 は、runtime を互いに import させない。CLI は Node の runtime で動くので、Worker の BFF（`worker`）にも、ブラウザの `web` にも混ぜず、独立した区画にする。root 直下の dir は層の名前とし、共有物は責務を名乗る package に置く（`ring-model` §4-7）。
3. 出力の実体（`console` への直接書き込み）は境界 1 点が持つ（`logging-boundary` §1・§6）。この repo では、その 1 点が `logger.ts` である。例外の file を増やすほど、1 点の保証が崩れる。
4. 除外を増やして gate を通すと、gate が守るものが減る。到達できない分岐はコードから消し、実在する入力は test で覆う。標準入力を読む部分を、テストできる関数に切り出せば、`main.ts` は分岐が 0 になり、除外せずに分母へ入る。

## 3. Rejected

1. **`worker/src/` の中に置く案** — Worker の BFF の層に、Node の運用 CLI が混ざる。`worker` の責務（R2 読取 BFF）と、変更の理由が別である。
2. **`test/support/` に置き続ける案** — test 専用の置き場に運用 code が混ざり、SRP を破る。coverage の分母にも入らない。
3. **`process.stdout` を直接書く案** — 出力の実体が、1 点から散る。
4. **`cli/` に専用の console helper を足し、biome の例外を 2 file にする案** — 例外が増え、`console` を呼ぶ箇所が 1 点でなくなる。既存の 1 点を使えば足りる。
5. **`logger.ts` を root 直下の共有 dir へ移し、worker と cli の両方から使う案** — root 直下に置けるのは層の dir だけ（`ring-model` §4-7）で、共有 dir は作れない。worker 側の import も全て書き換えることになる。
6. **`main.ts` を coverage から除外する案** — `main.ts` が分岐を持たなければ、除外は要らない。除外の理由が「分岐を持つ」なら、その分岐を切り出してテストする。
