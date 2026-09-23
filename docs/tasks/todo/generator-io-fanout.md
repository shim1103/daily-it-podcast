## 1. Summary

このIssueでは `apps/generator` の独立した I/O-bound 逐次処理を `errgroup` による fan-out へ変更する。完了後、対象 5 箇所（情報源統合・HackerNews story 取得・HackerNews comment 取得・Lobsters story detail 取得・`HasPair` の GET 走査）が並行実行される。

## 2. Context

1. `apps/generator` は特別パフォーマンスに困っているわけではないが、KISS を優先しつつ Goroutine 導入余地を洗い出した（session 内の executor agent 調査、[[2026-09-23T17-21-29-refactor-generator-go-performance]]）。
2. 対象 5 箇所はいずれも I/O-bound（外部 HTTP fetch）で、各要素の処理が他要素の結果に依存しない（独立している）ことを確認済み。
3. HackerNews / Lobsters / Publickey / TechCrunch / クラウドWatch のいずれも専用 rate limiter・backoff 実装を持たず、相手側の実際の許容量は不明。安全側に倒し concurrency 上限は 5 とした。
4. `HasPair` の早期終了（一致した時点で return）は、通常運用では「同日ペアが存在しない」ケースがほとんどであり、早期終了の恩恵が薄いと判断し廃止する。
5. `port.SpeechSynthesizer` の尺計算統合は別 Issue（[[generator-speech-duration]]）であり、本 Issue の scope 外。

## 3. Canonical Sources

1. Goroutine 導入方針の Decision: `docs/decisions/2026-09-23T17-21-29-refactor-generator-go-performance.md`
2. 並行処理の実装原則: `skills/1:terms/concurrency`
3. 既に固定済みの契約定数・contract-docs: `apps/generator/internal/infrastructure/hackernews/item_source.go`（`MaxConcurrentFetches`）、`apps/generator/internal/infrastructure/lobsters/item_source.go`（`MaxConcurrentFetches`）、`apps/generator/internal/infrastructure/r2/constants.go`（`maxConcurrentGets`）、`apps/generator/internal/composition/item_source.go`（`compositeItemSource` contract-docs）
4. test方針は `skills/1:terms/testing-strategy/SKILL.md` を参照する。

## 4. Scope

### In Scope

1. `apps/generator/go.mod` へ追加済みの `golang.org/x/sync/errgroup`（現在 `// indirect`）を実際に import し、direct dependency へ変更する。
2. `internal/composition/item_source.go` の `compositeItemSource.List` を `errgroup` による並行呼び出しへ変更する。
3. `internal/infrastructure/hackernews/item_source.go` の `List` 内 story 取得ループを `errgroup.SetLimit(MaxConcurrentFetches)` による fan-out へ変更する。早期終了（`limit` 到達 `break`）を廃止し、id 列全件を対象にする。
4. `internal/infrastructure/hackernews/item_source.go` の `fetchTopLevelComments` を `errgroup.SetLimit(MaxConcurrentFetches)` による fan-out へ変更する。
5. `internal/infrastructure/lobsters/item_source.go` の `List` 内 `fetchStoryDetail` ループを `errgroup.SetLimit(MaxConcurrentFetches)` による fan-out へ変更する。
6. `internal/infrastructure/r2/lookup.go` の `HasPair` 内 `jsonStems` 走査を `errgroup.SetLimit(maxConcurrentGets)` による fan-out へ変更する。早期終了（一致時点で return）を廃止し、全 stem を対象にする。
7. 上記変更に伴う既存 test の改修（並行実行後も既存の期待値・error 条件を満たすことの確認）。

### Out of Scope

1. `port.SpeechSynthesizer` まわりの尺計算統合（別 Issue [[generator-speech-duration]]）。
2. `EpisodeWriter.Write` の json/mp3 PUT 並行化（Decision で見送り済み）。
3. `ProduceEpisode.Run` の WAV 尺計算ループ・`ConcatWAV` 内 `parseWAV` ループの並行化（Decision で見送り済み）。
4. `CompletedEpisodeLookup.listObjectKeys` の pagination（並行化不適と確認済み）。
5. concurrency 上限値（5）のチューニング・実測。

