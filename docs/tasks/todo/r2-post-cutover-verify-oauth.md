# chore(storage): R2 本番同着切替を完了し System/E2E 緑化と OAuth 削除まで進める

## 1. Summary

このIssueでは、人手移行済みの R2 データ（prod/test 両 bucket）の確認から、generator 書込・playback 読取の **R2 本番同着 deploy**、`generator-system` / `playback-e2e` の緑化、Google OAuth / `DRIVE_*` 削除までを完了する。完了後、正本 storage は R2 になり Drive credential に依存しない。

## 2. Context

1. 事実: 実施順の正は Decision `2026-09-14T12-49-26`。旧列 6（`r2-smoke-migrate-cutover`）は完了削除し、残タスクを本 Issue（旧列 7）へ統合した（`2026-09-16T14-12-56`）。
2. 事実: generator/playback の本番結線（composition の呼び出し側）は R2 へ完全置換済み（develop、`2026-09-16T10-45-14`）。Drive config・OAuth token 取得コード自体の削除は本 Issue の Scope。
3. 事実: R2 疎通確認（test bucket の Put/List/Get 往復）は dispatch で PASS 済み（run `35054835880`）。実装は一過性のため確認後に削除済み（`2026-09-16T13-46-41`）。
4. 事実: prod/test 両 bucket への人手移行は shim が手動で実施済み（put 完了）。ただし stem 数・pair 整合の確認はまだ行っていない。「配置」と「確認」を分けて扱う。
5. 事実: System は `TEST_*` → test bucket。E2E は本番 hostname + R2 上 fixture。

## 3. Canonical Sources

1. 順番: `docs/decisions/2026-09-14T12-49-26-docs-generator-r2-write-details.md`
2. 実施の形 / error: `2026-09-14T11-04-30` / `11-19-21`
3. 結線方式・疎通範囲: `2026-09-16T10-45-14`
4. 疎通確認実装の寿命: `2026-09-16T13-46-41`
5. Issue 統合（列6→本Issue）: `2026-09-16T14-12-56`
6. 依存実装: `generator-r2-write-adapter` / `playback-r2-read-adapter`
7. 配置: `contracts/episode-layout.md`
8. 運用: `DEPLOY.md`

## 4. Scope

### In Scope

1. prod/test bucket に人手移行済みのデータを、R2 Dashboard または List で stem 数・pair 整合の目視確認をする
2. generator 書込と playback 読取を **分けて本番切替しない**同着 deploy（develop → master → 本番）
3. `DEPLOY.md` の storage 表を R2 前提へ更新する（Drive OAuth 行はこの Issue 完了時点で除去）
4. 本番 hostname での一覧・再生の人手 smoke
5. `generator-system.yml`（dispatch 可）が R2 test bucket 前提で緑
6. `playback-e2e`（週次または dispatch）が R2 正本で緑
7. Worker から `GOOGLE_OAUTH_*` / `DRIVE_FOLDER_ID` を削除
8. GHA から不要になった Drive OAuth / folder 登録を削除（本番・TEST の該当分）

### Out of Scope

1. R2 Adapter 実装・結線切替そのもの（develop へ merge 済み）
2. Drive→R2 の人手移送そのもの（shim 実施済み）・移送の script 化
3. 登録手順の記述
4. cache

## 5. Contract

1. prod bucket に人手移行した完成ペアが約 10 stem あることを確認済み
2. 本番 runtime の正本が R2（書込・読取とも）になる
3. 公開直 URL / `r2.dev` に依存しない
4. System 成功 postcondition が R2 test 空間で成立する
5. E2E が Access 付き本番 hostname で一覧・再生できる
6. production Worker / 該当 GHA が Google OAuth を要求しない

## 6. Constraints

1. mp3 列（`audio-mp3-cutover-migrate`）は完了済み
2. 書込だけ / 読取だけを本番に載せない
3. System/E2E が赤のまま OAuth を消さない
4. OAuth 削除は不可逆に近い。必ず deploy 後の同着確認・System/E2E 緑化の後に行う

## 7. Acceptance Criteria

1. [ ] prod bucket の stem 数・pair 整合を R2 Dashboard/List で確認した
2. [ ] generator と playback が同着で R2 を正本にしている（deploy 済み）
3. [ ] 本番 hostname での一覧・再生の人手 smoke が通る
4. [ ] `DEPLOY.md` が R2 切替後の latest に更新されている
5. [ ] `generator-system` が R2 test 前提で PASS
6. [ ] `playback-e2e` が R2 正本で PASS
7. [ ] Worker secrets/vars から OAuth / `DRIVE_*` が無い
8. [ ] GHA の該当 Drive credential が無い（または未使用で削除済み）
9. [ ] `DEPLOY.md` に Drive OAuth 必須行が残っていない

## 8. Verification

1. R2 Dashboard または List での stem 数観測
2. 本番 hostname での一覧・再生の人手 smoke
3. Actions run URL（system / e2e）
4. `wrangler secret` / Dashboard の欠落確認
5. `DEPLOY.md` の表

## 9. Dependencies

1. `audio-mp3-cutover-migrate` 完了済み
2. `generator-r2-write-adapter` / `playback-r2-read-adapter` 完了済み
3. 次: cache は別 Decision（`2026-09-13T14-23-30`）・本列の後続

## 10. Notes

旧 lane 名 `R-r2-cutover` の同着切替達成、および旧列 6（`r2-smoke-migrate-cutover`）の残タスクは本 Issue に吸収した。

Drive→R2 の人手移送・確認は script 化しない（`2026-09-14T12-49-26` / `2026-09-16T10-45-14`）。実行時の具体手順は本 Issue に固定せず、実行 session の `docs/daily/` に書く。

OAuth 削除は不可逆に近い。必ず deploy 後の同着確認・System/E2E 緑の後。
