## Playback web: API Client（worker HTTP → Result）

Issue draft。`create-issue` で正式化する前の一時置き場。

## 1. Summary

`apps/playback/web/src/api` に worker への HTTP Client を実装する。契約 path で fetch し、Response schema を parse、失敗を Result 型で返す。UI 状態・DOM は持たない。

## 2. Context

- web は worker HTTP のみ知る（`DESIGN.md` §1）。Drive / generator に触れない
- 契約 SSOT は `apps/playback/contracts/`。field 写し禁止
- worker 実装を待たず、Fake HTTP / Stub で Unit test 可能にする
- page / Feature / ViewModel は別 task（本 Issue Out of Scope）

## 3. Canonical Sources

- `apps/playback/contracts/` — path 定数、Request/Response schema、`classifyHttpStatus`、`playbackHttpErrorCodes`
- `docs/decisions/2026-08-17T17-40-00-feature-playback-web.md` — status 級・3 `code`
- `docs/decisions/2026-08-18T11-12-00-feature-playback-web.md` — web は TS + Pico.css（React なし）
- `DESIGN.md` §2 — API Client 層。`playback/contracts` import 可
- `architecture/frontend/api-client` skill — 3 段責務・Result 型・throw 禁止
- `http-boundary` skill — Outbound fetch 責務
- test 方針 — `testing-strategy` skill

## 4. Scope

### In Scope

- `listEpisodes(baseUrl)` → `ListEpisodesResult`
- `getEpisode(baseUrl, episodeId)` → `GetEpisodeResult`
- `fetchAudio(baseUrl, audioRef)` → 音声 byte Result（`audioRef` は opaque。分解しない）
- request 構築: `listEpisodesPath` / `episodePath` / `audioRef` 文字列をそのまま URL に結合（baseUrl は引数。契約に origin を書かない）
- response: status → 既知は契約どおり。未知は `classifyHttpStatus`。接続失敗は `network_error`（契約 enum 外・Client 固有）
- schema parse 失敗も失敗 Result（throw しない）
- API Client 隣の Unit test（network-level Stub）

### Out of Scope

- ViewModel（DOM）/ page / components
- worker 側実装
- Access / 認証 header（未確定。必要なら baseUrl または fetch wrapper 引数で後続）
- UI エラー表示の住み分け
- Vite 設定・toolchain 全体
- 契約 schema 変更

## 5. Contract

**公開 Client（概念）**

| 関数 | 成功 | 失敗 `error` |
|---|---|---|
| listEpisodes | `{ ok: true, data: ListEpisodesResponse }` | `episode_not_found` / `validation_error` / `unavailable` / `network_error` / 未知 4xx 級 → `client_error` 相当 |
| getEpisode | `{ ok: true, data: GetEpisodeResponse }` | 同上 |
| fetchAudio | `{ ok: true, data: ArrayBuffer \| Uint8Array }` | 同上 |

- worker が返す `code` は `ErrorResponseSchema` の enum のみパース。enum 外は失敗 Result
- `classifyHttpStatus`: 既知 404 は `episode_not_found` のまま。未知 status は `/100` 級（契約実装済み）

## 6. Constraints

- throw 禁止。全経路 Result
- Feature / ViewModel から import されない（逆: Client は UI を知らない）
- `audioRef` を parse して path 再構成しない

## 7. Acceptance Criteria

- [ ] AC-1: Stub が 200 + 合法 List JSON のとき success Result になる
- [ ] AC-2: Stub が 404 + `{ code: "episode_not_found" }` のとき同 `error` になる
- [ ] AC-3: fetch 例外が `network_error` になる
- [ ] AC-4: 未知 502 が `unavailable` 級に落ちる（`classifyHttpStatus` 経由）
- [ ] AC-5: 不正 JSON shape が success にならない

## 8. Verification

- `cd apps/playback && npm run test:unit` — `web/src/api/**/*.test.ts` Pass
- worker 実機不要

## 9. Dependencies

- 先行: `apps/playback/contracts/`（済）
- 並行可: `playback-worker-http.md`（Stub で完結）
- 後続: UI task（本 Client を ViewModel から呼ぶ）

## 10. Risks

- baseUrl 二重スラッシュ risk → Client 内で 1 箇所に normalize 規則を置く（契約には書かない）

## 11. Notes

- `network_error` は web Client 固有。worker `ErrorResponse` enum には含めない
- 次アクション: `create-issue` で Issue 化
