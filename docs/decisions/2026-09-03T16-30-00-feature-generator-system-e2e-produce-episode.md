---
name: System gate は 1 回通しだけにし、full 系 test と Drive fake 経路を廃止する
date: 2026-09-03T16:30:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `generator-system.yml`（cron 週次 + dispatch）が回す System test は **1 回ずつの通し**だけにする。N 回ループも `-count=N` も入れない。PASS 率は cron が定常で測る対象ではない。
2. `apps/generator/test/system/` から次の 3 file を削除する。
   1. `gemini_excluded_full_system_test.go`（GetX/Cursor/OAuth/Drive を実物、Speech だけ fake）。
   2. `produce_episode_system_test.go`（`//go:build system && full`。入口→出口すべて実物）。
   3. `drive_observe.go`（上記 2 file 専用の Drive 観測 helper）。
   `//go:build system && full` タグは消滅する。
3. PASS 率・所要の計測 test（`tts_rate_system_test.go` / `draft_rate_system_test.go`、`//go:build system && ratemeasure`）と専用 workflow は残す。位置づけを「cron の 1 回通しが落ちた後、原因を切り分けるために手動 dispatch で回す道具」に限定する。定常実行しない。
4. rate 計測 test が読む env は本番用 env 名（`config.*APIKeyEnv`）ではなく `TEST_GEMINI_API_KEY` / `TEST_CURSOR_API_KEY` を直読みする。`config` パッケージ（本番 env 名の正準）は変更しない。
5. `generator-system.yml` / rate 計測 workflow の長い inline shell（PASS 率集計等）は `scripts/generator/*.sh` へ切り出す。YAML には script 呼び出しだけを置く。

## 2. Reason

1. shim の明示指示「週次は 1 回だけの system test の all 通しに委ね、落ちた時に初めて log を見て rate 計測 test などすればよい」「system test は 1 回だけであり N 回 loop は不要」。1 回通しで壊れは検出できる（壊れていれば 1 回で落ちる）。率が要るのは「壊れてはいないが安定しない」を追う時だけで、それは常時ではなく事後調査。cron を軽く保ち、率計測は必要時に手動で回す方が、消費 req も CI 時間も小さい。
2. 「gemini_excluded_full はやりません。普通に system test で全部通してその rate を測るだけです」。Drive だけ fake する中間 test は、通し（実 Drive 書込を含む）があれば不要な重複。full 通しも「1 回だけの system test の all 通し」に含意され、`system && full` という別 gate を分ける前提が消えた。`drive_observe.go` は消える 2 file の専用 helper なので道連れに削除する。
3. 通しが赤のとき原因を Cursor 経路 / Gemini 経路 / prompt 精度に絞る道具は要る。だが常時走らせる必要はない。dispatch 専用のまま、事後調査の位置に置く。
4. rate 計測は本番と別 key（TEST 枠）で回す。test コードが本番 env 名を読むと「本番 key を渡してしまう」経路が残る。TEST_ を直読みして、本番 key が rate 計測へ流れる経路を塞ぐ。`config` は本番 Adapter の契約なので触らない。
5. 長い shell を YAML へ直書きすると、shellcheck も単体確認も効かず、workflow diff がノイズで膨れる。script 化すれば `scripts/generator/*_test.sh` で確認でき、YAML は意図だけになる。

## 3. Rejected

1. system test 自体を N 回ループ + 閾値判定へ変える案 — shim が明示的に不要とした。cron のたびに Cursor draft を複数回叩く（1 回 95〜320s）のは重く、率は事後調査で足りる。
2. `generator-system.yml` に `go test -count=N` を足す案 — 率が出ず（1 回でも落ちれば赤）、ループ化と同じく cron を重くするだけ。
3. `produce_episode_system_test.go`（full）を課金枠移行後の最終確認用に残す案 — 「1 回だけの all 通し」が既にその役目を負う。`system && full` の二重 gate を残す理由がなくなった。必要になれば新 Decision で復活させる。
4. rate 計測 test / workflow ごと削除する案 — 通しが落ちた後の切り分け（Cursor 単体の安定性、prompt variant の A/B、callGap の効き）に使う。消すと赤の原因追跡が手作業に戻る。dispatch 専用なら cron を汚さない。
5. `config` に `TEST_*APIKeyEnv` 定数を新設して rate 計測 test に参照させる案 — `config` は本番 Adapter の env 契約の正準。計測専用の env 名を混ぜると責務が濁る。test file 内の直読みで足りる。

## 4. Supersede

先行 Decision `2026-09-03T14-45-00` / `14-46-00` / `14-47-00`（file 名で参照。日付順で本 Decision が新しい）に対し、次を置き換える。

- `14-45-00` の「cron・dispatch とも `-tags=system` の全 System test を実行する」は維持。ただし対象集合から `gemini_excluded_full` を除く（本 Decision 2-1）。TEST key 固定・実行パターン撤去・inline smoke 削除は維持。
- `14-46-00` / `14-47-00` の「rate 計測を専用 test / workflow へ分離する」構造は維持。ただし env 名を `TEST_*` 直読みへ改める（本 Decision 4）。「先頭 1 束 N 回」「Op=run 除外」「閾値 t.Fatalf」の各判断は維持。
- 3 先行 Decision に暗黙で含まれていた「`gemini_excluded_full` / `produce_episode`(full) を System gate 群として維持する」前提を破棄する（本 Decision 2）。

先行 Decision の本文・frontmatter は書き換えない。読み手は本 file を正とする。
