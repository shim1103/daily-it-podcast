---
name: episode 一覧 page の JSX 化と hash 同期 hook 抽出
date: 2026-08-27T19:20:30
session_id: none
branch: feature/playback-web-page-jsx-mount
prev: なし
---

## 1. Summary

issue-manager で `apps/playback/web/src/pages/episode-list-page.ts` を JSX 関数コンポーネント `EpisodeListPage` へ書き換え、`main.ts` の mount を `ReactDOM.createRoot().render()` へ変更した。ViewModel hook 化時の一時橋2本を削除し、web 側 React 化（C3〜C6）を完了させた。code-reviewer の査読（blocker なし、should-fix 4件）を反映し、続けて shim の指示で hash 同期ロジックを `useHashSync` custom hook へ抽出、`EpisodeList` を `React.memo` でラップした。routing library 非採用と `useSyncExternalStore` 寄せの判断を decision に記録し、follow-up Issue を作成した。localhost:3000 での mount 成功を shim が確認し、issue file を削除、4 commit を push した。

## 2. Changes

1. `episode-list-page.ts` を `.tsx` 化し、`useEpisodeListViewModel` hook を直接使う `EpisodeListPage({ apiClient, baseUrl })` にした。一時橋 `mount-episode-list.ts` / `mount-episode-list-view-model.ts` を削除。`main.ts` を `app.replaceChildren` から `createRoot(app).render(createElement(EpisodeListPage, ...))` へ変更（mount 先 `<main id="app">` は不変）。
2. code-reviewer 査読の should-fix を反映: `selectRef` 撤去（ViewModel が `select` を `useCallback` 安定化済みのため）、mount 効果に cancel フラグ追加（in-flight 中の unmount 競合）、page test の `waitFor(() => toBeNull())` 二重 assert 解消、hash リセットを `beforeEach` 一本化。
3. hash ↔ selectedEpisodeId の双方向同期（`useEffect` 3本 + `useRef` 3個）を `view-models/use-hash-sync.ts` へ抽出。page は 107→77行、`useEffect` 3→1、`useRef` 3→1。「load 完了まで state→hash 同期を止める」現挙動は `selectedId` に `undefined`（未初期化）を渡す2状態表現で保った（`null`=選択なし、`undefined`=同期保留）。`initializedRef` ゲートは load ライフサイクルとの結合として page に残置。
4. `EpisodeList` を `React.memo` でラップ。`onSelect` の `useCallback` 安定化が子の再 render 抑止に効くようにした。`components/` に memo 前例ゼロのため「関数式 + memo」形を採用。
5. decision `2026-08-27T19-20-30`: URL hash 同期は React Router を入れず `lib/location-hash` + `useHashSync` で行い、外部ストア購読は `useSyncExternalStore` へ寄せる（現状の `useEffect` 手書きは暫定）。follow-up Issue `docs/tasks/todo/playback-web-use-sync-external-store.md` を作成。
6. shim との対話で React / JS / ブラウザの基礎（イベントループ、マイクロ/マクロタスク、クロージャと stale closure、`useRef`/`useState` の再 render 挙動、render と mount と commit の違い、`useEffect` の deps 比較と実行タイミング、`React.memo`/`useCallback`/`useMemo`、SSR vs SPA と SEO、`useSyncExternalStore`）を議論した。
7. Verification（typecheck / test:unit 37 files・211 tests / lint / lint:layers）全て pass、`web/src/pages` `web/src/view-models` カバレッジ 100/100/100/100 を manager が再実行で audit。4 commit（feat・refactor・perf・docs）を push。

7. pr-completion flow で 4 commit（feat・refactor・perf・docs）+ log commit を push し、PR [#73](https://github.com/shim1103/daily-it-podcast/pull/73)（base: develop）を `gh pr` 経由で作成した。`develop` 先行の PR #72 との `docs/lessons/index.md` conflict を両側保持（番号 219〜223 へ振り直し）で解消。CI（static-and-unit / integration）全 pass、mergeStateStatus CLEAN を確認。

### Commits

- `7693a73`
- `36da90a`
- `3692b17`
- `f55208f`
- `e5730bd`
- `84cd6bd`
