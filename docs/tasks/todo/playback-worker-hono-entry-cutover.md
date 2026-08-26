## 1. Summary

この Issue では、`apps/playback/worker/src/worker-entry.ts` の HTTP 入口を `routes/app.ts` の Hono `fetch` へ切り替え、旧自作 router（`fetch.ts`、`match-playback-route.ts`）と対応 test を削除する。

## 2. Context

1. `playback-worker-hono-route-definition.md` で `app.ts` に route 定義（Controller wiring）が完了していることが前提。
2. `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md` で自作 router を Hono へ置き換える決定が確定済み。
3. `apps/playback/web/vite.config.ts` の dev-only middleware（`createDummyBackendMiddleware`）は `match-playback-route.ts` と `composition/root.ts` を import している。`DESIGN.md` §1 の契約により production HTTP 入口（`worker/src/routes/fetch.ts` 相当）は変更しないという既存記述があるが、本 Issue で `fetch.ts` 自体を削除するため、この dev-only middleware も Hono 経由へ追従させる必要がある。

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md` — Hono 導入の決定。
2. `apps/playback/worker/src/worker-entry.ts` — 入口の正本。
3. `apps/playback/worker/src/routes/app.ts` — 切替先の Hono instance（`playback-worker-hono-route-definition.md` 完了後の状態）。
4. `apps/playback/web/vite.config.ts` — dev-only middleware の依存関係。
5. `DESIGN.md` §1 — production HTTP 入口の契約記述（本 Issue 完了後に file path 部分の追従が必要）。

## 4. Scope

### In Scope

1. `worker-entry.ts` の `fetch` export を `app.ts` の Hono instance の `fetch` へ委譲する形に変更する。
2. `routes/fetch.ts`、`routes/match-playback-route.ts` と対応する `*.sociable_unit.test.ts` を削除する。
3. `web/vite.config.ts` の dev-only middleware を Hono app 経由の呼び出しへ更新する。
4. `DESIGN.md` §1 の「production の HTTP 入口（`worker/src/routes/fetch.ts`）は変更しない」という記述を、新しい入口 file 名に合わせて更新する。

### Out of Scope

1. route handler の内部ロジック変更（`playback-worker-hono-route-definition.md` の scope）。
2. web 側の React 化（別 Issue 群）。

## 5. Contract

1. 切替後も `worker-entry.ts` の export interface（`fetch(request, env)` の signature）は変更しない。
2. dev server（`npm run dev`）から一覧・詳細・audio の各 endpoint が引き続き応答すること。

## 6. Constraints

1. 削除対象（`fetch.ts`、`match-playback-route.ts`）に依存する他 file が無いことを確認してから削除する。
2. `requestId`（`crypto.randomUUID()`）によるエラーレスポンスの trace 用 ID 付与など、既存 `fetch.ts` が持つ非機能的な挙動を落とさない。

## 7. Acceptance Criteria

- [ ] AC-1: `worker-entry.ts` 経由でのリクエストが `app.ts` の Hono route を通る。
- [ ] AC-2: `routes/fetch.ts`、`routes/match-playback-route.ts` とその test file が repo に存在しない。
- [ ] AC-3: `npm run dev` で dummy backend 経由の一覧・詳細・audio 取得が引き続き動作する。
- [ ] AC-4: `DESIGN.md` の該当記述が新しい入口構成と一致する。

## 8. Verification

```bash
cd apps/playback && npm run typecheck
cd apps/playback && npm run lint:layers
cd apps/playback && npm run test:unit
cd apps/playback && npm run dev
```

`npm run dev` 起動後、`curl localhost:3000/episodes` 等で手動確認する。

## 9. Dependencies

- blocked by: `playback-worker-hono-route-definition.md`

## 10. Risks

1. `fetch.ts` 削除時に、`mapRuntimeConfigErrorToExternal` のような error mapping ロジックを route 定義側へ移し忘れると、Configuration Error のハンドリングが欠落する risk がある。削除前に該当ロジックの移植先を確認する。

## 11. Notes

本 Issue 完了時点で worker 側の Hono 移行が完了する。web 側の React 化（`playback-web-*` 系 Issue）とは独立に進行できる。
