---
name: SourceItem の Detail/Discourse は SourceBody{Text,Links}、Meta は帰属のみ（空可）。記事URLは Detail.Links、議論URLは Discourse.Links
date: 2026-09-13T17:14:00
branch: feature/generator-textwriter-prompt-source-criteria
---

## 1. Decision

1. **`models.SourceItem` の本文側**は次とする（型の正本は A の `source_item.go`）。
   - `Summary string` — 短い要約・題名側
   - `Detail SourceBody` — purpose1（事実）側の材料
   - `Discourse SourceBody` — purpose2（議論・反応）側の材料
   - `Meta string` — item_id / actor など **帰属のみ**。作者等が無ければ **空文字でよい**
2. **`SourceBody`** は `{ Text string, Links []string }` とする。
   - `Text` — その Purpose 側の本文（Discourse は 1 テキスト塊。1 要素 = 1 人の反応、ではない）
   - `Links` — その Purpose 側の URL だけ
3. **URL の置き場（Purpose 空判定のため）**
   - **記事・事実の URL** → `Detail.Links`
   - **議論ページの URL** → `Discourse.Links`
   - `Meta` に Purpose 判定用 URL を置かない
4. **空の意味**
   - `Detail` が空（`Text` 空かつ `Links` 空）→ その item に **P1 経路なし**
   - `Discourse` が空 → **P2 経路なし**
   - 本文が空でも該当 `Links` が非空なら、その Purpose は **書いてよい**（必要なら URL を fetch）
5. **`links[]` を型で持つ理由（TextWriter tool だけでは足りない）**
   - tool 許可だけでは「どの URL が P1 / P2 か」と「空判定」が壊れない形で渡らない
   - Application が全 `Links` 件数を数え、TextWriter の fetch 上限に張りそうなら **先に UseCase が fetch して `Text` を厚くする**余地を残す（実装は C。Adapter は HTML 全文 fetch しない — 先行 `2026-09-02T14-41-02` を維持）
6. Port / `SourceItem` に purpose tag は置かない（`2026-09-13T15-08-55` / 本 Decision の空判定 + TextWriter）。

本 Decision は session Decision `2026-09-13T15-09-10` の「`Detail`/`Discourse` を string」「写像は空のまま」を **部分 supersede** する。置き換え範囲＝字段形と URL 置き場と空の意味。目的・源の主戦力は `2026-09-13T15-08-55`。源ごとの raw 写像表は後続 Decision `2026-09-13T17-14-10`。

先行 `2026-09-02T14-41-02` の「`links:` を Context に載せる」は、Context 廃止後の置き場として本 Decision の `Detail.Links` / `Discourse.Links` へ移す（Adapter が HTML fetch しない点は維持）。

## 2. Reason

opaque `Context` や単一 string に URL を混ぜると、TextWriter が「Detail 空＝P1 禁止」と読んだとき、記事 URL が Meta や Discourse 側にしか無いと誤判定する。記事 URL は P1、議論 URL は P2 に紐づける。

Gemini 等で tool を許可しても、モデルが任意 URL を選んで取ることはあっても、件数上限があり、Purpose 別の空判定契約にはならない。`Links` を型で持つと、brief 列挙と UseCase の件数カウントが同じ契約になる。

`Meta` に URL を置かない。帰属（id / 作者）と Purpose 材料を混ぜない。作者が無い源は `Meta` 空でよい。

## 3. Rejected

1. **参考 URL をすべて `Meta` に置く案** — P1/P2 空判定が壊れる（記事 URL があるのに Detail 空と誤読される）。
2. **`Detail`/`Discourse` を平文 string のまま URL を埋め込む案** — UseCase が件数を型安全に数えられない。Purpose 側の URL 集合が曖昧。
3. **`Discourse []string`（1 要素 = 1 反応）** — Adapter 共通契約ではない（`15-09-10` Rejected を維持）。
4. **tool 許可だけに任せ `Links` を持たない案** — Purpose 区別と fetch 予算管理が Port から消える。
5. **Adapter が外部記事 HTML を fetch する案** — `14-41-02` Rejected。先 fetch するなら Application（C）。
6. **`Meta` 必須（作者無しでも埋める）案** — 無いものを捏造しない。空 enable。
