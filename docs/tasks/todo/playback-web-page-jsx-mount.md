## 1. Summary

この Issue では、`apps/playback/web/src/pages/episode-list-page.ts` を JSX へ書き換え、`apps/playback/web/main.ts` の mount 処理を React の `createRoot` へ変更する。

## 2. Context

1. `playback-web-feature-component-jsx.md`（Feature Component 群の JSX 化）の完了が前提。
2. `docs/decisions/2026-08-20T19-29-21-playback-web-layer-layout.md` により、mount 先は `<main id="app">`、`class` 属性を書かない契約が既に確定している。React 導入後もこの mount 先要素自体は変更しない。

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md` — React 採用の決定。
2. `docs/decisions/2026-08-20T19-29-21-playback-web-layer-layout.md` — mount 先要素（`<main id="app">`）の契約。
3. architecture — `frontend/page-route.md`（Page/Route の責務）。
4. `apps/playback/web/src/pages/episode-list-page.ts` — 書き換え対象。
5. `apps/playback/web/main.ts` — mount 処理の書き換え対象。
6. `apps/playback/web/index.html` — mount 先 DOM 要素の正本。

## 4. Scope

### In Scope

1. `episode-list-page.ts` を `.tsx` へ書き換え、Feature Component 群を組み合わせる JSX を返す。
2. `main.ts` の mount 処理を `ReactDOM.createRoot(...).render(...)` へ変更する。

### Out of Scope

1. Feature/Primitive Component・ViewModel 自体の実装変更（別 Issue で完了済みを前提とする）。
2. `index.html` の mount 先要素（`<main id="app">`）自体の変更。

## 5. Contract

1. mount 先 DOM 要素は既存の `<main id="app">` のまま変更しない（`class` 属性を追加しない）。
2. Page が組み合わせる Feature Component の構成（一覧・再生・原稿表示）は既存の画面構成と同一に保つ。

## 6. Constraints

1. `docs/decisions/2026-08-25T18-42-00-chore-playback-worker-web-layer.md` の層違反検知（dependency-cruiser）を通過すること。
2. Pico.css classless の semantic tag 前提（`div` 主体の構造にしない）を維持する。

## 7. Acceptance Criteria

- [ ] AC-1: `episode-list-page.tsx` が Feature Component 群を組み合わせた JSX を返す。
- [ ] AC-2: `main.ts` が `createRoot` で `<main id="app">` へ mount する。
- [ ] AC-3: `npm run dev` で一覧・再生・原稿表示が画面上で確認できる。
- [ ] AC-4: `lint:layers`（dependency-cruiser）が通過する。

## 8. Verification

```bash
cd apps/playback && npm run typecheck
cd apps/playback && npm run test:unit
cd apps/playback && npm run lint:layers
cd apps/playback && npm run dev
```

`npm run dev` 起動後、ブラウザで一覧→詳細→再生の画面遷移を manual 確認する。

## 9. Dependencies

- blocked by: `playback-web-feature-component-jsx.md`

## 10. Risks

1. `createRoot` への切替時、既存 vanilla 実装が担っていた URL 同期（`lib/location-hash.ts`）や再 mount 時の state 初期化順序が変わる risk がある。既存の画面遷移（一覧⇔詳細）が同じ挙動を保つことを manual 確認で担保する。

## 11. Notes

1. 本 Issue の完了をもって、web 側の React 化（C3〜C6）が完了する。
2. ViewModel hook 化時の page 一時橋（`mount-episode-list-view-model`）は本 Issue で削除する（decision: `docs/decisions/2026-08-26T19-29-00-feature-playback-web-view-model-react-hooks.md`）。
3. worker 系（route / entry）と web 系は独立ではない。Hono `AppType` と RPC 境界の正は `docs/decisions/2026-08-26T19-27-00-feature-playback-web-view-model-react-hooks.md` / `2026-08-26T19-28-00-...` を参照する。
