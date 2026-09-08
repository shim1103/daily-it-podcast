---
name: draft rate 計測は api 入力で cursor / gemini を切り替え、1 API に対して PASS 率を測る
date: 2026-09-08T07:40:00
branch: feature/generator-gemini-textwriter-prompt-fix
---

## 1. Decision

1. `generator-draft-rate.yml`（`workflow_dispatch` 専用）に入力 `api`（`cursor` | `gemini`、既定 `cursor`）を足す。1 回の dispatch は 1 API に対してのみ計測する（両方同時には測らない）。
2. rate 計測 test（`//go:build system && ratemeasure`）を `DRAFT_RATE_API` env で分岐する。
   1. `cursor`: key env は `TEST_CURSOR_API_KEY`、実物は `cursorapi.NewTextWriter`。
   2. `gemini`: key env は `TEST_GEMINI_API_KEY`、実物は `geminiapi.NewTextWriter`。
   3. どちらも `port.TextWriter.Write(ctx, brief)` だけを叩く。分岐後は共通の計測 loop へ入る。
3. 環境要因の除外は API 非依存に統一する。`*cursorapi.Error` / `*geminiapi.Error` のどちらでも `Op == "do"`（`client.Do` が失敗 = API へ到達すらできない）なら、その回を分母から除外する。`do` 以降（応答が返った後）の失敗は prompt 精度の範疇として分母に含める（Decision `2026-09-03T14-47-00` の踏襲）。
4. 選んだ API に対応する key env が空なら `t.Skipf`（計測外）。本番 env 名（`config.*APIKeyEnv`）は読まない。
5. `gemini` で default variant（`constants.TextWriterBriefPrompt`）が PASS 閾値（既定 0.8）以上を通すよう、prompt を修正する。修正は `constants.TextWriterBriefPrompt` を直接編集する（gemini 専用 variant は作らない）。修正内容は cursor 側の PASS 率を落とさない範囲に限る。

## 2. Reason

1. fallback 実装（Decision `2026-09-07T19-06-00`）で Gemini generateContent 経路が primary 枯渇時の secondary になったが、`geminiapi.ModelID` が `ManuscriptDraft` 検証（topic 3〜7・各 field 文字数・全体 8〜12 分・JSON 1 オブジェクト）を何割で通すかは未実測だった（`generator-lane.md` D 表「Gemini fallback 原稿の品質・token・尺」）。cursor と同じ台帳・同じ閾値で測れるようにするのが最小の解。
2. `api` を入力にして 1 dispatch = 1 API にするのは、PASS 率は「その API がどれだけ prompt を守るか」の指標であり、2 API を混ぜると率が API 構成比で揺れて読めなくなるから。cursor と gemini で所要も課金枠も違うので、必要な側だけ回せる方が運用しやすい。
3. 環境除外を `Op == "do"` で API 非依存に統一できるのは、`cursorapi.Error` と `geminiapi.Error` が同型（`Op string` / `Err error`）で、どちらも `client.Do` 失敗を `Op: "do"` で包むから。片方だけ特別扱いする理由がない。
4. prompt を専用 variant でなく `const` 直編集で直すのは、fallback 経路は本番で実際に使われる secondary であり、「cursor では通るが gemini では通らない prompt」を放置すると枯渇時に原稿が出せないから。variant は A/B 用の仕組みで、確定した改善は `const` へ寄せるのが Decision `2026-09-03T14-47-00` §3 の方針。

## 3. Rejected

1. gemini 専用 brief variant（`testdata/brief_prompt_variant_gemini.txt`）を作り default を触らない案 — fallback は本番経路なので、gemini が通らない default prompt を温存する意味がない。variant は「現行と改善案の同条件比較」用であって、API ごとに別 prompt を常用する仕組みではない。
2. 1 dispatch で cursor と gemini を両方測り率を並べる案 — job 所要が 2 API 分に伸び、片方の環境 skip がもう片方の分母に影響しないことの説明も要る。1 API 固定なら台帳が単純。
3. gemini の環境除外を `Op == "do"` 以外（`http_status` の 5xx 等）にも広げる案 — 5xx は Adapter が retry 済みで、それでも失敗したなら計測対象（応答が返らない = 実運用でも原稿が出ない）。`do` だけが「そもそも API に触れていない」。
4. `api` 未指定時に gemini を既定にする案 — 既存 dispatch 履歴・runbook が cursor 前提。既定は現行踏襲の `cursor` にして、gemini は明示指定で回す。
