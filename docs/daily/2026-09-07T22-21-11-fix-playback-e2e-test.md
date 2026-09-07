---
name: playback-e2e の週次失敗を Google OAuth refresh token 失効と特定・OAuth を Production 化で恒久復旧・DriveError の切り分け粒度を修正
date: 2026-09-07T22:21:11
session_id: 2026-09-07T22-21-11
branch: fix/playback-e2e-test
prev: なし
---

## 1. Summary

週次 `playback-e2e`（`run 34067247581`、schedule）が 4 test 全滅していた原因を調査した。全 test が `page.goto("/")` 後の `.episode-list` not found で timeout していた。前回の schedule 成功（2026-08-31）以降 `master` に入ったのは Preview URL 無効化と release merge のみで、frontend / e2e spec / fixture は健全。手動 dispatch でも再現したため schedule 固有要因を排除。Worker Observability log で `GET /episodes` が `503`・`DriveError: "Drive HTTP 呼び出しが非 2xx: 400"` と判明し、Access は通過済み・storageState は生存と確定。`google-drive-episode-repository.ts` を読み、`request()` 共通経路が OAuth token 取得・files.list・download の 3 呼び出しを同一 message へ畳んでおり、`400` が token endpoint の `invalid_grant` か files.list の q 文法エラーか log から切り分け不能と特定。shim が OAuth consent screen を Testing → Production へ変更し refresh token を再発行、Worker Secret へ反映。`/episodes` が `200` を返すのを確認し、`playback-e2e` を再 dispatch して緑（`run 34126139797`）。切り分け不能問題の恒久対策として `request()` へ `DriveOperation` label（"OAuth token" / "files.list" / "file download"）を必須追加し message を種別付きにする修正を TDD で入れ、`wrangler deploy`（`Version ca7864ba`）後に e2e 再緑（`run 34126335119`）を確認した。

## 2. Changes

1. 調査対象 CI: `run 34067247581`（schedule）。手動再現: `run 34119303180`。OAuth 修正後の確認: `run 34126139797`。deploy 後の再確認: `run 34126335119`。すべて `master` ref。
2. Worker Observability log の error chain: `listEpisodesController` → `DriveError("Drive HTTP 呼び出しが非 2xx: 400")` → `UnavailableError("利用できない")` → `503`。`wallTimeMs` 約 120 で即時 503（timeout ではない）。`cf-access-authenticated-user-email` / `cf-access-jwt-assertion` が付いており Access 通過は確定。
3. `wrangler tail` で本番 `/episodes` を観測。修正前は `DriveError` の `cause.message: "Drive HTTP 呼び出しが非 2xx: 400"`。修正後は `logs: []` で `status: 200`（2 回とも）。`/favicon.ico` の 400（契約に無い path）は既存の別件で本 scope 外。
4. shim の OAuth 対応（session 外）: consent screen Testing → Production、`GOOGLE_OAUTH_REFRESH_TOKEN` 再発行、`wrangler secret put` で Worker へ反映。判断は `docs/decisions/2026-09-07T22-25-00-fix-playback-e2e-test.md`。
5. `commit 8d37082`: `request()` へ `DriveOperation` 必須引数を追加。message を `Drive {operation} 呼び出しが非 2xx: {status}` へ変更。3 call site（`fetchAccessToken` / `queryFolderEntries` / `downloadBytes`）へ label 付与。sociable_unit test に「非 2xx の DriveError message は失敗した Drive 呼び出しを名指しする」describe を追加（RED → GREEN）。unit 365 pass / coverage 100% 維持、typecheck / biome lint・format / depcruise 全緑、pre-push で integration 29 pass。
6. `wrangler deploy` 実行。`Version ID ca7864ba-3470-4b31-ab8b-cb07220a1399`。deploy diff に `DRIVE_FOLDER_ID` / `GOOGLE_OAUTH_CLIENT_ID` が local config に無い旨の警告が出たが、`keep_vars: true` により削除されず（deploy 後 e2e 緑が保持を裏付け）。
7. R2 への完全移行（Drive + OAuth 依存の廃止）は shim との会話で中期案として挙がったが未着手。`docs/tasks/todo/playback-lane.md` の未決 index へ記載。

### Commits

- `8d37082`
