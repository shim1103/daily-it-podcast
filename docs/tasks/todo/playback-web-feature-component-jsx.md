## 1. Summary

この Issue では、`apps/playback/web/src/components/feature/` 配下の5 component（`episode-list-item`, `episode-list`, `episode-manuscript`, `episode-player`, `episode-topic`）を JSX 関数コンポーネントへ書き換える。

## 2. Context

1. `playback-web-view-model-react-hooks.md`（ViewModel の hook 化）と `playback-web-primitive-component-jsx.md`（Primitive の JSX 化）の完了が前提。
2. 5 component は互いに依存し合わない独立した Feature Component であり、この Issue 内で並行して書き換えられる。
3. `episode-player.ts` は `<audio controls>` 等の標準 HTML 要素を直接記述する契約（`frontend/feature-component.md` §2-3）を持つ。React 化後もこの契約は変わらない。

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md` — React 採用の決定。
2. architecture — `frontend/feature-component.md`（Ring 対応、render error、Error Boundary、import ルール）。
3. `apps/playback/web/src/components/feature/*.ts` — 書き換え対象5 file。
4. `apps/playback/web/src/view-models/episode-list-view-model.ts`（hook 化後）— Feature Component が依存する hook の型。
5. `apps/playback/web/src/components/primitive/labeled-text.tsx`（JSX 化後）— Feature Component が利用する Primitive Component。

## 4. Scope

### In Scope

1. `episode-list-item.ts`, `episode-list.ts`, `episode-manuscript.ts`, `episode-player.ts`, `episode-topic.ts` を `.tsx` へ書き換える。
2. 各 component の render error 方針を `frontend/feature-component.md` §5（Error Boundary）に合わせる。
3. 対応する5本の `*.sociable_unit.test.ts` を JSX レンダリング検証へ更新する。

### Out of Scope

1. Page 層の JSX 化・mount 処理（別 Issue、`playback-web-page-jsx-mount` が担う）。
2. ViewModel・Primitive Component 自体の実装変更（別 Issue で完了済みを前提とする）。

## 5. Contract

1. 各 component が受け取る props の意味（domain 値の種類）は既存実装と同一に保つ。
2. `episode-player.ts` の `<audio controls>` 直接記述の契約を維持する（External Dependencies 層でラップしない）。

## 6. Constraints

1. `frontend/feature-component.md` §6 の import ルール（API Client・External Dependencies 層・境界共有型を直接 import しない）を守る。
2. throw しない、logging しない（`feature-component.md` §4）。
3. render 中の throw は Error Boundary で fallback UI に変換する契約に従う。

## 7. Acceptance Criteria

- [ ] AC-1: 5 component すべてが `.tsx` として存在し、JSX を返す。
- [ ] AC-2: 各 component の既存 test が JSX レンダリング結果に対する検証として書き換わっている。
- [ ] AC-3: `episode-player.tsx` が `<audio controls>` を直接描画する。
- [ ] AC-4: `.ts` 版の5 file が repo に存在しない。

## 8. Verification

```bash
cd apps/playback && npm run typecheck
cd apps/playback && npm run test:unit
cd apps/playback && npm run lint
cd apps/playback && npm run lint:layers
```

## 9. Dependencies

- blocked by: `playback-web-view-model-react-hooks.md`, `playback-web-primitive-component-jsx.md`
- blocks: `playback-web-page-jsx-mount.md`

## 10. Risks

1. 5 component を1 Issue にまとめているため、途中で一部だけ完了した中間状態が生じうる。component 間に依存が無いため、部分完了でも他の component の作業を妨げない。
