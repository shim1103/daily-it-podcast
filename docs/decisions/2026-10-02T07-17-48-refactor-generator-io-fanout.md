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

## 2. Reason

1. 「複数源を1つの Port に束ね、Application から源個数を隠す」は Composition Root の組み立て責務である（先行 [[2026-08-30T11-20-00-feature-generator-composition-produce-episode-wiring]]・[[2026-08-19T13-25-20-refactor-generator-source-port]]）。合成の実行（並行・fail-all・concat）は、その束ね方の一部であり、結線と同じ置き場に置く。
2. Application の `TextWriter` / `SpeechSynthesizer` 合成は **fallback 方針**（順序・番兵）を UseCase が所有する。ItemSource 合成は「源個数を UseCase から隠す」ための **graph 組み立て**であり、同型に見えても変更理由が違う。Application package へ移すと、UseCase 方針と graph 隠蔽が混ざる。
3. Composition Root の「ビジネスロジックを書かない」は、Domain 手順や vendor I/O を書かない意味であり、Port 合成による依存グラフの具体化まで禁じない。源個数の知識は Composition に閉じ、UseCase 依存面には出さない。

## 3. Rejected

1. **合成を Application（`fetch`）へ移す案**（本 file の一時的な答え）— Composition の「複数源 merge」責務と衝突する。shim の方針で撤回し、Composition 置き場へ戻す。
2. **合成を廃し、各 Adapter を UseCase が直接複数受け取る案** — Application が源個数を知ることになり [[2026-08-19T13-25-20-refactor-generator-source-port]] に反する。
3. **合成を Infrastructure の共有 helper に置く案** — vendor 非依存の graph 組み立てを infra に落とすと、Composition Root の特権結線とずれる。
