---
name: composition の production graph 組み立ては factory 型を作らず結線関数の直書きにする
date: 2026-08-30T11:20:00
branch: feature/generator-composition-produce-episode-wiring
---

## 1. Decision

Composition Root で `ProduceEpisode` の production graph を組み立てる単位は、`config.Config`
を受け取る**関数**（`newProduceEpisode` / `NewProduceEpisodeFromEnv`）とする。`FooFactory`
struct + `NewFooFactory` + `Build()` の 3 要素を持つ factory 型は作らない。

composite `ItemSource`（登録順 concat・error 透過・非 nil 空 slice）は残す。これは
先行 Decision（`2026-08-19T13-25-20-refactor-generator-source-port.md`）の「複数源 merge は
Composition が composite で行い Application は源個数を知らない」を満たすための分岐を持つ
実体であり、factory の有無とは別問題。

## 2. Reason

Composition の結線関数には分岐も状態もない。ctor を順に呼んで 1 個の struct を返すだけで、
`config.Config` を field に保持する以外に factory 型が足す責務がない。factory 化すると
`newProduceEpisode` が `NewProduceEpisodeFactory(cfg).Build()` へ 1 行委譲するだけの中間層に
なり、呼び出し側から見た構造が「関数 1 つ」から「型 + ctor + method」へ増える。
design-philosophy §2-3 KISS の「意図を表現するのに必要な最小の構造」に照らすと、増えた
3 要素は意図（＝結線）を表現していない。

factory パターンが正当化されるのは、組み立て途中の状態を跨いで持つ、生成物の種類を
引数で切り替える、部分適用した builder を呼び出し側へ渡す、といった振る舞いがあるとき。
現在の Composition Root にそれは無い。将来 `ProduceEpisode.Run`（D）や第 2 情報源 Adapter が
入っても、結線先の ctor が増えるだけで結線関数の形は変わらない見込み。

composite を残す理由は factory と切り離せる。`compositeItemSource.List` はループ concat と
error 早期 return の分岐を持ち、その振る舞いは `item_source_test.go` の Sociable Unit が
所有する。結線関数（分岐なし）は test を持たず `go build` と cmd 経路の compile で守る。

## 3. Rejected

1. `ProduceEpisodeFactory` 型を導入し `Build()` で組み立てる案 — 当初この branch の達成契約
   file が要求していたため一度実装した。しかし `Build()` の中身は旧結線関数の本体を struct +
   method へ移しただけで、状態も分岐も増えず、`newProduceEpisode` が 1 行委譲する死んだ
   中間層になった。契約 file の記述だけを根拠に構造を足す判断だったため撤回した。
2. factory を残しつつ結線関数を消して `NewProduceEpisodeFromEnv` から直接 `Build()` を呼ぶ案 —
   中間関数は 1 個減るが、型 + ctor + method の 3 要素は残る。KISS 後退の本体は factory 型
   そのものなので解消にならない。
3. composite も廃して GetXAPI Adapter を直接 `FetchSourceItems` へ渡す案 — 先行 Decision の
   「複数源 merge は Composition が composite で」に反する。GetXAPI 1 本の現在でも、
   第 2 情報源が入ったとき Application 側の変更を不要にする境界を先に置く。
