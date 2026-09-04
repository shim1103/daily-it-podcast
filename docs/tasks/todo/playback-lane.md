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
8. 原稿 body の opening / closing に seek 用 `startSec` を追加（3 bookend を「本文 + startSec」へ揃える）。判断は `docs/decisions/2026-09-04T16-44-46-feature-playback-topic-ending-startsec-contract.md`

### 未完了

（なし）

### 方針 index

各判断の Reason / Rejected は `docs/decisions/`。閾値・入口の正は `DESIGN.md` / `DEPLOY.md`。

音声は wav。generator 書込とは共有しない。
