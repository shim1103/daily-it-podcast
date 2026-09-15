## Playback 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md  
HTTP 契約の正: `apps/playback/contracts/`  
配置契約の正: `contracts/episode-layout.md`（音声は mp3）  
deploy・Access・GHA 運用の正: `DEPLOY.md`  
層規則・test 配置の正: `DESIGN.md`  
再発する判断の正: `docs/decisions/`

未完了の達成契約は `docs/tasks/todo/` の Issue file が正。本 lane は進捗 index のみ。GitHub Issue 化しない運用。storage 実施順の正は Decision `2026-09-14T12-49-26`（release 専用 Issue は作らない）。

### 済み（要約）

1. web↔worker HTTP / Drive adapter / UI / 層検知 / React+Hono
2. deploy 前準備・初回 `wrangler deploy`・Access OTP（許可/拒否）・同一 origin の一覧・原稿・再生
3. 認証済み browser E2E（週次 `playback-e2e` + `workflow_dispatch`。安定 fixture は本番 `DRIVE_FOLDER_ID`）
4. Narrow / Broad Integration（gate 内）と `test/integration` + `test/e2e` 配置
5. playback web 直交 Decision（B）と A 契約（型・stub・SU test）
6. playback web UI rewrite / ViewModel stack 差し替え（達成契約 file 削除済み）
7. 運用後続 docs（rollback / observability / 完了境界）
8. 原稿 body の opening / ending を `{ text, startSec }` object へ（3 bookend を「本文 + startSec」へ揃える。text は定型込み朗読全文）。判断は `docs/decisions/2026-09-04T19-30-00-feature-playback-e2e-redeploy-master.md`（先行 `16-44-46` / `16-00-00` を reconcile）
9. 音源 load 直後の seek が 0:00 に落ちる bug を修正（`readyState` 未達なら `loadedmetadata` を待って `currentTime` 代入）。e2e に seek 回帰 test 追加。安定 fixture を新契約 episode へ差し替え
10. 週次 `playback-e2e` の全滅を Google OAuth refresh token 失効と特定。consent screen を Production 固定して 7 日失効を運用から外す（`docs/decisions/2026-09-07T22-25-00-fix-playback-e2e-test.md`）。`DriveError` の非 2xx message を呼び出し種別付きにして切り分け可能化
11. 音声保存・配信の契約を mp3 へ（A）と Decision（`2026-09-13T13-40-29` / `13-41-00` / encode `16-32-57` / runtime 工場 `17-38-37`）。generator 側は encode Port + ffmpeg Adapter 本実装・`ProduceEpisode` 結線済み
12. playback 読取・HTTP・dev fake を mp3 契約へ揃え（`playback-audio-mp3-read`。達成契約 file 削除済み）
13. mp3 同着切替 + wav 一括 + fixture（`audio-mp3-cutover-migrate`。達成契約 file 削除済み）。安定 fixture README は `.mp3`。`playback-e2e` PASS

### 未完了（storage 順・Decision `2026-09-14T12-49-26`）

1. `playback-r2-read-adapter.md` — 列 5。R2 binding 読取 **Adapter 振る舞い**本実装（local binding infra は C1 済み）
2. `r2-smoke-migrate-cutover.md` — 列 6（共有）。疎通・人手移行・R2 同着切替
3. `r2-post-cutover-verify-oauth.md` — 列 7（共有）。System/E2E・OAuth 削除

### 実施順 index（Issue file 単位）

| 列 | Issue file | 備考 |
|---|---|---|
| 1 | `generator-audio-mp3-encode-write` | **済み**（達成契約 file 削除済み） |
| 2 | `playback-audio-mp3-read` | **済み**（達成契約 file 削除済み） |
| 3 | `audio-mp3-cutover-migrate` | **済み**（達成契約 file 削除済み） |
| 4 | `generator-r2-write-adapter` | generator lane。**Writer 済み**（達成契約 file 削除済み） |
| 4b | `generator-r2-completed-episode-lookup` | generator lane。Lookup 振る舞い。Adapter NI は httptest |
| 4c | `generator-r2-test-peer-scope` | C1=infra。**済み**（達成契約 file 削除済み）。local S3 peer 到達 + `getPlatformProxy`（`2026-09-16T00-20-08`） |
| 5 | `playback-r2-read-adapter` | 本 lane。**behavior**（infra は 4c） |
| 6 | `r2-smoke-migrate-cutover` | 旧 `R-r2-cutover` の切替達成を吸収。登録手順は書かない |
| 7 | `r2-post-cutover-verify-oauth` | System/E2E 後に OAuth 削除 |

R2 登録は完了済み（列に含めない）。R2 Adapter NI の正 peer は Decision `2026-09-16T00-20-08`（本番口は `11-04-30`）。

### 未決 index（D）

| topic | 概要 |
|---|---|
| 薄い cache の細部 | 方針は `2026-09-13T14-23-30`。未決: header 具体値・edge 設定・Access 下 browser cache 実測 |
| Access 下 audio の browser HTTP cache | 未実測。DevTools で確認が次 |

### 方針 index

各判断の Reason / Rejected は `docs/decisions/`。閾値・入口の正は `DESIGN.md` / `DEPLOY.md`。

1. 音声の保存・配信形式は **mp3**（`contracts/episode-layout.md` / `2026-09-13T13-40-29`）
2. 着手順: mp3 → R2 → 薄い cache（`2026-09-13T13-41-00`）。Issue 分割は `2026-09-14T12-49-26`
3. 現行 storage runtime は Drive。R2 方針 `14-22-55`・実施の形 `11-04-30`・error `11-19-21`。cache は R2 後（`14-23-30`）
4. generator 書込とは runtime 共有しない（読取専用）
