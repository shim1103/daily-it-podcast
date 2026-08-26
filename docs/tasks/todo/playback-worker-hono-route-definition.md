## 1. Summary

この Issue では、`apps/playback/worker/src/routes/app.ts` の Hono instance へ route 定義（Controller 呼び出しの wiring）を追加し、`match-playback-route.ts` が担っている method / path 分岐を Hono の route 構文へ移植する。

## 2. Context

1. `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md` で Hono 導入が決定済み。
2. A 区分（契約固定）で `routes/app.ts` は最小限の `Hono` instance と `AppType` export のみ既に存在する（route handler 本体は未実装）。
3. 現行の `routes/match-playback-route.ts` は method 判定・path segment 分解・exhaustive check を自作しており、`routes/fetch.ts` がこれを呼び出して Controller へ振り分けている。
4. 仮定: この Issue では `match-playback-route.ts` と `fetch.ts` を削除しない。新しい Hono route 定義を `app.ts` に追加するだけで、既存の入口切り替えは別 Issue（`playback-worker-hono-entry-cutover`）が担う。

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md` — React/Hono/Hono RPC 採用の決定。
2. `apps/playback/worker/src/routes/app.ts` — Hono instance の正本候補。
3. `apps/playback/worker/src/routes/match-playback-route.ts` — 移植元の method/path 分岐ロジック。
4. `apps/playback/worker/src/composition/root.ts` — Controller の組み立て。
5. `apps/playback/contracts/index.ts` — HTTP 契約（path 定数、status 級）の正本。
6. architecture — `backend/route-handler.md`（薄い入口の責務）。

## 4. Scope

### In Scope

1. `app.ts` へ一覧取得・単一 episode 取得・audio 取得の3 route を Hono 構文で定義する。
2. 各 route handler から既存 Controller（`listEpisodesController` 等）を呼び出す wiring。
3. 未一致 path / method の扱いを、現行 `match-playback-route.ts` の `unmatched` 相当の挙動に合わせる。

### Out of Scope

1. `worker-entry.ts` の入口切り替え（別 Issue）。
2. `fetch.ts` / `match-playback-route.ts` の削除（別 Issue）。
3. Controller 以下（UseCase、Infrastructure）の変更。

## 5. Contract

1. `app.ts` は `AppType` を継続して export する。route 追加によって型が変わる場合、`api/playback-rpc-client.ts` の `hc<AppType>()` が参照する型と矛盾しないこと。
2. 一覧取得は `GET {listEpisodesPath}`、単一取得は `GET {listEpisodesPath}/:episodeId`、audio 取得は `GET {listEpisodesPath}/:episodeId/audio` の契約を維持する（`apps/playback/contracts/index.ts` の path 定数を正とする）。
3. 成功時のレスポンス形式（status、body）は既存 `fetch.ts` の挙動と同一にする。

## 6. Constraints

1. Controller・UseCase・Infrastructure 層への変更は行わない（Ring 対応を変えない）。
2. `match-playback-route.ts` の exhaustive check（`never` による網羅性担保）と同等の安全性を、Hono route 定義でも損なわない。

## 7. Acceptance Criteria

- [ ] AC-1: `app.ts` の Hono instance へ一覧取得 route を呼ぶと、既存 `fetch.ts` 経由と同じ成功レスポンスが得られる。
- [ ] AC-2: 単一 episode 取得 route が episodeId を正しく Controller へ渡す。
- [ ] AC-3: audio 取得 route が既存 `createAudioResponse` 相当のレスポンス形式を返す。
- [ ] AC-4: 未定義 path / method へのリクエストが、契約上の `validation_error` 相当を返す。

## 8. Verification

```bash
cd apps/playback && npm run test:unit
cd apps/playback && npm run typecheck
cd apps/playback && npm run lint:layers
```

`app.ts` に対応する `app.sociable_unit.test.ts` を拡張し、route ごとの成功系・異常系を検証する。

## 9. Dependencies

- blocks: `playback-worker-hono-entry-cutover`（この Issue の route 定義が完了しないと入口切り替えができない）

## 10. Risks

1. Hono の path param 構文（`:episodeId`）と既存 `decodeURIComponent` 相当のデコード処理が食い違うと、日本語 episodeId 等で挙動が変わる risk がある。既存 `match-playback-route.ts` の `decodePathSegment` と同等のデコードを route handler 側で担保することで防ぐ。

## 11. Notes

`match-playback-route.ts` 自体の削除は本 Issue の scope 外。並行して古い router が残る期間が生じるが、`fetch.ts` の入口は変えないため本番挙動への影響はない。
