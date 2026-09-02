# playback web UI 差し替え（Row / Entry / Audio + page wire）

## 1. Summary

`EpisodeRow` / `EpisodeEntry` / `AudioControls` と `useEpisodeListPage` で一覧 page を組み直し、blocking error を全画面 1 UI に統一する。

## 2. Context

ViewModel 本実装（`playback-web-view-models`）完了後に着手する。旧 `episode-list` / `episode-list-item` / `episode-player` は参照を外すが、delete は次 Issue。

## 3. Canonical Sources

- UI 責務配置 — `docs/decisions/2026-09-02T15-00-00-feature-playback-list-episodes-audio-ref-playback-web-selection-playback-orthogonality.md`
- 1 page 前提 — `docs/decisions/2026-08-25T05-10-48-feature-playback-ui-structure`
- 視覚言語 — `docs/decisions/2026-08-28T19-20-01-docs-playback-list-page-design`
- component 契約 — `apps/playback/web/src/components/feature/episode-row.tsx` 他
- page compose — `apps/playback/web/src/view-models/use-episode-list-page.ts`
- Feature 層 — `architecture/frontend/feature-component.md`

## 4. Scope

### In Scope

- `episode-list-page.tsx` を `useEpisodeListPage` へ wire
- 一覧を `EpisodeRow` + 条件付き `EpisodeEntry` + `AudioControls` で組み立て
- blocking error の全画面表示（list / detail 分離なし）
- 旧 component の **rename**（`.legacy.tsx` 等）または未参照化
- 対応 page / feature の SU test 更新

### Out of Scope

- 旧 file の delete（`playback-web-legacy-cleanup`）
- seek UI の新規追加
- E2E の大幅な scenario 追加（既存が壊れたら修正は In Scope）

## 5. Contract

- Row は `isSelected` / `isPlayed` / `isPlaying` を props で受け取り、自身では derive しない
- Entry は manuscript のみ。`title` / `date` を重ねない
- AudioControls は再生中 episode 用。選択と独立に表示できること

## 6. Constraints

- 旧 component file は delete しない（rename のみ）
- 物語・色の Decision 本文を code comment へ写さない

## 7. Acceptance Criteria

- [ ] AC-1: 一覧表示・選択展開・再生・停止が手動で確認できる
- [ ] AC-2: 原稿を閉じても再生が継続する
- [ ] AC-3: catalog / audio 失敗時に同一 Error UI が表示される
- [ ] AC-4: `episode-list-page.sociable_unit.test.ts` が pass する
- [ ] AC-5: `npm run test:unit` が pass する

## 8. Verification

```bash
cd apps/playback && npm run test:unit && npm run dev
```

手動: select / deselect / play / switch episode / deselect-while-playing

## 9. Dependencies

- blocked by: `docs/tasks/todo/playback-web-view-models.md`

## 10. Risks

- 旧 CSS class 名への依存が残り、見た目が一時的に崩れる — 視覚 Decision 参照で段階的に統合する

## 11. Notes

- 旧 `episode-list-item` の play 導線は Row へ集約する
