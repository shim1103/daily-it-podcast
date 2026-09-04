---
name: playback の実行コードを持たない型宣言のみの file は coverage.include から外す。runtime-config-bindings.ts が該当
date: 2026-09-04T18:30:01
branch: docs/frontend-coverage-broad-integration
---

## 1. Decision

1. playback の Vitest coverage 計測対象から、**実行可能な文を 1 つも持たない型宣言のみの file**（`export type` / `export interface` だけの file）を外す。`vitest.config.mjs` の `coverage.include` の glob を絞るか `coverage.exclude` に追加するかは A artifact の判断とし、除外した file と理由をこの Decision で記録する。
2. 現時点の該当 file は `worker/src/composition/runtime-config-bindings.ts`（`PlaybackVariables` / `PlaybackSecrets` / `PlaybackEnv` の型宣言のみ）。
3. `*.d.ts`・`*.config.*`・test fixture は既に既定 exclude（`coverageConfigDefaults.exclude`）で外れており、本 Decision の対象外。本 Decision が足すのは「repo 内 source として書かれているが実行コードを持たない型 module」だけ。
4. 型 module に対して、型が存在することを確認するだけの test（`expectTypeOf` の羅列等）を書いて分母に載せることはしない。

## 2. Reason

1. v8 coverage は実行可能文を持たない file を分母 0 の 0% として表示する。これは「検証漏れ」ではなく「計測対象が無い」状態だが、per-file 表に 0% 行が残ると、`2026-08-25T23-28-28` §1-2 が求める「global（原則 100%）と個別 glob の間に矛盾が残らない状態」の見た目が崩れ、次に coverage を見た agent が「未カバーの実ロジック」と誤読する。`vitest/coverage.md` §2 も「除外前に計測すると 0% file が閾値を引き下げ、閾値の意味がなくなる」として `*.d.ts` 等の除外を求めており、型のみ file はこれと同性質。
2. `runtime-config-bindings.ts` は Cloudflare Workers bindings の shape を型で表明するだけの module で、分岐も代入も関数呼び出しも無い。この型が正しいことは、`env` を消費する `runtime-config.ts`（実 parse・validation を持ち SU で 100%）と `root.ts`（結線、SU で 100%）の test で間接的に検証されている（`testing-strategy/minimization.md` §4「値そのものが SSoT である確定値定義層に対し、値をそのまま assert で書き写す test は二重管理で検出力を持たない」と同じ理由で、型宣言に専用 test を付けない）。
3. `testing/coverage.md` §1 は「結線専用の入口 package が実行経路の都合で未実行のまま閾値を割る時、製品ロジックを歪めて稼ぐのではなく除外へ寄せてよい。ただし除外した事実は coverage 計測の SSOT へ残す」とする。型のみ file は「結線専用 package」よりさらに実行成分が無く、除外の妥当性は明確。この Decision が「除外した事実の SSOT」を兼ねる。
4. 型の存在確認 test を書いて分母に載せると、`minimization.md` §4 が禁じる「値をそのまま書き写す存在確認 test」と同型の、型定義を変えるたび test も書き換わる二重管理になる。型の回帰は `tsc --noEmit`（`typecheck` script）が担う。

## 3. Rejected

1. `runtime-config-bindings.ts` に `expectTypeOf<PlaybackEnv>()` 等の型 test を書いて 100% にする — 実行文が増えるわけではなく v8 の per-file 表示は変わらない（型 test も実行時は no-op）。かつ型定義の変更ごとに test を追随させる二重管理になる（`minimization.md` §4）。
2. file 全体ではなく行単位の `/* v8 ignore */` を型宣言へ付ける — `v8 ignore` は「到達不能な実行分岐」への注記（`2026-08-25T23-28-28` §1-3）。実行文が存在しない型宣言に付けるのは用途違いで、file 冒頭に file 全体 ignore を書くのは include から外すのと同義かつ意図が読みにくい。
3. 型宣言を実ロジック file（`runtime-config.ts` 等）へ吸収して独立 file を無くす — `root.ts` と `runtime-config.ts` の両方が `PlaybackEnv` を import しており、片方へ寄せると依存の向きが不自然になる。境界の型を独立 module に置く構成自体は妥当で、変えるべきは coverage の計測対象の方。
4. global threshold を下げて 0% file を許容する — `2026-08-25T23-28-28` Rejected §1 が退けた「global を諦める」方向。命名規約から漏れた file を機械検出する global 100% の基本線を崩す。
