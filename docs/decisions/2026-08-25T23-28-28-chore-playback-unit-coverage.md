---
name: playback Unit coverage gate は branch coverage・per-glob threshold・到達不能分岐は v8 ignore で設計する
date: 2026-08-25T23:28:28
branch: chore/playback-unit-coverage
---

## 1. Decision

playback の Vitest Unit coverage gate は次の3点で構成する。詳細な数値・除外 path は `DESIGN.md` §5（Test 配置）を正本とし、ここでは方針のみ書く。

1. `coverage.provider: "v8"` かつ Branch Coverage 基準で判定する。
2. threshold は global（原則 100%）と、外部境界・状態分岐を持つ層への個別 glob threshold（90%）を併用してよいが、**個別 glob 対象 file を実質 100% まで unit test で埋め、global と個別 glob の間に矛盾が残らない状態にする**。
3. TypeScript の型安全のためだけに残る到達不能分岐（union の exhaustiveness check、フィルタ済み値への再チェックなど）は、test を書いて無理に通さず `v8 ignore` コメント＋理由コメントで計測から除外する。
4. Vitest の `test.projects` 配下では coverage を project 単位に設定できない。coverage 設定は必ず `defineConfig` の root top-level `test.coverage` に置く。

## 2. Reason

1. **なぜ global と個別 glob の両立ではなく「個別 glob 対象を実質 100% まで埋める」を選んだか**：Vitest の global threshold（`branches: 100` のような top-level 数値 key）は "All files" summary 全体（個別 glob 該当 file も合算）に対して評価される。個別 glob に 90% を許容していても、その file の実branch %が90%台である限りglobalの合算値は下がり続け、global 100% とは原理的に共存しない。これは vitest-dev/vitest issue #6165 で報告されている既知の挙動であり、vitest 側の bug ではなく仕様。回避策は「global を諦めて個別 glob だけで運用する」か「個別 glob 対象を実質 100% まで引き上げる」の二択で、今回は後者を選んだ。理由は、未カバー分岐の大半が外部 I/O の異常系（Drive API の不正 JSON 応答、token 欠落）であり、mock で十分再現でき test 追加コストが低かったため。個別 glob 90% の設定自体は、将来この層に新規 file が増えた時のセーフティネットとして残す。
2. **なぜ到達不能分岐を test で無理に通さず `v8 ignore` にしたか**：`testing-strategy/coverage.md` の既定方針（到達不能分岐は計測より除外が正確）に加え、実際に型で保証されている分岐（`string[]` の `??` fallback、2値 union の `default` 節）へ test を書こうとすると、mock で型安全装置を迂回する不自然な test になり、test の意図（実際に起こりうる異常系の検証）から外れる。`match-playback-route.ts` の `?? ""` は型上も実行上も不要と判明したため `v8 ignore` ではなく削除した。「削除できるか」と「削除できず ignore するしかないか」を分けて判断した。
3. **なぜ coverage 設定を root top-level に置くか**：Vitest 4.x でも `test.projects` 配下の個別 project へ coverage 設定を書ける type-safe な方法は提供されていない（vitest-dev/vitest issue #9470）。root で 1 つの coverage 設定を持ち、`--coverage` flag を `test:unit` script にのみ紐付けることで、`test:integration` 側には thresholds を持ち込まない（`vitest/coverage.md` の invariant: integration config に unit threshold を課さない）。

## 3. Rejected

1. **global threshold を諦め、全 file を個別 glob で管理する**：管理対象 glob が増えるたびに一覧の保守が必要になり、新規 file が glob に一致しない限りその file は無検査になる。global 100% を基本線として維持し、緩和が必要な層だけ個別 glob で明示するほうが、命名規約から漏れた file を機械的に検出できる。
2. **`coverage.exclude` に到達不能分岐を含む file 全体を追加する**：file 単位の除外は、同じ file 内の到達可能な分岐まで検査対象から外してしまう。分岐単位の `v8 ignore` の方が粒度が正しい。
3. **`vitest.config.mjs` の `projects` 配下 unit project に `coverage` を書いたまま運用する**：実際に動かして `exit=0` になったが、これは coverage 設定が無視され threshold 判定自体が起きていなかっただけだった（`ERROR:` ログが一切出ない状態で `exit=0` になっていたことから発覚）。見た目上 gate が通っているように見えても、設定が効いているかを実際の ERROR ログの有無で確認する必要がある。
