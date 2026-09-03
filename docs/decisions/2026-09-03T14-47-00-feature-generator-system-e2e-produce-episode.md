---
name: CursorCLI draft の prompt 精度は brief template を注入で差し替えつつ rate 計測し、環境経由 error は分母から外す
date: 2026-09-03T14:47:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. CursorCLI が prompt どおりの正しい draft 形式（`build.ManuscriptDraftFromWriterOutput` が valid Draft を返す）で応答する率を測る専用 test（`//go:build system && ratemeasure`）と専用 workflow（`generator-draft-rate.yml`、`workflow_dispatch` のみ）を新設する。
2. `build.ComposeBrief` を、brief template 文字列を **引数で受け取る** `ComposeBriefWithTemplate(items, template)` へ分離する。既定の `ComposeBrief(items)` は `constants.TextWriterBriefPrompt` を渡して委譲する。rate 計測 test は workflow input（`prompt_variant`）で `default`（現行 `constants.TextWriterBriefPrompt`）か代替 template（`apps/generator/test/system/testdata/brief_prompt_variant_*.txt`）を選び、`ComposeBriefWithTemplate` へ渡す。
3. 1 回の計測は、固定の擬似 `SourceItem`（既存 `seedSourceItems()` 相当）から組んだ brief で `cursorcli.TextWriter.Write` を呼ぶ。
   1. `*cursorcli.Error` の Op が `run`（launcher / subprocess 起因 = processenv 等の環境経由 error）なら、その回を **計測対象から除外**（分母に入れない）。
   2. 正常に応答が返った回だけを分母にし、`ManuscriptDraftFromWriterOutput` が valid なら PASS、parse / 尺 range / 件数で失敗したら FAIL とする。
4. `pass / (pass + fail)` が閾値（input `pass_threshold`、既定 0.8）以上で緑、下回れば `t.Fatalf`。各回の valid 可否・全体文字数・使用 variant・環境 skip 回数を `t.Logf` へ出す。繰り返し回数は input `runs`（既定 5）。

## 2. Reason

1. この workflow の目的は prompt 精度の向上であり、測るべきは「LLM が prompt の指示（field 文字数 range・topic 件数・全体尺・JSON 1 オブジェクト）をどれだけ守れるか」。processenv や `agent` 未解決、subprocess の exit 非 0 といった環境要因は prompt 精度と無関係で、これを分母に入れると率が環境ノイズで揺れて改善効果が読めなくなる。よって Op が `run` の error はスキップ扱い。`decode_envelope` 以降（Cursor が JSON envelope を返した後）は応答内容の問題なので prompt 精度の範疇として分母に含める。
2. prompt template を注入にするのは、改善案（逆算誘導の強さ・Self-check 項目・example の具体度）を `constants` 書き換えなしで A/B したいから。`testdata/` に variant を置き、input で切り替えれば、現行 prompt と改善案を同じ台帳で比較できる。`text_writer_brief_prompt.go` の `const` は既定値として温存し、確定した改善だけを後で反映する。
3. 入力を固定の擬似ソースにするのは、GetX の Fetch 窓・情報源の状態に依存せず prompt 調整の効果を再現性ある形で測るため（先行 Decision `2026-09-02T18-26-00` の踏襲）。GetX の実到達は別 test が持つ。
4. cron に載せない理由は、Cursor draft 1 回が 95〜320s かかり N 回で長時間になるため。dispatch 専用にして必要な時だけ回す。

## 3. Rejected

1. `cursorcli_draft_system_test.go`（既存の `system` tag、N=3 連続）に rate 計測を統合する案 — cron gate で毎回 N 回 Cursor を叩き、所要が読めなくなる。既存 test は 1 回版へ縮小し（回帰確認）、率の計測は専用 test へ分ける。
2. 環境経由 error（Op=run）も分母に含める案 — 率が Cursor CLI の install 状態・runner の負荷で揺れ、prompt 改善の効果測定にならない。
3. `run` に加えて `decode_envelope` も分母から外す案 — Cursor が envelope を返さなかったケースまで環境側とみなすと、prompt 起因で応答が壊れた失敗を見逃す。envelope が返った後の失敗は prompt 精度の対象。
4. prompt を `const` のまま直接編集して 1 案ずつ試す案 — 現行 prompt との同条件比較ができず、変更のたびに commit が要る。注入口を作れば variant 切替だけで A/B できる。
5. 閾値未達でも緑にして数値だけ記録する案 — dispatch 専用なので cron を汚さない。閾値で `t.Fatalf` にして run の赤/緑で改善要否を示す。
