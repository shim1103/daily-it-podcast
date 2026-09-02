---
name: TextWriter brief prompt に尺の逆算誘導とセルフチェックを入れ、CursorCLI 単体で安定性を検証する
date: 2026-09-02T18:26:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `constants.TextWriterBriefPrompt`（embed 専用 template）を次の方向で強化する。数値は `{{…}}` placeholder のまま（Domain 定数は変更しない）。
   1. **Length strategy**: 全体合計を満たすには topic を目安件数、各 detail を目安文字数で書く必要があることを、合計から逆算する手順として明示する。
   2. **Self-check**: 提出前に各 field の文字数 range・topic 件数・全体合計・各文の末尾「。」・日本語のみ・JSON 1 オブジェクトのみを確認するチェックリストを置く。
   3. **parse 注意**: この JSON は機械 parse され、1 つでも range 外だと不合格で書き直しになることを伝える。
2. `build.marshalWriterOutputExample`（`{{JSON_EXAMPLE}}` の生成元）を 1 topic の最小骨組みから、目安件数の topic を持ち各 field が「長さの目安になる」ダミー文の shape へ変える。LLM が example の形へ引きずられて短く書く傾向を減らす。
3. `apps/generator/test/system/` に **CursorCLI 単体 System test** を新規で置く（`//go:build system`）。GetX を呼ばず、ニュース生成に十分な固定の擬似 `SourceItem` を `build.ComposeBrief` へ渡し、実 `CURSOR_API_KEY` で `cursorcli.TextWriter.Write` → `build.ManuscriptDraftFromWriterOutput` を **複数回連続**で通す。毎回 valid な Draft が返ることと各試行の所要秒を検証・記録する。Gemini / Drive は呼ばない。
4. TTS 単体 System test（`speech_synthesis_system_test.go`）の `workflow_dispatch` は当面行わない。無料枠 RPD=15 を使い切っており、parse 修正の緑は run 33610705667 で実証済み。

## 2. Reason

1. run 33612303963 の「Gemini 以外 full」が `invalid_manuscript_draft: total rune count 2858 is out of range [3360, 5040]` で落ちた。`ProduceEpisode.Run` は 306s（Cursor draft + retry）走った後に検証失敗。前回 run 33610705667 では通っていたので flaky。原因は Cursor 出力の尺ばらつきで、TextWriterMaxAttempts=5 の retry でも下限に届かなかった。
2. Domain 定数上、全体下限 3360 は「3 topic × 各 field 下限」では構造的に満たせない（intro 140 + closing 140 + 3×(preface 70 + detail 336) = 1498）。目安件数（5）で各 detail を target 付近に書かせないと通らない。現行 prompt は各 field の range を並べるだけで、全体合計から件数・detail 長を逆算する誘導が無く、JSON_EXAMPLE も 1 topic・極短文なので、LLM が短い出力へ寄る。
3. 尺の誘導強化とセルフチェックは prompt 文言の調整で閉じる。Domain 定数（尺モデルは Decision `2026-08-30T03-06-53`）は変更しない。下限を下げる案は今回の対象外。
4. Cursor 出力の安定性は、GetX / Gemini / Drive を巻き込まずに単体で速く回して確かめたい。固定の擬似ソースなら Fetch 窓・情報源の状態に依存せず、prompt 調整の効果を再現性ある形で測れる。
5. TTS 単体の再 dispatch は quota を焼くだけで新しい情報が無い。parse 修正の効果は実証済み。

## 3. Rejected

1. `TextWriterMaxAttempts` を 5 から増やす案 — 先行 Decision（`2026-08-30T19-50-00` 系）で「モデルが系統的に短いため回数では収束しない」と分析済み。prompt の誘導不足を直す方が筋。
2. Draft 下限（`DraftTotalMinSec` 等の Domain 定数）を下げる案 — shim の明示許可が無い定数変更。尺モデルの意図を prompt の失敗で歪めない。
3. 擬似ソースでなく録画した GetX 応答を使う案 — GetX の実到達は別 test（Narrow / Gemini 以外 full）が持つ。CursorCLI 単体の目的は「十分な入力に対し安定して valid Draft を返すか」なので、入力は固定の擬似ソースで足りる。
4. CursorCLI 単体を 1 回だけ通す案 — flaky の検証にならない。複数回連続で valid を要求する。
