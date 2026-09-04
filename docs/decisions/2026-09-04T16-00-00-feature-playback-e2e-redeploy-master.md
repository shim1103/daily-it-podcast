---
name: manuscript body.opening / body.ending は定型込みの朗読全文であり Drive 上の SSoT とする
date: 2026-09-04T16:00:00
branch: feature/playback-e2e-redeploy-master
---

## 1. Decision

1. `contracts/manuscript.schema.json` の `body` 必須 field は `opening` / `topics` / **`ending`** とする（旧 `closing` を `ending` へ改名）。
2. `body.opening` は **定型挨拶（`OpeningGreetingTemplate` を date 注入した文）+ intro** をこの順で連結した朗読全文とする。
3. `body.ending` は **closingSummary + 定型締め（`ClosingFarewell` を date 注入した文）** をこの順で連結した朗読全文とする。
4. Drive 上の `{episodeId}.json` は **TTS が読み上げる原稿そのもの**の SSoT である。Playback は同じ本文を表示に使う。Generator は TTS 束の先頭・末尾と同一の全文を `body.opening` / `body.ending` へ入れる（契約へ「渡す」のではなく、読み上げ原稿を契約へ載せる）。
5. 定型を application 内だけに隠し、契約 JSON から落とす設計は採らない。連結 delimiter は TTS 束と同じ改行 1 個とする。
6. 本 Decision は先行 Decision（`2026-08-29T14-13-00` の「OpeningGreeting を TTS 前に組み立てる」）を維持しつつ、完成稿 JSON へ定型込み全文を載せる点で上書きする。farewell を JSON に載せない運用・テスト前提は破棄する。

## 2. Reason

1. Drive の JSON が Gemini 朗読の正なら、定型を JSON 外に置くと契約と音声が乖離し、Playback 表示も音声と一致しない。
2. application が定型を知って組み立てるのは実装都合であり、契約の field 意味を狭めてはならない。
3. `closing` は「まとめだけ」と誤読されやすい。締め定型を含む朗読全文として `ending` と改名する。

## 3. Rejected

1. `body.closing` に closingSummary だけを残し farewell は TTS 専用とする案 — 契約が朗読 SSoT でなくなる。
2. `body.opening` に挨拶だけを残し intro は topics 前に暗黙とする案 — 同上。
3. 定型を schema の別 field（`greeting` / `farewell`）へ分ける案 — Playback と TTS が 2 field を再結合する必要が生まれ、Drive 1 file の朗読単位が増える。

## 4. 後続 supersede

`origin/develop`（PR #127 `2026-09-04T16-44-46`）との merge で、`body.opening` / `body.ending` は `{ text, startSec }` object・delimiter は改行 3 個へ更新した。§1 の field 名（`opening` / `topics` / `ending`）と「朗読全文を契約へ載せる」方針は維持。詳細は `2026-09-04T19-30-00-feature-playback-e2e-redeploy-master.md`。
