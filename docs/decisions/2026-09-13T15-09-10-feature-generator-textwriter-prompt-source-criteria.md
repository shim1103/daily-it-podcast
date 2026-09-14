---
name: SourceItem は opaque Context ではなく Summary・Detail・Discourse で渡し、purpose tag は置かない。raw→field 写像は未確定のまま残す
date: 2026-09-13T15:09:10
branch: feature/generator-textwriter-prompt-source-criteria
---

## 1. Decision

1. **`models.SourceItem` の本文側 field**は次の 3 つとする（型・字段名の正本は A の `source_item.go`）。`Context string` は廃止する。
   - `Summary string` — 短い要約・題名側
   - `Detail string` — 事実の厚い説明側
   - `Discourse string` — comment / reaction / discussion / reply / repost / quote 等を包括する **1 テキスト**（空文字可）
2. **`Discourse` の契約は「議論・反応を表すテキスト塊」であり、「1 要素 = 1 人の反応」ではない。** Adapter が「みんなの反応」として既に 1 本へまとめた文を返す場合も、そのまま `Discourse` に載せる。複数発言を載せる場合の区切り・順序・要約方針は Adapter 内部の写像（C）であり、Port は配列契約を持たない。
3. **Port / `SourceItem` に purpose（P1/P2）tag は置かない。** Purpose の意味と源の主戦力分担は Decision `2026-09-13T15-08-55` を正とする。振り分けの実行は TextWriter と prompt（原稿詳細）だけが知る。
4. **各情報源の raw（API/RSS）を Summary / Detail / Discourse のどれへどう載せるかの写像は、本 Decision では確定しない。** 推測写像を書かない。写像の確定は後続の実装 Issue（C）の範囲とし、未決の間は Adapter 実装を意図的に空（または title→Summary のみの一時接続）にしてよい。
5. 外部記事 HTML の全文取得を Adapter / Application に置くかは、先行 Decision（`2026-09-02T14-41-02-feature-hackernews-api-adapter.md`）の「`links:` に載せ Adapter は fetch しない」を維持する前提とし、採用後の tool fetch 運用は未実測のため本 Decision の確定対象にしない。

本 Decision は先行 Decision（`2026-08-19T13-25-20-refactor-generator-source-port.md`）の「余りは opaque `Context`」を **部分 supersede** する。置き換え範囲＝本文の載せ方を 3 field に分ける点。維持範囲＝`SourceID` + `OccurredAt` 必須、Application は源種類をカタログしない、`ItemSource.List` の形。

## 2. Reason

opaque `Context` 1 本だと、TextWriter が「事実の厚み」と「議論」を見分けにくい。Purpose 判断は prompt 側に置く方針（Fetch / ProduceEpisode は原稿構造を知らない）なので、材料の形だけを分けて渡し、purpose tag で二重管理しない。

`Comment` では reply / repost / quote / reaction を表しきれない。`Discourse` を包括名にする。`Description` は RSS の短摘要と外部記事本文のどちらを指すか紛らわしいため、厚い事実側は `Detail` とする。

`Discourse` を `[]string` にすると、「1 スロット = 1 発言者」という契約が暗示される。Adapter 共通契約としては、既にまとめた反応文を返す源と、個別 comment 列を持つ源の両方がある。後者を要素列に固定すると前者は `Discourse[0]` に押し込む歪みになる。Port は **1 string** に揃え、まとめ方は Adapter 写像（C）に閉じる。

raw→field 写像を今決めないのは、源ごとに「RSS 短文が Summary か Detail か」「複数 comment をどう 1 テキストへ載せるか」など未実測・未合意が残っているためである。推測で写像を Decision に書くと C がそれを契約と誤読する。空のまま残し、C で源ごとに確定する。

## 3. Rejected

1. **`Context` を残し行書式だけで Summary/Detail を表現する案** — agent が判別しやすい構造を型で固定したい。opaque のままでは同じ判断が再発する。
2. **`Comments` / `Comment` を字段名にする案** — reply / repost / quote / reaction を含意しにくい。
3. **`Reactions` または `Engagement` を字段名にする案** — reaction 偏重、または metrics（score 等）と誤読されやすい。本文テキスト列の名前としては `Discourse` を採る。
4. **`Description` を字段名にする案** — RSS `description` 短文と外部記事本文が同名衝突しやすい。
5. **`Body` を字段名にする案** — HTML body 全体や wire JSON と誤読されやすい。
6. **`Discourse []string`（1 要素 = 1 反応）にする案** — Adapter 共通契約ではない。まとめた反応文を返す Adapter が `[0]` に押し込むことになる。Port は `string` とする。
7. **`SourceItem` に `Purpose` tag を足す案** — 原稿詳細の知識が Port に漏れ、Adapter 固定 or Application 分岐を招く。意味と主戦力分担は `2026-09-13T15-08-55`、実行は TextWriter。
8. **本 Decision で源ごとの raw→field 写像表を確定する案** — 未実測・未合意を Decision 化したことになる。C / 必要なら追加 Decision へ送る。
