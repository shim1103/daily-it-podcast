## 1. Summary

Composition で `ProduceEpisode` production graph を結線する。Application は情報源個数を知らず、composite `ItemSource` 経由で Fetch する。`ProduceEpisodeFactory` を実装する。

## 2. Context

1. `FetchSourceItems` は `ItemSource.List` を 1 回呼ぶ。情報源 merge は Composition / Infrastructure 側である（`docs/decisions/2026-08-19T13-25-20-refactor-generator-source-port.md`）。
2. 現行 `NewProduceEpisode()` は GetXAPI を直接結線している。
3. `ProduceEpisode.Run` 本体は D。本 Issue は factory 結線のみ。

## 3. Canonical Sources

1. `docs/decisions/2026-08-19T13-25-20-refactor-generator-source-port.md`
2. `docs/decisions/2026-08-29T14-14-00-docs-produce-episode-run-spec-fetch-zero-source-items.md`
3. `apps/generator/internal/composition/produce_episode.go`
4. `apps/generator/internal/composition/produce_episode_factory.go`
5. `apps/generator/internal/application/produce_episode.go`（ctor のみ。Run は触らない）
6. `DESIGN.md` — Generator 層表・ItemSource

## 4. Scope

### In Scope

1. composite `ItemSource`（登録順に各 `List` 結果を concat。現状 GetXAPI 1 本でも composite 経由でよい）。
2. `ProduceEpisodeFactory(config.Config)` の実装（検証済み Config から Adapter を結線）。
3. 既存 production Adapter ctor への配線更新。
4. composition sociable test（factory 非 nil、graph compile）。

### Out of Scope

1. `ProduceEpisode.Run` orchestration（D）。
2. M1 HTTP / `secrettransport` 除去（済。本 Issue は現行 tree の結線慣習に倣う）。
3. 第 2 情報源 Adapter 実装（D）。
4. composite の dedup / sort 高度化（D）。
5. Brief 文案・limits 数値チューニング（D）。

## 5. Contract

1. Application は監視対象一覧・情報源種類・個数を知らない。
2. composite merge は登録順 concat のみ（追加规则は D）。
3. Run stub の `@ensure` を本 Issue で変更しない。

## 6. Constraints

1. `secrettransport` は target architecture から除去済み。再導入しない。本 Issue は現行 tree の結線慣習に倣う最小差分。
2. credential 値を test / log / Error に出さない。
3. 過去 Decision を変更・削除しない。

## 7. Acceptance Criteria

1. [ ] production graph が composite `ItemSource` 経由で Fetch する。
2. [ ] `ProduceEpisodeFactory` が config から `*ProduceEpisode` を返す。
3. [ ] `ProduceEpisode.Run` は未実装 stub のまま（panic 維持可）。
4. [ ] `./scripts/generator/check-static.sh` と `test-unit.sh` が pass する。

## 8. Verification

```bash
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
cd apps/generator && go test ./internal/composition/... -count=1
```

## 9. Dependencies

1. A artifact（`ProduceEpisode` ctor、`ItemSource` Port、factory type stub）。
2. build 系 Issue（compose-brief / parse / wav）とは並行可。Run 実装は不要。

## 10. Risks

1. composition を広く触ると ProduceEpisode 結線以外の diff が混ざる。結線に scope を限定する。

## 11. Notes

1. GitHub Issue 化は別判断。本 file が達成契約の正。
