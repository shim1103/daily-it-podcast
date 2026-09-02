# feat(generator): 3 情報源 Adapter を composite へ結線し Broad Integration と docs を更新する

## 1. Summary

このIssueでは HackerNews / Lobsters / ITmedia の3 Adapter を Composition Root の composite `ItemSource` へ結線済みの状態を検証可能にし、`apps/generator/test/produce_episode_broad_integration_test.go` の3つの `t.Skip` を外して3源の upstream double で緑にする。完了後、`go test ./...` の `test` package が全緑になり、DESIGN / README / lane が3源構成と一致する。

## 2. Context

- 情報源を GetXAPI から3公式源へ入替える Decision（`docs/decisions/2026-09-02T14-41-00-feature-hackernews-api-adapter.md`）の実装 task の最後の1本。
- scope-split の A で結線コードは既に書かれている: `internal/composition/produce_episode.go` の `newCompositeItemSource(newHackerNewsItemSource(...), newLobstersItemSource(...), newITmediaItemSource(...))`。`internal/composition/{hackernews,lobsters,itmedia}.go` の 5 行結線も存在。
- `apps/generator/test/integration_support_test.go` は前作業で GetXAPI harness を除去し `compositeItemSource{}`（空）へ差し替え済み。`// C: 3 情報源（HackerNews/Lobsters/ITmedia）の upstream double はここで結線する` のマーカーがある。
- `produce_episode_broad_integration_test.go` の3本が `t.Skip("C: ...")` になっている: `uploadsEpisodeArtifactsOnce_whenAllProductionAdaptersSucceed` / `writesNothing_whenTextWriterFails` / `writesNothing_whenSynthesizeFails`。`returnsNoSourceItemsWithoutDownstreamCalls_whenFetchReturnsEmpty` は空 composite で正常に緑なので触らない。

## 3. Canonical Sources

- `docs/decisions/2026-09-02T14-41-00-feature-hackernews-api-adapter.md` — 3源構成、composite で束ねる、Application は源個数を知らない。
- `docs/decisions/2026-09-02T15-27-00-feature-hackernews-api-adapter.md` — 失敗伝播、fallback しない（1源 error で composite 全体 error）。
- `internal/composition/item_source.go` — `compositeItemSource` の契約（登録順 concat / error 透過 / 非 nil 空 slice）。変更しない。
- `apps/generator/test/integration_support_test.go` — Broad harness。`// C:` マーカー位置に3源 double を結線する。
- `apps/generator/test/produce_episode_broad_integration_test.go` — Skip を外す対象。既存の downstream assertion（`assertBroadDownstreamCalls` 等）は変えない。
- `DESIGN.md` §3 外部 I/O 表 / `README.md` 技術選定 / `docs/tasks/todo/generator-lane.md` — 3源構成に更新済み。この task では最終確認のみ。
- test 分類・命名は `skills/1:terms/testing-strategy/SKILL.md`。Broad の最小化は該当 Decision（`2026-08-30T15-42-00` / `2026-08-30T16-23-00`）を参照。

## 4. Scope

### In Scope

- `integration_support_test.go` の `// C:` マーカー位置に、HackerNews / Lobsters / ITmedia の upstream を TLS redirect で double する結線を追加。3源が「1件以上の SourceItem を返す」success 経路と、Broad が必要とする失敗経路（`emptyGetXAPI` 相当の「0件」や個別源の失敗）を harness config で切り替えられるようにする。
- `produce_episode_broad_integration_test.go` の3つの `t.Skip` を外し、既存 assertion を満たす。
- 必要なら Broad harness の config struct（旧 `broadProduceEpisodeConfig`）を3源用に整理。
- DESIGN / README / lane が3源構成と一致していることの最終確認（差分があれば直す）。

### Out of Scope

- 3 Adapter の `List` 実装（各 `generator-<source>-adapter.md`）。この task は Adapter が実装済みである前提。
- `compositeItemSource` の契約変更（sort / dedup / 部分成功。lane D 項目）。
- delivery の kind 分類 type switch（現ブランチに無い。merge 時作業。lane D）。
- System suite / 本番 produce workflow の3源対応（別途）。

## 5. Contract

