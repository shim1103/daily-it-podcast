---
name: 情報取得 Port は SourceID と発生時刻だけを必須にし、残りは Context へ寄せる
date: 2026-08-19T13:25:20
branch: refactor/generator-source-port
---

## 1. Decision

Generator の情報取得 Port は情報源非依存の `ItemSource` とする。戻りは `SourceItem`。必須 field は次の2つだけ。

1. `SourceID` — Adapter が自己申告する slug。Entities に情報源カタログを置かない
2. `OccurredAt` — **各情報源データに付いている発生時刻**（UTC の `time.Time`）。取得時刻でも CLI 実行時刻 `now` でもない。text 化は RFC3339

必須以外（本文、item_id、actor_id、permalink、links、media、帰属名など）はすべて `Context`（`string`。非 schema）。空文字は余りなし。Application は `Context` を parse しない。TextWriter へそのまま渡す。

`ItemSource.List(ctx, since)` は時間窓内の全件を 1 配列で返す。1 要素の粒度は Adapter が決める。Application は情報源の個数・種類・監視対象一覧を知らない。

X の GetXAPI と TwitterAPI.io は同一情報源である。両方の `SourceID` は `"x"`。vendor 名を `SourceID` にしない。切替は Composition Root。並列取得の対象ではない。

UseCase の `now` は CLI 実行時刻であり、`since = now - FetchWindow` を作るためだけに使う。`OccurredAt` には入れない。

## 2. Reason

Application の sort / group に必要なのは発生時刻と情報源識別子だけである。optional field を Port に置くと構造化になり、情報源ごとの欠落を schema が表現し始める。余りは opaque text に閉じ、取れた行だけを Adapter が書く。

同一人物の推論に使う `item_id` / `actor_id` は TextWriter が `Context` を読むための材料であり、一意性保証（Adapter の義務）のための必須 schema ではない。

X の2 vendor は代替 mechanism であり情報源が2つではない。既存 decision（vendor を facade にしない）と両立する。

## 3. Rejected

1. 現行 `Post` を optional field だらけにして情報源を足す案 — Port の構造化。空 field が情報源の形を漏らす
2. `ItemID` / `ActorID` / `ActorName` を必須にする案 — 全情報源が埋められない。一意性は Adapter `@ensure`
3. 単一 Actor を必須にする案 — ニュースの会社と記者、会社のみ、匿名を1 field に潰す
4. Application が `WatchUserIDs` を回す案 — Application が X の監視対象を知る
5. Entities に情報源カタログ（id+名の表）を置く案 — 種類の集合を内側が管理する
6. `OccurredAt` を取得時刻または CLI `now` にする案 — 情報の時刻ではない
7. GetXAPI と TwitterAPI.io を別 `SourceID` または並列情報源にする案 — 同一 X の代替 vendor
8. 第2情報源向けの空 Adapter / 並列 composite を今足す案 — YAGNI
