## 1. Summary

このIssueでは、`apps/playback/web/src/view-models/use-hash-sync.ts` の外部ストア購読部を React 標準の `useSyncExternalStore` ベースへ移行し、手書きの hashchange listener 登録・`lastSyncedHashRef`・`onHashSelectRef`（stale closure 対策）を除去する。routing library を入れない自前実装のまま、購読の仕組みだけを React 公式手段へ寄せて非自明さを下げる。

## 2. Context

1. `docs/decisions/2026-08-27T19-20-30-feature-playback-web-page-jsx-mount.md` で、hash 同期は routing library を入れず `lib/location-hash.ts` + `useHashSync` で行うと確定済み。同 decision §1-3 が「外部ストア購読部は `useSyncExternalStore` へ寄せる。現状の `useEffect` 手書きは暫定」と定めている。
2. 現状の `useHashSync` は `useEffect` + `window.addEventListener("hashchange", ...)` を手書きし、listener を貼り替えず最新の `onHashSelect` を呼ぶために `onHashSelectRef` を、書き戻し無限ループ抑止のために `lastSyncedHashRef` を持つ。いずれも `useSyncExternalStore` が返す最新値を使えば簡略化できる。
3. 本Issueは `feature/playback-web-page-jsx-mount` の PR（`episode-list-page.ts` の JSX 化）が merge されてから着手する。JSX 化Issueの Scope 外として意図的に分離された follow-up である。

## 3. Canonical Sources

1. `docs/decisions/2026-08-27T19-20-30-feature-playback-web-page-jsx-mount.md` — routing library 非採用と `useSyncExternalStore` 寄せの判断。
2. `apps/playback/web/src/view-models/use-hash-sync.ts` — 移行対象の現行実装と公開契約（`@require` / `@ensure` / `@invariant`）。
3. `apps/playback/web/src/lib/location-hash.ts` — hashchange 購読・hash 読み書きの Driven Adapter。signature は変えない。
4. `apps/playback/web/src/pages/episode-list-page.tsx` — `useHashSync` の呼び出し側。`onHashSelect` ハンドラと `selectedId` の算出。
5. React `useSyncExternalStore` の公式ドキュメント — `subscribe` / `getSnapshot` の契約、`getSnapshot` の参照安定性要件。
6. `architecture/frontend/view-model.md` / `architecture/frontend/external-dependencies.md` — hook の置き場と import 制約。
7. test方針は `skills/1:terms/testing-strategy/SKILL.md` を参照する。

## 4. Scope

### In Scope

1. `use-hash-sync.ts` の外部ストア購読部を、`useSyncExternalStore(subscribe, getSnapshot)` を使う形へ書き換える（`lib/location-hash.ts` の `onLocationHashChange` を `subscribe` に、`getLocationHash` を `getSnapshot` に渡すラッパー）。
2. `lastSyncedHashRef` と `onHashSelectRef` を除去する。`useSyncExternalStore` が返す最新 hash と `selectedId` の直接比較で、書き戻し無限ループ抑止と「hash が外部から変わったか」判定を置き換える。
3. `use-hash-sync.sociable_unit.test.ts` の更新。既存の検証観点（selectedId 追従・hashchange→onHashSelect の引数・書き戻し無限ループ無し・unmount で購読解除）を維持する。

### Out of Scope

1. `lib/location-hash.ts` の signature 変更。既存 Driven Adapter をそのまま `subscribe` / `getSnapshot` へ渡す。
2. `episode-list-page.tsx` の `onHashSelect` ハンドラの責務変更、`selectedId` の `undefined`（未初期化）表現の変更。
3. state→hash の書き込み（`setLocationHash`）を担う `useEffect` の除去。読み取り方向のみ `useSyncExternalStore` へ移す。
4. routing library の再検討。

## 5. Contract

1. `useHashSync` の公開 signature `(selectedId: string | null | undefined, onHashSelect: (id: string | null) => void): void` は変更しない。
2. `useHashSync` の `@ensure` の外部挙動（selectedId 変化で hash 追従、hashchange で onHashSelect 呼び出し、undefined の間は hash を書き換えない、書き戻し無限ループ抑止、unmount で購読解除）を保つ。
3. `getSnapshot` は文字列（`#` を除いた hash 値）を返す。

## 6. Constraints

1. `getSnapshot` はオブジェクトや配列を返さない。`useSyncExternalStore` は `getSnapshot()` の戻り値を `Object.is` で前回と比較するため、毎回新しい参照を返すと再 render の無限ループになる。文字列なら値比較で安定する。
2. dependency-cruiser（`lint:layers`）の層違反を出さない。`use-hash-sync.ts` の import は `lib/location-hash.ts` と `react` のみ。

## 7. Acceptance Criteria

- [ ] AC-1: `use-hash-sync.ts` が `useSyncExternalStore` を使って hash の現在値を読む。
- [ ] AC-2: `lastSyncedHashRef` と `onHashSelectRef` が `use-hash-sync.ts` から消える。
- [ ] AC-3: 既存の hash 同期シナリオ（`use-hash-sync.sociable_unit.test.ts` の全ケース、`episode-list-page.sociable_unit.test.ts` の全ケース）が無修正または挙動同値の修正で green。
- [ ] AC-4: `npm run typecheck` / `npm run lint:layers` / `npm run lint` が green。
- [ ] AC-5: `npm run test:unit` を 5 連続実行して flaky ゼロ。

## 8. Verification

```bash
cd apps/playback && npm run typecheck
cd apps/playback && npm run test:unit
cd apps/playback && npm run lint:layers
cd apps/playback && npm run lint
```

`test:unit` は 5 連続で安定 green を確認する。`npm run dev` で一覧⇔詳細の hash 同期（選択で hash が付く、戻る/進むで選択が動く、`#ep-xxxx` 直リンクで該当 episode が展開する）を manual 確認する。

## 9. Dependencies

- blocked by: `feature/playback-web-page-jsx-mount` の PR merge（`use-hash-sync.ts` はその PR で新規追加されるため）。

## 10. Risks

1. `useSyncExternalStore` の `getSnapshot` が毎 render 新参照を返すと再 render 無限ループになる。文字列を返す契約（Constraint 1）で防ぐ。移行時に `getSnapshot` の戻り値型を test で固定する。
2. 書き戻し無限ループ抑止のロジックを `lastSyncedHash` 比較から「最新 hash と selectedId の直接比較」へ変える際、エッジケース（load 完了前の `undefined`、選択解除、外部からの hash クリア）の挙動が微妙にずれる可能性がある。`use-hash-sync` と `episode-list-page` の既存全ケース + 5 連続 flaky チェックで担保する。

## 11. Notes

1. state→hash 書き込みの `useEffect` は残る。`useSyncExternalStore` は外部→React の読み取り方向専用で、書き込み側を hook 標準に載せる仕組みは React に無い（decision §2-5）。
2. この移行で `useHashSync` が「読み: `useSyncExternalStore`、書き: `useEffect`」の非対称構造になるのは意図どおり。双方向同期を1つの標準 hook で表す手段は存在しない。
