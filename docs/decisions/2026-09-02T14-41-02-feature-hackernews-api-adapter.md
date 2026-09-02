---
name: SourceItem.Context は源データの写像だけを持ち、外部記事 URL は links 行に載せるが Adapter は fetch しない
date: 2026-09-02T14:41:02
branch: feature/hackernews-api-adapter
---

## 1. Decision

主 Decision（`2026-09-02T14-41-00-feature-hackernews-api-adapter.md`）で追加する3 Adapter が `SourceItem.Context` に何を書き、外部 URL をどう扱うかを定める。

1. **`Context` の行書式は源ごとに違ってよい。** 先行 Decision（`2026-08-19T13-25-20-refactor-generator-source-port.md`）の「必須は `SourceID` と `OccurredAt` の2つだけ、残りは opaque な `Context` に取れた行だけ書く」を3源それぞれで具体化したもの。行の並びや key 名は Adapter stub の実装コメントを正本とする。
2. **engagement 指標を `Context` に載せない。** HackerNews の `score`、Lobsters の comment 数などの人気度は書かない。題材の価値判断は TextWriter の責務であり、Adapter は源データの写像だけを持つ。
3. **外部記事 URL は `links:` 行に載せる。Adapter は fetch しない。** main text（記事本文 / comment 群 / 記事要約）だけで解説が書ける状態を Adapter が保証したうえで、補助参照として元記事 URL を `links:` に置く。その URL を読むかどうかは TextWriter（Cursor CLI）の判断に委ねる。Adapter 内でも Application 層でも URL fetch はしない。

## 2. Reason

- **`Context` の行書式が源ごとに違うのは port 契約どおり。** `ItemSource` は `Context` を parse しない opaque text として扱う（先行 Decision `2026-08-19T13-25-20-refactor-generator-source-port.md`）。HackerNews は `text:` に comment 群、ITmedia は `title` + `description`、と中身が違うのは「取れた行だけ書く」の帰結であって不整合ではない。共通なのは `SourceItem` 型へ落とすことだけ。
- **engagement 指標を載せない理由。** score や comment 数は「その story がどれだけ話題か」の指標だが、podcast の題材選定は TextWriter が担う。Adapter が人気度を `Context` に混ぜると、原稿が「HN で N point」のような人気度を判断材料にし始め、philosophy §5-1 一貫性（源をまたいで同じ判断は同じ結論）を崩す。先行 Adapter（getxapi）も engagement 指標を `Context` に入れていなかった前例に揃える。
- **URL を `links:` に載せるだけで fetch しない理由。** Adapter が URL 先を fetch すると、その Adapter の変更理由が「源 API の変更」と「任意サイトの HTML 抽出ロジックの変更」の2つになり SRP に反する。媒体ごとに崩れる HTML 抽出は philosophy §4-2 が避けるべき強力（＝壊れやすい）処理でもある。一方「main text があり、さらなる参照として URL がある」形は先行の X 経路（tweet 本文＋参照 URL）と同型で、TextWriter が web_fetch できなくても解説が成立する。fetch の要否は TextWriter が判断すればよく、Application 層に fetch 機構を置くと TextWriter vendor の能力に Application が依存してしまう。

## 3. Rejected

1. **score / comment 数を `Context` に載せる案（HackerNews / Lobsters の特性を反映）** — 人気度が原稿の判断材料に混じり、源をまたいだ一貫性を崩す。題材の価値判断は TextWriter の責務。
2. **`Context` の行書式を3源で完全に統一する案** — 取れない field を空行で埋めるか、源固有の情報（comment / 記事要約）を捨てるかになる。port 契約は「取れた行だけ書く」であって「全源同じ行」ではない。
3. **URL fetch を Adapter 内で行う案** — Adapter の変更理由が2つ（源 API / HTML 抽出）になり SRP 違反。壊れやすい HTML 抽出を Adapter が抱える。
4. **URL fetch を Application 層に置く案** — TextWriter vendor（Cursor CLI）が web_fetch できるかに Application が依存する。fetch の要否は TextWriter の判断で、Application は源の取得と合成だけを持つ。TextWriter が実際に web_fetch できるかの実測と、できない場合に TextWriter 経路の内側で補完する設計は未決（lane 送り）。
