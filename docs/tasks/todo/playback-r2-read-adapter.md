# feature(playback): R2 binding で episode を読む Adapter を本実装する

## 1. Summary

このIssueでは playback Worker の R2 binding `EpisodeRepository` を本実装し、結線可能な状態にする。**本番正本の切替と OAuth 削除は列 6・7**。

## 2. Context

1. 事実: A が `infrastructure/r2` stub・`episode-layout.md` を固定済み。
2. 事実: 順番は Decision `2026-09-14T12-49-26`（本 Issue は列 5）。実施の形 `11-04-30` / error `11-19-21`。
3. 事実: R2 登録は完了済み。登録手順は書かない。
4. 事実: 現行本番は当面 Drive。本 PR 単独で本番正本を切替えない。
5. 運用: 1 Issue = 1 PR。

## 3. Canonical Sources

1. 順番: `docs/decisions/2026-09-14T12-49-26-docs-generator-r2-write-details.md`
2. 契約: `contracts/episode-layout.md` / `episode-repository.ts` / `infrastructure/r2`
3. 実施の形 / error: `2026-09-14T11-04-30` / `11-19-21`
4. 方針: `2026-09-13T14-22-55`
5. HTTP / Access: `apps/playback/contracts/` / `DEPLOY.md` / `2026-08-25T17-10-00`
6. test 方針: testing-strategy（再掲しない）

## 4. Scope

### In Scope

1. R2 `EpisodeRepository` 本実装（A 足場 test を behavior test へ置換）
2. `wrangler.jsonc` の R2 binding と `PlaybackEnv` 形（本番 OAuth 削除は列 7）
3. Composition の R2 結線口（本番正本切替は列 6）
4. HTTP error 写像維持。Drive 固有 Error 名の掃除準備

### Out of Scope

1. generator 書込（列 4）
2. 疎通 dispatch・人手移行・本番同着切替（列 6）
3. System/E2E 緑化・OAuth 削除（列 7）
4. 登録手順の記述
5. cache

## 5. Contract

1. `listManuscripts` / `getAudio` は Port 契約を満たす。音声欠落は `undefined`。
2. storage I/O 失敗は Infrastructure Error（`11-19-21`）。
3. 成功音声の `Content-Type` は `audio/mpeg`（mp3 前提）。
4. HTTP 写像は既存維持。R2 専用外部 code は増やさない。
5. Error message / log に bucket・key・secret 実値を載せない。

## 6. Constraints

1. 公開直 URL / `r2.dev` で音声を返さない。
2. Application は R2 vendor 型を知らない。
3. 本 PR だけで本番正本を R2 にしない。OAuth を本 Issue で消さない。

## 7. Acceptance Criteria

1. [ ] A 足場 test が behavior test に置換されている
2. [ ] SU / Narrow / Broad が R2（または double）経路で緑
3. [ ] 音声欠落は throw せず、I/O 失敗だけが Infra Error になる
4. [ ] Error message に bucket / key / secret 実値が含まれない
5. [ ] unit coverage gate が緑
6. [ ] 本番正本切替・OAuth 削除が列 6・7 である旨が Notes に明示されている

## 8. Verification

1. `scripts/playback` の unit / 静的 check
2. Narrow / Broad

## 9. Dependencies

1. `audio-mp3-cutover-migrate`（列 3）完了後に本番向け結線を進める
2. 対: `generator-r2-write-adapter`（列 4）
3. 次: `r2-smoke-migrate-cutover`（列 6）→ `r2-post-cutover-verify-oauth`（列 7）

## 10. Notes

本番同着切替は列 6。OAuth / `DRIVE_*` 削除と正式 System/E2E は列 7。
