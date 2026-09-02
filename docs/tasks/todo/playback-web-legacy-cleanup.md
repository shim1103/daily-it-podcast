# playback web 旧 artifact 撤去

## 1. Summary

新 UI / ViewModel への移行完了後、legacy rename 済み file と未参照 CSS・test を安全に削除する。

## 2. Context

`playback-web-ui-rewrite` で旧 component は rename または未参照化される。本 Issue で delete と import 整理を行う。

## 3. Canonical Sources

- 新契約の正 — `apps/playback/web/src/view-models/use-episode-list-page.ts`、`episode-row.tsx` 他
- 直交 Decision — `docs/decisions/2026-09-02T15-00-00-feature-playback-list-episodes-audio-ref-playback-web-selection-playback-orthogonality.md`

## 4. Scope

### In Scope

- `episode-list-view-model.ts` と test の削除
- `*.legacy.tsx` および未参照旧 component / CSS の削除
- `use-hash-sync.ts` 等、置換済み hook の削除（統合完了時）
- `grep` による旧 symbol ゼロ確認

### Out of Scope

- 新機能追加
- 視覚デザインの再設計

## 5. Contract

- 削除後も `npm run test:unit` と `lint:layers` が pass すること

## 6. Constraints

- UI rewrite の手動確認完了後にのみ着手する

## 7. Acceptance Criteria

- [ ] AC-1: `episode-list-view-model` への import が repo 内に無い
- [ ] AC-2: legacy suffix file が残っていない
- [ ] AC-3: `npm run test:unit` / `lint:layers` / 既存 E2E が pass する

## 8. Verification

```bash
cd apps/playback && npm run test:unit && npm run lint:layers
rg 'episode-list-view-model|episode-list-item' apps/playback/web/src
```

## 9. Dependencies

- blocked by: `docs/tasks/todo/playback-web-ui-rewrite.md`

## 10. Risks

- 孤立 CSS が残り bundle に載り続ける — delete 前に参照 grep する

## 11. Notes


