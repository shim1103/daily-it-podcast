## Playback 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md  
HTTP 契約の正: `apps/playback/contracts/`  
Drive 読みの正: `contracts/drive-layout.md`  
deploy・Access・GHA 運用の正: `DEPLOY.md`  
層規則・test 配置の正: `DESIGN.md`  
再発する判断の正: `docs/decisions/`

未完了の達成契約は `docs/tasks/todo/playback-*.md` が正。本 lane は進捗 index のみ。GitHub Issue 化しない運用。

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

### 未完了

（なし）

### 未決 index

1. Drive + OAuth をやめ **R2 へ完全移行**して refresh token 依存を無くすか。playback worker（読取経路）と generator（書込経路）の両方に跨る中規模移行。`2026-09-07T22-25-00` Decision の Rejected #3 から派生。着手判断は未。

### 方針 index

各判断の Reason / Rejected は `docs/decisions/`。閾値・入口の正は `DESIGN.md` / `DEPLOY.md`。

音声は wav。generator 書込とは共有しない。
