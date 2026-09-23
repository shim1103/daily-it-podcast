---
name: apps/generator の I/O-bound 逐次処理を errgroup による fan-out へ移行する（対象と見送りの選別）
date: 2026-09-23T17:21:29
branch: refactor/generator-go-performance
---

## 1. Decision

1. `apps/generator` の並行化実装は `golang.org/x/sync/errgroup` を使う。`sync.WaitGroup` 手動実装・`channel` 手動 fan-in は使わない。
2. 以下を fan-out 化の対象とする。
   1. `compositeItemSource.List`（HackerNews / Lobsters / Publickey / TechCrunch / クラウドWatch の 5 情報源、`errgroup.SetLimit(5)`）。既存の `@invariant` 「並列化しない」は本 Decision で更新する。
   2. HackerNews `List` 内の個別 story 取得。早期終了（`limit` 到達 `break`）は廃止し、id 列全件を fan-out する。
   3. HackerNews `fetchTopLevelComments` 内の個別 comment 取得。
   4. Lobsters `List` 内の個別 story detail 取得。
   5. `CompletedEpisodeLookup.HasPair` の `jsonStems` 走査。早期終了（一致時点で return）は廃止し、全件 fan-out へ変更する。
3. 以下は見送る（fan-out 化しない）。
   1. `EpisodeWriter.Write` の json / mp3 PUT 並行化。
   2. `ProduceEpisode.Run` 内の WAV 尺計算ループ（segment 単位）。
   3. `ConcatWAV` 内の `parseWAV` ループ単体の並行化（重複計算の除去は別 Decision で扱う。[[2026-09-23T17-21-30-refactor-generator-go-performance]] 参照）。
4. 並行化しない（技術的に不適と確認した）箇所。
   1. `CompletedEpisodeLookup.listObjectKeys` の pagination（次ページ token が前ページ結果に依存）。
   2. TextWriter / SpeechSynthesizer の fallback chain（無料枠 → 有料枠の順で試す意図的な順序依存）。
   3. `ConcatWAV` の PCM 結合本体（発話順に連結する必要がある single accumulator）。

## 2. Reason

### 2-1. 前提事実（判断時に確認済みの現状）

1. Decision時点で `apps/generator` に `Goroutine`・`channel`・`sync.WaitGroup`・`errgroup` 等の並行処理は 1 箇所も存在しなかった（全 106 non-test file を対象に `go func`・`sync.WaitGroup`・`errgroup` 等を grep し 0 件を確認）。
2. HackerNews `List`（`internal/infrastructure/hackernews/item_source.go`）は、`fetchTopStoryIDs` で `topstories.json` から最大 500 件の id 列を取得し、先頭から 1 件ずつ `fetchStoryInWindow` して window（`time >= since`）に合う story が `MaxStoriesScanned`（20 件）に達したら `break` する構造だった。window に合う件数が少ない日は、id 列全長（最大 500 件）まで逐次 fetch する。
3. `CompletedEpisodeLookup.HasPair`（`internal/infrastructure/r2/lookup.go`）は、`jsonStems` を先頭から走査し、date 一致が見つかった時点で `true` を return する早期終了を持っていた。`@ensure` 契約自体は「同一 stem の json+mp3 があり json の date が一致するとき true」という結果のみを保証しており、早期終了は契約化されていない実装上の最適化に過ぎないことを確認済み（早期終了を撤廃しても契約違反にならない）。
4. `compositeItemSource.List`（`internal/composition/item_source.go`）は、5 source（HackerNews / Lobsters / Publickey / TechCrunch / クラウドWatch）を登録順に逐次呼び、結果を登録順に concat していた。既存の `@invariant` に「並列化しない」と明記されていた。結果配列内の source 順序を後続処理が意味付けて使っている契約は確認できなかった（`merged` の順序保証を求める `@ensure` は無い）。

### 2-2. 選択の理由

