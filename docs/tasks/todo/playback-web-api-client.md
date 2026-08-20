## Playback web: API Client の応答処理

Issue draft。`create-issue` で正式化する前の一時置き場。

type: `feat` / scope: `playback-web`

## 1. Summary

このIssueでは、`apps/playback/web/src/api/` の API Client が worker の HTTP 応答を解釈し、成功と失敗を `ApiResult` として返せるようにする。型・写像・factory の骨格は実装済みで、現在は fetch 後に必ず `invalid_response` を返す状態にある。完了後は、成功応答が data を持つ Result になり、status 種別・network 障害・schema 不適合がそれぞれ対応する error code へ落ちる。

## 2. Context

1. `apps/playback/web/src/api/` に型・写像表・factory と `*.sociable_unit.test.ts` が既にある。3 method は URL を組んで `deps.fetch` を呼ぶところまで動く
2. 3 method の応答処理部分に `todo:` tag 付き comment が置いてある。この Issue はその箇所を実装する
3. worker 側（`apps/playback/worker/`）は実装済み。status と `code` を 1 対 1 で返す。runtime config 不備は `500 / configuration_error`、外部service一時不能は `503 / unavailable`
4. `toApiErrorCode` は export 済みだが、まだどの method からも呼ばれていない
5. 失敗 body の JSON を読まない方針は決定済み（`docs/decisions/2026-08-20T13-44-08-playback-web-api-client.md`）

## 3. Canonical Sources

1. `apps/playback/web/src/api/playback-api-error.ts` — API error code 集合と `toApiErrorCode` 写像
2. `apps/playback/web/src/api/api-result.ts` — `ApiResult<T>`
3. `apps/playback/web/src/api/playback-api-client.ts` — client interface、`buildRequestUrl`、factory
4. `apps/playback/contracts/` — path 定数、Response schema、`classifyHttpStatus`
5. `docs/decisions/2026-08-20T13-44-08-playback-web-api-client.md` — status 単一情報源、1 対 1 変換、`Blob` 採用の根拠
6. `DESIGN.md` §2 — web は worker の HTTP のみ知る。`playback/contracts` は API Client から import 可
7. `architecture/frontend/api-client` skill — 3 段責務、Result 型、throw 禁止
8. test 方針は `testing-strategy` skill を参照する

## 4. Scope

### In Scope

1. `listEpisodes` / `getEpisode` の応答処理（status 分類 → JSON parse → schema 検証 → Result）
2. `fetchAudio` の応答処理（status 分類 → `Blob` 取得 → Result）
3. `deps.fetch` の reject と body 読み取り失敗の吸収
4. `toApiErrorCode` を実経路へ配線する
5. 上記に対応する Unit test の追加

### Out of Scope

1. 型・写像表・factory・`buildRequestUrl` の変更（実装済み）
2. ViewModel / page / components
3. `baseUrl` の注入元（`playback-runtime-config-boundary.md`）
4. UI 向け表示文への変換
5. Access / 認証 header
6. Vite 設定・toolchain
7. 契約 schema の変更

## 5. Contract

公開 interface は `playback-api-client.ts` にある `PlaybackApiClient` を変更しない。振る舞いだけを与える。

| method | 成功時の `data` | 使う schema |
|---|---|---|
| `listEpisodes` | `ListEpisodesResponse` | `ListEpisodesResponseSchema` |
| `getEpisode` | `GetEpisodeResponse` | `GetEpisodeResponseSchema` |
| `fetchAudio` | `Blob` | schema 検証なし |

失敗時の `error` は `PlaybackApiErrorCode`。決め方は次の順で 1 つに定まる。

1. `deps.fetch` が reject → `network_error`
2. `classifyHttpStatus(status)` が `success` 以外 → `toApiErrorCode` の戻り値
3. body 読み取りまたは schema 検証が失敗 → `invalid_response`

失敗応答の body は読まない。

## 6. Constraints

1. 公開 method は throw しない。全経路で `ApiResult` を返す。`toApiErrorCode` は契約に無い kind で throw するため、呼ぶ側が catch して `invalid_response` へ落とす
2. status 分類を Client 側へ再実装しない。`classifyHttpStatus` へ委譲する
3. 契約 error code を web 語彙へ写す時は `toApiErrorCode` を通す。契約 code を直接 `error` へ代入しない
4. `audioRef` を分解して path を組み直さない
5. `deps.fetch` に既定値を与えない
6. 音声形式の literal（MIME 等）を web 側に書かない

## 7. Acceptance Criteria

1. [ ] 200 と合法な一覧 JSON の応答で、`listEpisodes` が `ok: true` と `ListEpisodesResponse` を返す
2. [ ] 200 と合法な 1 件 JSON の応答で、`getEpisode` が `ok: true` と `GetEpisodeResponse` を返す
3. [ ] 200 の音声応答で、`fetchAudio` が `ok: true` と `Blob` を返す
4. [ ] 404 の応答で `error` が `episode_not_found` になる
5. [ ] 400 の応答で `error` が `validation_error` になる
6. [ ] 503 の応答で `error` が `unavailable` になる
7. [ ] 500 の応答で `error` が `configuration_error` になる
8. [ ] `classifyHttpStatus` が `client_error` と分類する status で `error` が `client_error` になる
9. [ ] `deps.fetch` が reject する時、`error` が `network_error` になる
10. [ ] 200 だが schema に合わない JSON で `ok: true` にならず `error` が `invalid_response` になる
11. [ ] 失敗応答の body を読まずに `error` が決まる
12. [ ] 公開 method がどの経路でも throw しない

## 8. Verification

```bash
cd apps/playback
npm run test:unit
npm run typecheck
npm run lint
npm run format:check
```

4 つすべて pass する。worker の実機は不要。

## 9. Dependencies

先行: `apps/playback/contracts/` と `apps/playback/web/src/api/` の型・写像・factory（いずれも実装済み）

後続: UI task（本 Client を ViewModel から呼ぶ。未切り出し）

## 10. Risks

1. 失敗 body を読む実装を足すと status と二重の情報源になる risk。AC-10 で検出する
2. 契約 code を `toApiErrorCode` を通さず `error` へ直接代入すると層の境界が消える risk。Constraints 3 と review で防ぐ

## 11. Notes

1. `network_error` と `invalid_response` は web Client 固有で、worker の `ErrorResponse` enum には含まれない
2. status 分類そのものの正しさは `contracts/http-error.sociable_unit.test.ts` の責務。この Issue の test で具体的な status 番号と分類結果の対応を再検証しない
3. 次アクション: `create-issue` で Issue 化する
