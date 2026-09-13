## Playback 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md  
HTTP 契約の正: `apps/playback/contracts/`  
配置契約の正: `contracts/drive-layout.md`（音声は mp3）  
deploy・Access・GHA 運用の正: `DEPLOY.md`  
層規則・test 配置の正: `DESIGN.md`  
再発する判断の正: `docs/decisions/`

未完了の達成契約は `docs/tasks/todo/playback-*.md` および共有の `audio-mp3-batch-migration.md` が正。本 lane は進捗 index と release 単位のみ。GitHub Issue 化しない運用。1 Issue file = 1 PR。

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
11. 音声保存・配信の契約を mp3 へ（A）と Decision（`2026-09-13T13-40-29` / `13-41-00`）。`EncodeWAVToMP3` stub 未結線

### 未完了

1. `playback-audio-mp3-read.md` — 読取・HTTP を mp3 契約で完了（1 PR）
2. `audio-mp3-batch-migration.md` — 既存 wav 一括 mp3・安定 fixture（1 PR。generator lane からも参照）

### release 単位（1 Issue ≠ 1 deploy）

| release | 同着 PR（Issue） | 備考 |
|---|---|---|
| `R-mp3-cutover` | `generator-audio-mp3-encode-write` + `playback-audio-mp3-read` | 書込/読取を分けて本番 deploy しない |
| `R-mp3-migrate` | `audio-mp3-batch-migration` | cutover 後 |

### 未決 index（D）

方針が Decision 済みで、実施契約に落ちない残りだけ置く。R2 / cache は **C Issue にしない**（Decision のみ）。

| topic | 概要 |
|---|---|
| R2 移行の細部 | 方針は `2026-09-13T14-22-55`（完全移行・mp3 後）。未決: binding / bucket・Worker proxy vs 直 URL・credential 注入・cutover 手順。Issue 化は細部が埋まるまでしない |
| 薄い cache の細部 | 方針は `2026-09-13T14-23-30`（R2 後・Cache-Control + CF edge・厚い Worker cache しない）。未決: header 具体値・edge 設定・Access 下 browser cache 実測 |
| Access 下 audio の browser HTTP cache | 未実測。DevTools で確認が次 |

### 方針 index

各判断の Reason / Rejected は `docs/decisions/`。閾値・入口の正は `DESIGN.md` / `DEPLOY.md`。

1. 音声の保存・配信形式は **mp3**（`contracts/drive-layout.md` / `2026-09-13T13-40-29`）
2. 着手順: mp3 → R2 → 薄い cache（`2026-09-13T13-41-00`）
3. 現行 storage runtime は Drive。将来 R2（`2026-09-13T14-22-55`）。cache は R2 後（`2026-09-13T14-23-30`）
4. generator 書込とは runtime 共有しない（読取専用）
