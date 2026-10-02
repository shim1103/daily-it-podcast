---
name: HasPair の stem 候補組み立てはメモリ上の逐次分類に留め、errgroup 対象外とする
date: 2026-10-02T07:47:56
branch: refactor/generator-io-fanout
---

## 1. Decision

1. `CompletedEpisodeLookup.HasPair` 内の `pairStemCandidates`（List 済み object key からの json∩mp3 stem 抽出）は **逐次のメモリ処理**のままにする。`errgroup` / goroutine を使わない。
2. 並行化の対象は **stem への GET**（`matchDateOnStems` + `maxConcurrentGets`）に限定する。公開契約の「同時実行してよい」は GET を指し、候補組み立てを含まない。
3. 契約（Port / Adapter `@ensure`）の変更は不要。既存文言は GET 上限のみを述べている。

## 2. Reason

1. `pairStemCandidates` の入力は、既に完了した `listObjectKeys`（HTTP List）の結果である。ここでの仕事は suffix 判定と集合交差だけで、**HTTP でもローカル OS I/O でもない**。
2. I/O-bound fan-out（[[2026-09-23T17-21-29-refactor-generator-go-performance]]）の対象は外部待ちの独立 fetch。CPU の短い走査に goroutine を足すと、scheduling コストと共有 map の同期が増えるだけで実利が薄い（KISS）。
3. mp3 stem の set を先に作り json と突き合わせる現形は、1 パス分類 + 1 パス交差で足り、並列化してもアルゴリズム上の利点が小さい。

## 3. Rejected

1. **`pairStemCandidates` を errgroup で並行化する案** — I/O がなく、共有 map か分割マージが要る。安全側の concurrency 上限を足す意味も薄い。
2. **契約を「key 分類も同時実行してよい」へ広げる案** — 実装しない方針と矛盾し、caller が頼る観測も増えない。
3. **List と GET と候補組み立てを1つの巨大 fan-out にまとめる案** — pagination 依存の List と独立 GET を混ぜ、既存の「List は並行化しない」判断と衝突する。