- `internal/composition/produce_episode.go` の `newProduceEpisode` は3源を登録順（HackerNews → Lobsters → ITmedia）で composite へ渡す。この結線コードは A で確定済み。変更しない。
- Broad Integration の観測は「合成 postcondition のみ」（既存 Decision `2026-08-30T15-42-00` の最小化）: 成功時に json+wav の書込1組、途中失敗時に書込なし、代表 call 回数。HTTP status 列 / schema 全 field / 個別源の request path 枝は assert しない（下位 Scope 所有）。
- `t.Skip` を外した3本の assertion（`assertBroadDownstreamCalls` の期待値等）は現状のまま満たす。テストの意図（成功時アップロード1組 / TextWriter 失敗時書込なし / Synthesize 失敗時書込なし）を変えない。

## 6. Constraints

- Broad は真外部のみ double（httptest TLS redirect）。production Adapter 型・順序は composition と同型に保つ（`2026-08-30T15-42-00`）。
- Broad の fixture は「valid な代表が通る」最小。下位 Scope が持つ range 網羅を Broad に持ち込まない。
- 3源のうち1源でも upstream が error を返す設定にしたとき、composite 契約どおり composite 全体が error になり `ProduceEpisode.Run` が中止することを1 case で確認してよい（`docs/decisions/2026-09-02T15-27-00` §2 の fallback しない挙動）。ただし個別源の request path 枝までは踏み込まない。

## 7. Acceptance Criteria

- [ ] `integration_support_test.go` に3源の upstream double 結線があり、`// C:` マーカーの TODO コメントが解消されている。
- [ ] `produce_episode_broad_integration_test.go` の3本から `t.Skip` が外れ、既存 assertion を満たして緑。
- [ ] `returnsNoSourceItemsWithoutDownstreamCalls_whenFetchReturnsEmpty` が引き続き緑（3源すべてが0件を返す設定で `no_source_items`）。
- [ ] `cd apps/generator && go test ./...` が全 package `ok`、FAIL 0、想定外の SKIP 0（3源 Adapter の Sociable Unit / Narrow は各 adapter task で緑化済みの前提）。
- [ ] `go build ./...` / `go vet ./...` / `golangci-lint run ./...` / `gofmt -l .` がすべて clean。
- [ ] `DESIGN.md` §3・`README.md`・`docs/tasks/todo/generator-lane.md` が3源構成（HackerNews・Lobsters・ITmedia、facade なし）と一致。
- [ ] Broad が Integration gate（secret なし Narrow + Broad）で走る。

## 8. Verification

```bash
cd apps/generator
go build ./...
go vet ./...
go test ./... 2>&1 | tail -30
golangci-lint run ./...
gofmt -l .
./scripts/test-integration.sh   # secret なし Narrow + Broad
```

- `go test ./...` の `test` package が `ok`。3つの `TestProduceEpisodeBroadIntegration_*`（旧 Skip）が PASS。
- `./scripts/test-integration.sh` が緑（gate 契約: `docs/decisions/2026-08-30T11-56-00`）。

## 9. Dependencies

- Blocked by: `generator-hackernews-adapter.md` / `generator-lobsters-adapter.md` / `generator-itmedia-adapter.md`。3 Adapter の `List` が実装済みで各 package が緑でないと、Broad harness が `List` の `panic` を踏む。
- Blocks: なし（この branch の source 入替え完了）。

## 10. Risks

- 3源の upstream double を1つの harness で同時に立てると、TLS redirect の host 振り分け（`hacker-news.firebaseio.com` / `lobste.rs` / `rss.itmedia.co.jp`）を `DialTLSContext` で3宛先に分岐する必要がある。既存の `integrationTLSRoutes` の map パターンを踏襲する。
- Broad が3源すべての success double を要求すると fixture が肥大化する。各源「1件返す」最小の JSON / XML に留める（`2026-08-30T16-23-00` の最小化）。
- `returnsNoSourceItems` の緑を壊さないこと。3源すべて0件の設定を明示的に持つ。

## 11. Notes

- 前作業で `broadProduceEpisodeConfig{emptyGetXAPI: true}` は `broadProduceEpisodeConfig{}` に置換済み。3源用の config field 名（例: `emptySources bool` / `hackernewsFail bool`）は実装者判断でよいが、既存の他 vendor（`cursorFail` / `geminiFailAt`）の命名に揃える。
- この task の完了で、この branch の scope-split C は全消化。残りは lane D（sort / 別媒体 / comment 再帰 / web_fetch 実測 / merge 時の delivery type switch）。
