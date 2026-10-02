---
name: 複数 ItemSource の並行合成は Composition が持ち、Application UseCase は単一 ItemSource だけを知る
date: 2026-10-02T07:17:48
branch: refactor/generator-io-fanout
---

## 1. Decision

1. 複数 `port.ItemSource` を並行に `List` し結果を連結する合成（`composition.compositeItemSource`）は、**Composition Root** が所有する。
2. Composition は各情報源 Adapter を生成し、合成 Port を組み立てて `FetchSourceItems` へ **単一** `port.ItemSource` として渡す。Application UseCase（`FetchSourceItems` / `ProduceEpisode`）は情報源の個数・種類を知らない（[[2026-08-19T13-25-20-refactor-generator-source-port]]）。
3. 合成の fan-out 手段（`errgroup`・limit）は [[2026-09-23T17-21-29-refactor-generator-go-performance]] に従う。本 Decision は**置き場**だけを決める。
4. Composition の結線関数に factory 型を足さない結論（[[2026-08-30T11-20-00-feature-generator-composition-produce-episode-wiring]]）は維持する。
5. Adapter 内部の I/O fan-out（HN id 列・Lobsters detail・R2 GET 等）は各 Adapter が所有する。Application の `TextWriter` / `SpeechSynthesizer` **fallback 方針**は Application が所有する。これらは ItemSource 合成と**別責務**であり、同じ「合成」語で一括移動しない。

## 2. Reason

1. 問いは「Composition か Application か、どちらへ丸ごと移すか」ではない。**責務ごとに分ける**ことである。源個数を UseCase から隠す graph 組み立て・並行 List・fail-all・concat は Composition の束ね方の一部。fallback の順序・番兵は UseCase 方針。Adapter 内の vendor I/O 並行は Adapter 実装詳細。
2. 「Composition はビジネスロジックを書かない」は Domain 手順や vendor I/O を書かない意味であり、Port 合成による依存グラフの具体化まで禁じない。源個数の知識は Composition に閉じ、UseCase 依存面には出さない。
3. Application の TextWriter / Speech 合成と形が似ていても、変更理由が違う。同型に見えたから Application へ寄せるのは、責務分けではなく層間の**移設**になる。

## 3. Rejected

1. **TextWriter と同型だから合成を Application（`fetch`）へ移す案** — 「集約は Application」と読んで層間移設した。ItemSource 合成は源個数隠蔽の graph 組み立てであり、fallback 方針とは別責務。正は責務で分けることであり、package 間の丸ごと移設ではない。
2. **合成を廃し、各 Adapter を UseCase が直接複数受け取る案** — Application が源個数を知ることになり [[2026-08-19T13-25-20-refactor-generator-source-port]] に反する。
3. **合成を Infrastructure の共有 helper に置く案** — vendor 非依存の graph 組み立てを infra に落とすと、Composition Root の特権結線とずれる。
4. **Composition 例外としてだけ固定し、責務軸を書かない案** — 「Composition か Application の 2 択」に見える。実際は Composition / Application fallback / Adapter 内部の少なくとも 3 軸がある。
