## 1. Summary

この Issue では、`apps/playback/web/src/view-models/episode-list-view-model.ts` を React hooks 実装へ書き換え、`apps/playback/web/src/api/playback-api-client.ts` の内部実装を Hono RPC client（`playback-rpc-client.ts`）へ差し替える。

## 2. Context

1. `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md` で React・Hono RPC 導入が決定済み。
2. A 区分で `apps/playback/web/src/api/playback-rpc-client.ts` に `hc<AppType>()` の型配線のみ存在する（実際の呼び出し実装は未着手）。
3. `frontend/view-model.md`（architecture skill）は ViewModel 層の platform 実装を `React hooks` と定義しており、現行の vanilla 実装（閉じた module + 購読関数）はこの定義の読み替えだった。React 導入によりこの skill の想定通りの実装へ戻す。
4. 仮定: `PlaybackApiClient` interface（`listEpisodes()`, `getEpisode(episodeId)` が `ApiResult` を返す契約）は変更しない。内部実装のみ Hono RPC 呼び出しへ差し替える。

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md` — React/Hono RPC 採用の決定。
2. architecture — `frontend/view-model.md`（Ring 対応、throw/logging、import ルール、unit test 観点）。
3. `apps/playback/web/src/api/playback-api-client.ts` — 差し替え対象の interface 正本。
4. `apps/playback/web/src/api/playback-rpc-client.ts` — 差し替え先の RPC client 型配線。
5. `apps/playback/web/src/view-models/episode-list-view-model.ts` — 書き換え対象。
6. `apps/playback/web/src/api/api-result.ts` — throw しない Result 型契約。

## 4. Scope

### In Scope

1. `episode-list-view-model.ts` を `useState` / `useEffect` ベースの custom hook へ書き換える。
2. `playback-api-client.ts` の `listEpisodes` / `getEpisode` 内部実装を `playback-rpc-client.ts` の `hc<AppType>()` 呼び出しへ差し替える。
3. 既存の event / async error 処理方針（`frontend/view-model.md` §5：throw を UI へ伝播させず state で表現する）を維持する。

### Out of Scope

1. Component（Feature/Primitive）の JSX 化（別 Issue）。
2. worker 側の Hono route 定義（別 Issue）。

## 5. Contract

1. `PlaybackApiClient` の public interface（method 名、引数、`ApiResult<T>` の戻り値型）は変更しない。
2. hook の戻り値（state 遷移の形：loading / success / error 等）は既存 vanilla 実装の状態遷移と同一の意味を保つ。

## 6. Constraints

1. `frontend/view-model.md` §4 の throw 禁止・logging 条件（外部ライブラリ直接呼び出し時のみ許可）を守る。
2. `frontend/view-model.md` §6 の import ルール（Feature/Primitive Component・境界共有型を直接 import しない）を守る。

## 7. Acceptance Criteria

- [ ] AC-1: hook 化後も一覧取得の成功系が既存 test と同じ state 遷移を返す。
- [ ] AC-2: API 失敗時に throw ではなく state で失敗が表現される。
- [ ] AC-3: `playback-api-client.ts` の呼び出し先が Hono RPC client に置き換わっている。
- [ ] AC-4: `PlaybackApiClient` の呼び出し元（既存 component）に変更が不要である。

## 8. Verification

```bash
cd apps/playback && npm run typecheck
cd apps/playback && npm run test:unit
cd apps/playback && npm run lint:layers
```

test 方針は `testing-strategy` を参照する。hook の検証観点は `frontend/view-model.md` §8（state 遷移、API Client 結果の処理、cleanup）に従う。

## 9. Dependencies

- blocks: `playback-web-feature-component-jsx`（Feature Component が hook の型に依存する）

## 10. Risks

1. Hono RPC client のレスポンス型と既存 `ListEpisodesResponse` / `GetEpisodeResponse` 型が完全一致しない場合、型変換が必要になる risk がある。`apps/playback/contracts/` の型を正として一致を確認する。

## 11. Notes

`components/primitive/labeled-text.tsx` の JSX 化（`playback-web-primitive-component-jsx`）はこの Issue と依存関係が無く、完了済み。
