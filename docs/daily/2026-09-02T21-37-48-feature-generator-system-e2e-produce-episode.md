---
name: Generator System e2e 緑化（Gemini 応答 parse バグ修正・TTS 削減・CursorCLI 尺安定化）
date: 2026-09-02T21:37:48
session_id: none
branch: feature/generator-system-e2e-produce-episode
prev: 2026-08-30T23-47-00-feature-generator-system-e2e-produce-episode.md
---

## 1. Summary

前回まで「Gemini 日次 quota」「preview の Limitation」と記録していた System 失敗の真因が、Gemini Interactions API の応答 parse バグ（audio は `output_audio.data` ではなく `steps[].content[].data`）だったと判明し、修正して既定 System gate を緑化した。あわせて TTS 呼び出しを topic+2 束へ削減、Gemini retry を同種 error 2 連続で打ち切りへ変更、System suite を用途別 test（TTS 単体 / Gemini 以外 full / CursorCLI 単体 / full）へ分割、`workflow_dispatch` で単一 test を選べるようにした。Cursor 原稿が尺下限を割る flaky は brief prompt へ尺の逆算誘導と Self-check を入れて解消し、CursorCLI 単体 test が 3 回連続 PASS。判断は Decision `2026-09-02T13-55-00` / `13-56-00` / `13-57-00` / `16-57-00` / `18-01-00` / `18-26-00`。lessons へ 7 件追記。

## 2. Changes

- `workflow_dispatch` 検証: run 33581258235 / 33582558173（parse バグで decode_pcm・429）→ 33607924229 / 33609034783（診断 snippet 拡充で `steps[].content[].data` 構造を特定）→ 33610705667（parse 修正で TTS 単体 + Gemini 以外 full が緑、go test 本体 403s）→ 33612303963（Cursor 原稿 2858 文字で尺下限割れ flaky + TTS 当日 RPD 枯渇 429）→ 33627209650（`-f test_run=TestCursorCLIDraftSystem` で CursorCLI 単体のみ、3 連続 PASS。全体 3825 / 3620 / 3473 文字）。
- run 33616026675 は `workflow_dispatch` が全 system test を一括実行する仕様を見落として dispatch し、TTS 系が走る前に即キャンセル（quota 追加消費なし）。再発防止に `SYSTEM_TEST_RUN` / dispatch input `test_run` を追加。
- Gemini 以外 full test（`gemini_excluded_full_system_test.go`）で実 OAuth + 実 Drive の list / write（create+upload）/ read / delete が初めて実到達・緑化（run 33610705667）。実 Cursor draft 1 回の所要は 95〜320s（run 33627209650）。
- Decision は本 session 内で 6 本作成済み。lane（`generator-lane.md`）と引き継ぎ（`generator-system-e2e-produce-episode.md`）を更新済み。
- 診断用に一時拡大した `bodySnippetMax`（400→4000）は構造確定後 400 へ戻し、トップレベルキー一覧の付与（`topLevelKeysHint`）は恒久で残置。
- Domain 定数（`entities/constants` の `Draft*` / `CharsPerSecond` / 秒数 / `ModelID`）は全変更を通して不変。

### Commits

- `41a751a`
- `a2c8686`
- `b78b310`
- `f9c2494`
- `b4d2c6d`
- `4ca9589`
- `c776db6`
- `eb72d8d`
- `92eb16e`
- `7e9e569`
- `cdfcf78`
- `9d0dc9e`
- `197a6b4`
- `2b23274`
- `7a1c2a1`
