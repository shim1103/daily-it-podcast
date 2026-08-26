## 1. Summary

この Issue では、`apps/playback/web/src/components/primitive/labeled-text.ts` を JSX 関数コンポーネントへ書き換える。

## 2. Context

1. `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md` で React 導入が決定済み。
2. A 区分で `tsconfig.json` に `"jsx": "react-jsx"`、`vite.config.ts` に `@vitejs/plugin-react` が既に設定済み。
3. `labeled-text.ts` は他 component への依存を持たない Primitive Component であり、この Issue は他の JSX 化 Issue と依存関係が無い。

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md` — React 採用の決定。
2. architecture — `frontend/primitive-component.md`（Ring 対応、何を置く/置かない、import ルール）。
3. `apps/playback/web/src/components/primitive/labeled-text.ts` — 書き換え対象。

## 4. Scope

### In Scope

1. `labeled-text.ts` を `.tsx` へ書き換え、`HTMLElement` を返す関数から JSX を返す関数コンポーネントへ変更する。
2. 対応する `labeled-text.sociable_unit.test.ts` を React Testing Library 相当の検証へ更新する。

### Out of Scope

1. Feature Component の JSX 化（別 Issue）。
2. 呼び出し元（Feature Component）の更新（別 Issue、`playback-web-feature-component-jsx` が担う）。

## 5. Contract

1. component の props interface（受け取る値の型）は既存の意味を維持する。呼び出し元の使用方法が変わらないこと。

## 6. Constraints

1. `frontend/primitive-component.md` の責務境界（domain knowledge zero で再利用可能な部品）を維持する。
2. throw しない、logging しない（`primitive-component.md` の throw/logging 方針を参照）。

## 7. Acceptance Criteria

- [ ] AC-1: `labeled-text.tsx` が JSX を返す関数コンポーネントとして存在する。
- [ ] AC-2: 既存 test が JSX レンダリング結果に対して同等の検証を行う。
- [ ] AC-3: `.ts` 版の `labeled-text.ts` が repo に存在しない。

## 8. Verification

```bash
cd apps/playback && npm run typecheck
cd apps/playback && npm run test:unit
cd apps/playback && npm run lint
```

## 9. Dependencies

- blocks: `playback-web-feature-component-jsx`（Feature Component がこの Primitive Component を利用する場合）

## 10. Risks

意味のある risk なし。
