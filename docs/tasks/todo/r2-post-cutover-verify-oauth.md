# chore(storage): R2 切替後の System/E2E 緑化と OAuth 削除を完了する

## 1. Summary

このIssueでは R2 を正本にした後、`generator-system` と `playback-e2e` を緑にし、Google OAuth / `DRIVE_*` を Worker・GHA・運用表から削除する。完了後、Drive credential に依存しない。

## 2. Context

1. 事実: 実施順の正は Decision `2026-09-14T12-49-26`（本 Issue は列 7）。
2. 事実: 正本切替は列 6（`r2-smoke-migrate-cutover`）済みが前提。
3. 事実: System は `TEST_*` → test bucket。E2E は本番 hostname + R2 上 fixture。

## 3. Canonical Sources

1. 順番: `docs/decisions/2026-09-14T12-49-26-docs-generator-r2-write-details.md`
2. 実施の形 / error: `2026-09-14T11-04-30` / `11-19-21`
3. 運用: `DEPLOY.md`
4. 前段: `r2-smoke-migrate-cutover`

## 4. Scope

### In Scope

1. `generator-system.yml`（dispatch 可）が R2 test bucket 前提で緑
2. `playback-e2e`（週次または dispatch）が R2 正本で緑
3. Worker から `GOOGLE_OAUTH_*` / `DRIVE_FOLDER_ID` を削除
4. GHA から不要になった Drive OAuth / folder 登録を削除（本番・TEST の該当分）
5. `DEPLOY.md` から Drive OAuth 行を除去し R2 のみにする最終確認

### Out of Scope

1. R2 Adapter 実装・人手移行・切替そのもの（列 4–6）
2. 登録手順の記述
3. cache

## 5. Contract

1. System 成功 postcondition が R2 test 空間で成立する
2. E2E が Access 付き本番 hostname で一覧・再生できる
3. production Worker / 該当 GHA が Google OAuth を要求しない

## 6. Constraints

1. System/E2E が赤のまま OAuth を消さない
2. 切替未完了（列 6 未完）で本 Issue を完了扱いにしない

## 7. Acceptance Criteria

1. [ ] `generator-system` が R2 test 前提で PASS
2. [ ] `playback-e2e` が R2 正本で PASS
3. [ ] Worker secrets/vars から OAuth / `DRIVE_*` が無い
4. [ ] GHA の該当 Drive credential が無い（または未使用で削除済み）
5. [ ] `DEPLOY.md` に Drive OAuth 必須行が残っていない

## 8. Verification

1. Actions run URL（system / e2e）
2. `wrangler secret` / Dashboard の欠落確認
3. `DEPLOY.md` の表

## 9. Dependencies

1. `r2-smoke-migrate-cutover` 完了
2. 次: cache は別 Decision（`2026-09-13T14-23-30`）・本列の後続

## 10. Notes

OAuth 削除は不可逆に近い。必ず System/E2E 緑の後。
