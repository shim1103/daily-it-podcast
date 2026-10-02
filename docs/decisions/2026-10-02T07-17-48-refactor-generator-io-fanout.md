---
name: ItemSource 合成の構成は Composition、並行 List・fail-all・連結の振る舞いは Application
date: 2026-10-02T07:17:48
branch: refactor/generator-io-fanout
---

## 1. Decision

1. **構成**（どの情報源 Adapter を何本選び、単一 `port.ItemSource` として `FetchSourceItems` へ渡すか）は **Composition Root** が所有する。`FetchSourceItems` / `ProduceEpisode` は源個数・種類を知らない（[[2026-08-19T13-25-20-refactor-generator-source-port]]）。
2. **振る舞い**（並行 `List`・concurrency 上限・fail-all・結果連結）は **Application**（`fetch.CompositeItemSource`）が所有する。TextWriter / Speech の合成と同型に、Composition は Adapter 列を Application ctor へ渡すだけにする。
3. 合成の fan-out 手段（`errgroup`・limit）は [[2026-09-23T17-21-29-refactor-generator-go-performance]] に従う。本 Decision は**構成と振る舞いの分け方**を決める。
4. Composition の結線関数に factory 型を足さない結論（[[2026-08-30T11-20-00-feature-generator-composition-produce-episode-wiring]]）は維持する。
5. Adapter 内部の I/O fan-out は各 Adapter が所有する。Application の TextWriter / Speech **fallback 方針**も Application が所有する。ItemSource 合成の振る舞いと変更理由が違っても、どちらも「N 実装を 1 Port 面でどう呼ぶか」は Application 側に置く。

## 2. Reason

1. 「わける」は package を動かさないことではない。`composite` という語に **構成**（誰を・いくつ）と **振る舞い**（どう呼ぶか）の 2 軸が同居していたのを切ることである。移設＝概念ごと丸投げ、わける＝軸ごとに owner を分ける。
2. Composition Root は logic を持たない（繋ぐだけ）。観測可能な呼び出し規約（並行・fail-all・連結）と Sociable Unit が要る対象は振る舞いであり、Application の方針に属する。構成の選択だけが Composition に残る。
3. 「源個数を UseCase から隠す」は `FetchSourceItems` の依存面が単一 `port.ItemSource` であることであり、合成型の実装 package が Composition であることを要求しない。`manuscript.NewTextWriter([]port.TextWriter{...})` と同型。

## 3. Rejected

1. **並行 List・fail-all・連結を Composition に置く案**（本 file の一時的な答え）— Composition が logic を持つことになる。「構成は Composition」と「振る舞いは Application」を混同した。
2. **合成を廃し、各 Adapter を UseCase が直接複数受け取る案** — Application が源個数を知ることになり [[2026-08-19T13-25-20-refactor-generator-source-port]] に反する。
3. **合成を Infrastructure の共有 helper に置く案** — vendor 非依存の呼び出し方針を infra に落とすと、Composition の結線特権とずれる。
4. **同一 package 内の file/関数分割だけで責務分けとみなす案** — 層の線を越えないので、構成と振る舞いが同じ Composition に残る。
