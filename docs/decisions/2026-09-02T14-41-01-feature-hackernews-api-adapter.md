---
name: 議論系の情報源（HackerNews / Lobsters）は本文・comment を text 抽出できる JSON 経路で取る
date: 2026-09-02T14:41:01
branch: feature/hackernews-api-adapter
---

## 1. Decision

主 Decision（`2026-09-02T14-41-00-feature-hackernews-api-adapter.md`）で採用した議論系2源の取得経路を定める。

1. **HackerNews / Lobsters は JSON 経路を採る。** HackerNews は Firebase JSON（`topstories.json` → `item/<id>.json`）、Lobsters は `hottest.json` → `/s/<short_id>.json`。
2. **判定軸は「main text（記事本文または comment 本文）を安定して text として取り出せるか」。** 両源とも story 自己テキストは外部リンク型で空になりうるため、`kids` / `comments` の本文を main text として `Context` に載せ、comment だけで解説が成立する状態にする。
3. **報道系（ITmedia NEWS）は専用 Adapter で公式 RSS（`rss.itmedia.co.jp` の RSS 2.0）を読む。** `title` + `description` を main text にする。comment は無い。`description` が無い item は `title` のみで `SourceItem` 化する。「RSS 汎用 Adapter」は作らず、`infrastructure/itmedia/` を HackerNews / Lobsters と同格に置く。RSS 2.0 の parse は標準 `encoding/xml` で行う。

## 2. Reason

実測で判定軸を確認した。

- **HackerNews に運営元公式の RSS は無い。** Firebase JSON が唯一の公式経路。story ごとに `kids` を個別取得する N+1 になるが、1日1回・rate limit 無しの用途では取得上限件数の足切り（Adapter stub の定数が正本）で有界に収まる。
- **Lobsters の RSS（`/rss`）は判定軸を満たさない。** `description` が `<p><a>Comments</a></p>` のみで、記事本文も comment も含まない。JSON（`/hottest.json` → `/s/<short_id>.json`）は `comment_plain` を含む全 comment を story 1 件あたり1リクエストで返し、HTML 除去の手間も少ない。
- **議論系は comment を main text にする。** HackerNews / Lobsters とも外部リンク型 story では story 自己テキストが空になる。`title` だけでは原稿の題材にならないため、comment 本文を `Context` に載せて「記事を読まなくても議論の中身で解説できる」状態を Adapter が保証する。
- **報道系は要約で足りる。** ITmedia に読者 comment は無く、RSS `description` の記事要約が「話題になった事実」を伝える main text になる。深い議論は HackerNews / Lobsters が担うので、ITmedia に本文全文取得を要求しない。
- **「RSS 汎用 Adapter」を作らず ITmedia 専用にする。** 実源は ITmedia 1 本で、複数媒体の RSS 差分を 1 module で吸収する「汎用」は将来の投機（philosophy §2-4 YAGNI）。先行 Decision（`2026-08-17T17-15-01-feature-x-getxapi-adapter.md`）が「vendor ごとに Adapter を置き、vendor 差分を吸収する facade にしない」「偽の共通概念を YAGNI で作らない」と定めており、汎用 RSS Adapter はその facade に当たる。ITmedia 専用にすれば `SourceID` と dir 名と中身が一致し、HackerNews / Lobsters と対称になる。別媒体（Publickey 等）を足す時は各々専用 Adapter を新設し、RSS 2.0 parse ロジックの重複が三度現れてから共通化を検討する（同 Decision「同じ知識が三度現れてから見直す」）。

## 3. Rejected

1. **Lobsters を RSS 経路で取る案** — RSS の `description` に本文も comment も無く、判定軸「main text を安定して text 抽出できるか」を満たさない。Lobsters は専用 JSON Adapter にする。
2. **HackerNews / Lobsters も `title` だけで `SourceItem` 化する案（comment を取らない）** — 外部リンク型 story が大半で、`title` のみでは技術的トレードオフの解説に踏み込めない。comment 取得の N+1 コストは1日1回・上限件数の足切りで許容範囲。
3. **ITmedia の記事本文全文を取得する案** — RSS に本文は無く、記事 URL 先の HTML 抽出が要る。媒体ごとに崩れる強力（＝壊れやすい）処理で、報道枠の「事実を伝える」目的には要約で足りる（philosophy §4-2）。
4. **RSS 汎用 Adapter（feed 種別リストで複数媒体を1 module が捌く）案** — 複数媒体の RSS 差分を吸収する facade であり、先行 Decision（`2026-08-17T17-15-01-feature-x-getxapi-adapter.md`）が退けた「vendor 差分を吸収する facade」「偽の共通概念（YAGNI）」に当たる。`SourceID` が feed 種別ごとに要るのに 1 dir 1 `SourceID` と食い違う。ITmedia 専用にする。
5. **RSS parser に外部ライブラリ（`gofeed` 等）を最初から入れる案** — ITmedia の RSS は方言の少ない RSS 2.0 で、標準 `encoding/xml` で読める。依存は最小に保ち（philosophy §4-3）、別媒体で方言に当たってから検討する。