## 5. Contract

1. `compositeItemSource.List`・HackerNews `List`・HackerNews `fetchTopLevelComments`・Lobsters `List`・`HasPair` いずれも、既存の signature（引数・戻り値の型）は変更しない。
2. 上記いずれも、結果の**順序を保証しない**契約へ変更済み（`internal/composition/item_source.go`、`internal/infrastructure/hackernews/item_source.go`、`internal/infrastructure/lobsters/item_source.go` の contract-docs 参照）。
3. いずれかの並行呼び出しが error を返した場合、全体は error を返す（部分成功を返さない）。どの error が返るかは非決定的でよい（`errgroup` の性質上、最初に検出された error）。

## 6. Constraints

1. 新しい定数・型・dir を追加しない。契約として既に固定済みの `MaxConcurrentFetches`（hackernews / lobsters）・`maxConcurrentGets`（r2）をそのまま使う。
2. 各 fan-out 内の goroutine は、書き込み先を要素ごとに独立した slice index に限定するか、`errgroup` の error 集約のみに依拠し、`sync.Mutex` 等の追加同期機構を導入しない（該当箇所はいずれも「書き込み先が競合しない」設計で足りる）。
3. HackerNews の story 取得（fetchTopLevelComments の呼び出し元）と comment 取得（fetchTopLevelComments 内部）は、共有 semaphore を持たせず、それぞれ独立した `errgroup.SetLimit(MaxConcurrentFetches)` で管理する（Decision Rejected §5 参照）。

## 7. Acceptance Criteria

- [ ] `apps/generator/go.mod` で `golang.org/x/sync/errgroup` が `// indirect` ではなく direct dependency になっている。
- [ ] `compositeItemSource.List` が 5 source を並行に呼び出す実装になっている。
- [ ] HackerNews `List` が早期終了なしで id 列全件を対象に fan-out し、`MaxConcurrentFetches` 件まで同時実行する実装になっている。
- [ ] HackerNews `fetchTopLevelComments` が `MaxConcurrentFetches` 件まで同時実行する実装になっている。
- [ ] Lobsters `List` の `fetchStoryDetail` ループが `MaxConcurrentFetches` 件まで同時実行する実装になっている。
- [ ] `HasPair` が早期終了なしで全 `jsonStems` を対象に fan-out し、`maxConcurrentGets` 件まで同時実行する実装になっている。
- [ ] いずれかの並行呼び出しが error を返すケースで、全体が error を返すことを確認する test がある。
- [ ] 既存 test が並行実行後も pass する。

## 8. Verification

```
cd apps/generator && go build ./... && go vet ./... && go test ./...
```

1. race condition の有無を `go test -race ./...` で確認する。
2. 既存の `go test ./...` が pass することを確認する。

## 9. Dependencies

なし。[[generator-speech-duration]] とは独立に実装できる。

## 10. Risks

1. **rate limit超過**：HackerNews / Lobsters API の実際の rate limit は不明。`MaxConcurrentFetches = 5` は安全側の初期値であり、運用後に 429 が頻発する場合は値の見直しが必要（本 Issue の scope 外、別途対応）。
2. **goroutine leak**：`errgroup.WithContext` を使わずに実装すると、error 発生時に残りの goroutine が `ctx` cancel を受け取らず不要な fetch を継続する可能性がある。`errgroup.WithContext(ctx)` を用いて各 fetch へ渡す `ctx` を統一する。

## 11. Notes

1. Rejected 案（`sync.WaitGroup` 手動実装、`channel` 手動 fan-in、worker pool、chunk 分割 fan-out、共有 semaphore 案）は `docs/decisions/2026-09-23T17-21-29-refactor-generator-go-performance.md` を参照。
