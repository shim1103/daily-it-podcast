# playback web ViewModel 本実装

## 1. Summary

playback web の catalog / selection / hash-sync / playback hook を、A 契約 stub から本実装へ置き換える。B で固定した selection と playback の直交を、ViewModel 層で守る。

## 2. Context

A artifact（型・stub・SU test）と B Decision（直交・edge case）は本 branch で固定済み。page と Feature component はまだ旧 ViewModel を参照している。

## 3. Canonical Sources

- 直交と edge case — `docs/decisions/2026-09-02T15-00-00-feature-playback-list-episodes-audio-ref-playback-web-selection-playback-orthogonality.md`
- hash 同期の React 方針 — `docs/decisions/2026-08-27T19-20-30-feature-playback-web-page-jsx-mount`
- hook / derive 契約 — `apps/playback/web/src/view-models/`（`playback-state.ts`、各 `use-episode-*.ts`）
- hash adapter — `apps/playback/web/src/lib/hash-selection-adapter.ts`
- 層責務 — `architecture/frontend/view-model.md`
- test 方針 — `testing-strategy` SKILL

## 4. Scope

### In Scope

- `useEpisodeCatalog` — `listEpisodes` fetch と cache
- `useEpisodeSelection` — select / deselect / toggle
- `useHashSelectionSync` — `HashSelectionAdapter` 経由の双方向同期（`use-hash-sync` 統合または置換）
- `useEpisodePlayback` — audio lifecycle、`play` / `stop`、`playbackPhase` 更新
- `useEpisodeListPage` — compose、derive、`blockingError` 判定
- 対応 `*.sociable_unit.test.ts` を stub から振る舞い test へ拡張

### Out of Scope

- page / Feature component の wire（`playback-web-ui-rewrite`）
- 旧 `episode-list-view-model` の削除（`playback-web-legacy-cleanup`）
- 視覚デザイン・CSS の本実装

## 5. Contract

- B Decision の不変条件を満たす（`Select`/`Deselect` は playback を変えない、`Play`/`Stop` は selection を変えない）
- A の export 型・関数名を変えない（変更が要る場合は先に A を更新する）
- catalog load 完了前は hash 同期を `undefined`（同期保留）で開始しない

## 6. Constraints

- `throw` しない。失敗は state / Result で表現する（`view-model.md`）
- 旧 `episode-list-view-model.ts` は本 Issue では削除しない

## 7. Acceptance Criteria

- [ ] AC-1: `npm run test:unit`（apps/playback）が pass する
- [ ] AC-2: `npm run lint:layers` が pass する
- [ ] AC-3: 別 episode へ `Play` したとき、直前 audio が stop し reset される（test で検証）
- [ ] AC-4: `Deselect` 後も `playedEpisodeId` が維持される（test で検証）
- [ ] AC-5: invalid `selectedEpisodeId` で `deriveBlockingError` が non-null になる（test で検証）

## 8. Verification

```bash
cd apps/playback && npm run test:unit && npm run lint:layers
```

## 9. Dependencies

- なし（A/B 完了が前提。本 PR で A/B は固定済み）

## 10. Risks

- `use-hash-sync` と `use-hash-selection-sync` の統合で race が再発する risk — 既存 hash 同期 test の観点を移行時に維持する

## 11. Notes

- 旧 `use-hash-sync.ts` は統合後 deprecated とし、C3 で削除する
