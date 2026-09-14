# feature(generator): R2 へ episode を書く Adapter を本実装する

## 1. Summary

このIssueでは generator の R2（S3 互換）`EpisodeWriter` を本実装し、Composition が R2 へ結線できる状態にする。完了後、test / 結線可能な入口から `{episodeId}.json` / `{episodeId}.mp3` を公開順で put できる。**本番正本の切替は列 6**。

## 2. Context

1. 事実: A が `infrastructure/r2` stub・`episode-layout.md`・R2 env 名定数を固定済み。
2. 事実: 順番は Decision `2026-09-14T12-49-26`（本 Issue は列 4）。実施の形 `11-04-30` / error `11-19-21`。
3. 事実: R2 登録は完了済み。登録手順は書かない。
4. 事実: 現行本番 runtime は当面 Drive。本 PR 単独で本番正本を切替えない。
5. 運用: 1 Issue = 1 PR。

## 3. Canonical Sources

1. 順番: `docs/decisions/2026-09-14T12-49-26-docs-generator-r2-write-details.md`
2. 契約: `contracts/episode-layout.md` / `port/episode_writer.go` / `infrastructure/r2`
3. 実施の形 / error: `2026-09-14T11-04-30` / `11-19-21`
4. 方針: `2026-09-13T14-22-55`
5. 公開順・upsert: `2026-08-30T23-32-00` / `23-31-00`
6. test 方針: testing-strategy（再掲しない）

## 4. Scope

### In Scope

1. R2 `EpisodeWriter` 本実装（A 足場 test を behavior test へ置換）
2. config Load・Composition 結線口（本番正本切替は列 6）
3. Narrow: S3 互換 double で put 名・MIME・公開順・upsert・retry を観測

### Out of Scope

1. playback 読取（列 5）
2. 疎通 dispatch・人手移行・本番同着切替（列 6）
3. System/E2E 緑化・OAuth 削除（列 7）
4. 登録手順の記述
5. cache

## 5. Contract

1. `EpisodeWriter.Write` 成功時、対象 bucket に非空 `{episodeId}.json` と `{episodeId}.mp3` がある。
2. put 順は json → mp3。同 key は upsert。
3. network / 5xx / 429 は有限 retry。その他 4xx は fail-fast（`11-19-21`）。
4. Application の error 写像は変えない。
5. Error message / log に bucket・key・Account ID・Access Key・secret 実値を載せない。
6. production `EpisodeWriter` に delete を公開しない。

## 6. Constraints

1. Application は R2 / S3 SDK を import しない。
2. Port を `PutObject` 単位に下げない。
3. 本 PR だけで本番正本を R2 にしない。

## 7. Acceptance Criteria

1. [ ] A 足場 test が behavior test に置換されている
2. [ ] Narrow で json→mp3 順・MIME・upsert が観測できる
3. [ ] Narrow で 5xx/network の有限 retry と 4xx fail-fast が観測できる
4. [ ] Error message に bucket / key / secret 実値が含まれない
5. [ ] generator unit / static gate が緑
6. [ ] 本番正本がまだ Drive（または未切替）のままである、または列 6 と同着である旨が Notes に明示されている

## 8. Verification

1. `scripts/generator/check-static.sh`
2. `scripts/generator/test-unit.sh`
3. Narrow

## 9. Dependencies

1. `audio-mp3-cutover-migrate`（列 3）完了後に本番向け結線を進める
2. 対: `playback-r2-read-adapter`（列 5）
3. 次: `r2-smoke-migrate-cutover`（列 6）

## 10. Notes

本番同着切替・人手移行・疎通は列 6。OAuth 削除は列 7。