1. **`errgroup` を選んだ理由**：`apps/generator` は特別パフォーマンスに困っているわけではなく、KISS を優先する（shim 方針）。`errgroup` は semaphore（`SetLimit`）・`WaitGroup`・error 集約・`ctx` cancel 伝播を 1 API に束ねており、`sync.WaitGroup` 手動実装より実装量が少ない。`channel` 手動 fan-in は今回のような「並行 fetch → 集約」だけの単純な形には冗長（overengineering）。
2. **fan-out 対象を選んだ理由**：対象はすべて I/O-bound（外部 HTTP fetch）であり、各要素の処理が他要素の結果に依存しない（独立している）。特に 2-5（`HasPair`）は、通常運用では「同日ペアが存在しない」（= 早期終了が効かない）ケースがほとんどであり、レア（CI 設定 bug 等）な早期終了ケースを最適化するより、頻発する usecase（全件走査になるケース）に合わせて fan-out する方が実利がある（当初は非推奨と判断したが、頻発 usecase 優先の観点で撤回した）。2-1-3 の通り契約変更を伴わない。
3. **2-2（HackerNews story 取得）で早期終了を廃止した理由**：2-1-2 の通り、早期終了と fan-out は本質的に相性が悪い（並行実行では「何件目で `limit` に達したか」の判定が難しく、判定ロジックを足すと KISS を崩す）。id 列は最大 500 件で有界であり、全件 fetch のコスト増（HTTP request 回数増）より、実装の単純さを優先した。
4. **concurrency 上限 5 を選んだ理由**：HackerNews / Lobsters / Publickey / TechCrunch / クラウドWatch のいずれも専用 rate limiter・backoff 実装を持たない（= 相手側の実際の許容量は不明。executor agent による全 106 file 調査で確認済み）。相手の許容量が分からない以上、安全側に倒し、まず 5 で開始する。5 source 間の fan-out（2-1）は source 数自体が 5 で固定のため、`SetLimit` を明示する意味は薄いが、内部 fan-out（2-2〜2-4）には明示する。
5. **HackerNews の story 取得と comment 取得を別々の `SetLimit` に分けた理由（案1 採用）**：両方とも同一 host（HN API）へ飛ぶため、共有 semaphore で合計同時接続数を一元管理する案（案2）も検討したが、共有 semaphore をどこに持たせるかという設計課題が増え、KISS に反する。別々に保守的な値（5）を設定する方が実装が単純。
6. **3-1（json/mp3 PUT）を見送った理由**：2 並行のみで効果が小さく、`EpisodeWriter.Write` の `@ensure` 契約（「json 先行完了保証、途中失敗は成功にしない」）を崩すコストに見合わない。
7. **3-2（WAV 尺計算ループ）を見送った理由**：`parseWAV` は header 解析のみで処理が軽く、`Goroutine` 生成・scheduling の overhead が処理時間を上回る可能性が高い（実測はしていない推測）。

## 3. Rejected

1. **`sync.WaitGroup` + 手動 slice index 書き込み** — semaphore・error 集約を自前実装する必要があり、`errgroup` と本質的に同じことをより多いコード量で行うだけ。KISS に反する。
2. **`channel` 手動 fan-in** — 今回の要件（並行 fetch → 結果集約）には過剰。順序保証も自前実装が必要になる。
3. **worker pool（固定数 worker + task queue）** — concurrency 上限を厳密に固定したい場合の選択肢だが、`errgroup.SetLimit(n)` で同じことができるため不要。
4. **HackerNews story 取得での chunk 分割 fan-out（早期終了の一部維持）** — id 列を N 件ずつの chunk に区切り、20 件揃った時点で以降の chunk を取りやめる案。chunk 境界判定ロジックが増え、KISS に反するため見送った。全件 fan-out（本 Decision 採用案）を選んだ。
5. **HackerNews 内で story 取得と comment 取得の同時接続数を共有 semaphore で一元管理する案（案2）** — 共有 semaphore の置き場という新規設計課題が増えるため、案1（別々の `SetLimit`）を採用した。
6. **`EpisodeWriter.Write` の json/mp3 PUT 並行化** — `@ensure` 契約変更のコストが、2 並行のみという効果に見合わない。
7. **`ProduceEpisode.Run` の WAV 尺計算ループの並行化** — goroutine overhead が処理時間を上回る可能性が高いと判断し、見送った。
