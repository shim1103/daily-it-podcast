---
name: 原稿 JSON の structured 出力拘束は Gemini TextWriter だけが持つ
date: 2026-09-29T02:10:39
branch: cursor/gemini-structured-yesterday-window-6a75
---

## 1. Decision

1. `generateContent` の `generationConfig` に `responseMimeType=application/json` と WriterOutput 形の `responseSchema` を載せるのは **`geminiapi.TextWriter` のみ**。
2. schema の正本は repo 根 `contracts/writer-output.schema.json`（`contracts.WriterOutputSchema`）。Adapter 内へ JSON を手写ししない。
3. Cursor Cloud Agents Adapter（`cursorapi`）には structured 拘束を足さない。prompt + `ManuscriptDraftFromWriterOutput` の Domain 検証のまま。
4. schema が保証するのは wire の **形**（object / required field / topics 要素の field）まで。topic 件数・rune 数・日本語含有・句点は Domain 検証が正本のまま。
5. brief prompt は全 field（とくに `topic.title`）を日本語必須と明示し、英語・ASCII だけの見出しを不合格例で禁じる。

## 2. Reason

1. Cursor Cloud Agents API（`POST /v1/agents`）に JSON schema / response format パラメータは無い（公式 endpoints・OpenAPI）。拘束を Cursor へ足すには API 外の別経路が要り、現行 Adapter 契約を壊す。
2. Gemini `generateContent` は structured output を公式に持つ。primary / spare の Gemini 経路だけでも free text 崩れ（fence・欠落 key）を減らせる。
3. `contracts/` は言語横断の wire 正本（既存 manuscript schema と同型）。Adapter が embed 経由で読むので手写し schema が第二の SSOT にならない。
4. 本番失敗（`topic[n].title has no japanese`）は shape ではなく Domain。structured と prompt 日本語強化を同時に置く。

## 3. Rejected

1. **Cursor にも structured 相当を求める案** — API に手段が無い。prompt 強化だけが残る。
2. **Cursor primary を捨てて Gemini のみにする案** — fallback 順序の別 Decision と独立。本判断の scope 外。
3. **schema に日本語 `pattern` や rune `minLength` を全部載せる案** — Gemini schema subset と Domain 正本の二重管理になる。形だけ schema、意味は Domain。
4. **geminiapi 内に schema JSON を直書きする案** — WriterOutput 形の第二 SSOT。`contracts/` 経由が正。
5. **tools（google_search / url_context）を外して schema だけ取る案** — source URL 深掘りの現行契約を捨てる。tools 併用は Gemini 3 系の公式案内に従い維持し、実 API で 400 なら別 Decision で切り替える。
