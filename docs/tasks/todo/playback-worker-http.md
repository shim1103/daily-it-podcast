## Playback worker: HTTP 境界（Route / Controller / Error 写像）

Issue draft。`create-issue` で正式化する前の一時置き場。

## 1. Summary

worker の Route / Controller / Composition で、契約どおりの 3 HTTP 操作（List / Get JSON / Get 音声）を公開し、unknown 入力を schema 検証、UseCase 結果を Response schema、内側 Error を External status + `ErrorResponse` に写す。

## 2. Context

- path・Response・`code` enum は `apps/playback/contracts/` が SSOT
- Drive 読取 Port / UseCase / Fake は済。Google Drive API 本番は `playback-worker-drive-adapter.md`。本 task は HTTP 層のみ
- Access 検証・wrangler 本番設定は未確定のため Out of Scope

## 3. Canonical Sources

- `apps/playback/contracts/http.ts` — path（`listEpisodesPath` / `episodePath` / `episodeAudioPath`）、`episodeAudioContentType`、Request/Response schema
- `apps/playback/contracts/http-error.ts` — `classifyHttpStatus`（web 側用。worker Route は既知 status → `code` の写像を Route/Controller が持つ）
- `docs/decisions/2026-08-17T17-40-00-feature-playback-web.md` — status 400/404/503、`episode_not_found` 等
- `DESIGN.md` §2 — Route / Controller / Composition。`playback/contracts` は Route / Controller のみ import 可
- `http-boundary` skill — Driving Adapter / Controller 3 段分離
- test 方針 — `testing-strategy` skill

## 4. Scope

### In Scope

- `GET listEpisodesPath` → 200 + `ListEpisodesResponseSchema`
- `GET episodePath({episodeId})` → 200 + `GetEpisodeResponseSchema`（`audioRef` 必須）
- `GET episodeAudioPath({episodeId})` → 200 + `episodeAudioContentType`（`audio/wav`）body（JSON なし）
- path param / body を `unknown` で受け、Controller が schema parse
- Domain 不在 → 404 + `{ code: "episode_not_found" }`
- 入力 validation 失敗 → 400 + `{ code: "validation_error" }`
- Drive Infrastructure 失敗 → 503 + `{ code: "unavailable" }`
- Composition Root が UseCase + Controller を結線
- Controller / Route の Unit test（UseCase は Fake）

### Out of Scope

- Drive API 呼び出し本体（→ `playback-worker-drive-adapter.md`）
- `apps/playback/web/`
- UI / ViewModel
- Cloudflare Access JWT 検証（将来 task）
- wrangler 本番 deploy・Access ポリシー
- CORS 詳細（必要最小の Response header のみ。過剰設計しない）
- 契約 schema 変更

## 5. Contract

| 操作 | Method + path | 成功 | 失敗 body |
|---|---|---|---|
| List | `GET` + `listEpisodesPath` | 200 JSON `ListEpisodesResponse` | 503 `unavailable` |
| Get JSON | `GET` + `episodePath(id)` | 200 JSON `GetEpisodeResponse` | 400 `validation_error` / 404 `episode_not_found` / 503 `unavailable` |
| Get 音声 | `GET` + `episodeAudioPath(id)` | 200 `audio/wav` | 404 / 503 同上 |

- 失敗 JSON は `ErrorResponseSchema`（`code` のみ。`message` 無し）
- 不完全ペア専用 `code` は返さない（404 に畳む）

## 6. Constraints

- Route は schema parse しない。Controller が parse（http-boundary）
- Infrastructure を Route から直接呼ばない
- External Error 写像は 1 箇所（宣言的 data 推奨）。`classifyHttpStatus` は web 向け。worker は 400/404/503 を直接返す

## 7. Acceptance Criteria

- [ ] AC-1: Fake UseCase 成功時、List/Get JSON Response が契約 schema を Pass する
- [ ] AC-2: 空 `episodeId` が 400 + `validation_error` になる
- [ ] AC-3: UseCase が Domain 不在を返すと 404 + `episode_not_found` になる
- [ ] AC-4: UseCase が Infrastructure 失敗を返すと 503 + `unavailable` になる
- [ ] AC-5: Get 音声 Route が `Content-Type: audio/wav` で byte を返す（Fake UseCase）
- [ ] AC-6: 契約外 `code` を Response に出さない

## 8. Verification

- `cd apps/playback && npm run test:unit` — worker Route/Controller 隣 `*.test.ts` Pass
- wrangler dev による manual 確認は任意（本 Issue 必須にしない）

## 9. Dependencies

- 先行: `EpisodeRepository` + Fake（済）
- 並行可: `playback-worker-drive-adapter.md`
- blocks: なし（web API Client は Fake HTTP で並行可）

## 10. Risks

- Route にビジネス判断が漏れる risk → UseCase 結果のみを Controller が写像

## 11. Notes

- local 実行用 wrangler 最小 config は In Scope 边缘。Access 未確定なら dev 用 bypass は document せず、test は Fake で完結させる
- 次アクション: `create-issue` で Issue 化
