---
name: SU から Narrow へ移した実プロセス境界の production code は generator coverage gate の計測分母へ secret なし Narrow を算入して担保する
date: 2026-08-30T11:52:00
branch: feature/generator-cursorcli-narrow-integration
---

## 1. Decision

問い: SU（Sociable Unit）から Narrow Integration へ移した production code（実プロセス境界を持ち Unit では埋められない）を、generator の coverage gate でどう担保するか。

1. generator の Unit coverage 計測の分母に、**secret なし Narrow Integration（`apps/generator/test/`）を含める**。SU + secret なし Narrow Integration の合算 profile に対して statement coverage gate を課す。計測 flag と閾値の数値は正本の `scripts/generator/test-unit.sh` と `DESIGN.md` を参照し、本文へ写さない。
2. 計測から除外するのは Composition Root（`internal/composition/**`）・CLI Driving Adapter（`cmd/**`）・build tag 付き local 実物 suite・Broad Integration 以上（結線 / 状態伝播 / 連携を対象とする test。現時点で generator に存在しないが、追加時は除外対象）とする。
3. Integration gate の**実行対象**分離（secret なし Narrow のみ）は先行 Decision `2026-08-26T17-42-00-docs-infra-test-discussion.md` を正とし、本 Decision はそれを変更しない。「gate 実行対象の分離」と「coverage 計測分母」は別軸である。
4. 本 Decision は先行 Decision `2026-08-17T14-45-00-chore-test-and-ci-coverage-layer.md` の「除外は Composition Root と cmd のみ／Integration に coverage を載せない」を**部分 supersede** する。置換範囲＝coverage 計測分母に secret なし Narrow を算入する点。維持範囲＝statement 基準・閾値の水準・`depguard` 層 lint・除外方針の管理方法（方針で管理し薄い file だけ個別列挙）。先行 Decision の file 本文（frontmatter 含む）は一切書き換えない。

## 2. Reason

1. なぜ Narrow を計測分母に入れるか: `testing-strategy/coverage.md` §3 の表が「外部境界を持つ処理単位は Unit 90%+ / Narrow Integration 高め、実 I/O 契約は Integration でしか検証できない」と明記する。同 §4 最終段落は、全体 threshold と層別 threshold を併用する時「層別で緩めた分だけ全体閾値も緩めるか、層別対象を全体閾値の水準まで引き上げるか」を選べと要求する。SU / NI の boundary を `testing-strategy/levels.md` §7 通り分離すると、実プロセス起動が必須のコード（`processenv.Launch` / `buildChildEnv`）は SU から必ず外れる。これを計測分母から外したままにすると §4 が禁じる「層別で緩めた分を全体で放置」に該当し、実ロジックが gate をすり抜ける（`buildChildEnv` のバグが Unit gate で検出されない）。secret なし Narrow は実プロセスを起動して `Launch` の全経路・`buildChildEnv` の大半を通しており、算入すれば §4 の後者（全体閾値の水準まで引き上げる）を満たす。
2. なぜ Broad Integration 以上は入れないか: `testing-strategy/minimization.md` §2「上位 Scope は下位 Scope が所有する内部詳細を再 assert しない」。Broad Integration は結線・状態伝播・error 伝播を対象とし、その行を初めて通す test ではない。Broad で稼いだ coverage は「配線が通った」ことの副産物で、行のロジックの正しさは SU / NI が担保しているはずである。Broad を分母に入れると「Broad さえ通れば SU を書かなくてよい」逃げ道ができ、Pyramid（`minimization.md` §1）が崩れる。
3. なぜ先行 Decision `2026-08-26T17-42-00-docs-infra-test-discussion.md` と矛盾しないか: あちらの Reason 3 は Integration gate の**実行対象**（どの test を CI で走らせるか、secret を持つか）の分離であり、「CI 緑＝本番経路実測」の誤読回避が目的である。本 Decision は coverage の**計測分母**（走った test が production の何行を通したか）の話である。secret なし Narrow は既に gate で走っており（`2026-08-26T17-42-00` で許可済み）、`scripts/generator/test-unit.sh` は build tag を付けないため走るのは secret なし Narrow に限られる。その実行結果を coverage 集計へ算入するだけで、新たな credential リスクも本番副作用も生じない。
4. なぜ閾値の数値を据え置くか: 先行 Decision `2026-08-17T14-45-00-chore-test-and-ci-coverage-layer.md` と `2026-08-22T16-50-00-chore-generator-ci-test-configuration-hardening.md` が statement 90% を固定済みである。分母に secret なし Narrow が入ると `processenv` 等の実測値は上がる（本 session 実測で package coverage 32% → 97%、gate 全体 87.8% → 91.9%）ため、閾値を変えずとも gate は通る。閾値変更は別問いであり、本 Decision では扱わない。

## 3. Rejected

1. `processenv` package を coverage 除外リストへ個別追加する案 — `coverage.md` §5 の除外条件（実行環境制約で動かない / framework runtime 依存 / 外部 service wrapper で実通信必須）のいずれにも該当しない。`processenv` は child environment 構築という実ロジックを持ち、test 環境で実プロセスを起動して検証できる。除外すると `buildChildEnv` のバグが gate をすり抜ける。`coverage.md` §5「除外は『なぜ除外できるか』の方針で管理」に反する。
2. 閾値を下げる案（90 → 実効値） — `coverage.md` §2「数値は結果指標であり目標ではない」の逆で、gate の検出力を落として数値へ合わせる操作である。`processenv` 以外の劣化も同時に見逃す。
3. SU から実 child test を除去した変更を撤回し、SU に実プロセス観測を戻す案 — `levels.md` §7 の SU / Narrow 二重化へ逆戻りする。reviewer が §7 準拠と判定した boundary 整理を無効化する。同じ観測を SU と NI の両方が持つ重複になる（`minimization.md` §2）。
4. Broad Integration も一律に coverage 計測へ含める案 — Pyramid（`minimization.md` §1）が崩れ、上位 Scope で下位の穴を埋める運用を許してしまう。
