# feature(storage): R2 疎通・人手移行・本番同着切替を完了する

## 1. Summary

このIssueでは R2 Adapter（列 4・5）実装後に、`workflow_dispatch` による疎通確認、Drive からの人手 ~10 episode 移行、generator 書込と playback 読取の **R2 本番同着切替**までを完了する。完了後、正本 storage は R2 になる（OAuth 削除は次列）。

## 2. Context

1. 事実: 実施順の正は Decision `2026-09-14T12-49-26`（本 Issue は列 6）。
2. 事実: R2 登録は完了済み。登録手順は本 Issue に書かない。
3. 事実: 移送は人手 ~10（script / Super Slurper しない）。
4. 事実: 疎通は System / E2E とは別の dispatch。
5. 運用: release 専用の空 Issue ではない。切替の達成契約を持つ。

## 3. Canonical Sources

1. 順番: `docs/decisions/2026-09-14T12-49-26-docs-generator-r2-write-details.md`
2. 実施の形: `docs/decisions/2026-09-14T11-04-30-docs-generator-r2-write-details.md`
3. error / retry: `docs/decisions/2026-09-14T11-19-21-docs-generator-r2-write-details.md`
4. 依存実装: `generator-r2-write-adapter` / `playback-r2-read-adapter`
5. 配置: `contracts/episode-layout.md`
6. 運用 latest: `DEPLOY.md`（切替時に差し替え開始。値の百科はここに書かない）

## 4. Scope

### In Scope

1. `workflow_dispatch` による R2 疎通（test bucket Put/List/Get、binding read の最小確認）
2. Drive から人手で約 10 episode を prod bucket へ配置（`{id}.json` + `{id}.mp3`）
3. generator 書込と playback 読取を **分けて本番切替しない**同着 deploy
4. `DEPLOY.md` の storage 表を R2 前提へ差し替え開始（OAuth 行の削除は次列と整合）

### Out of Scope

1. Adapter 本実装（列 4・5）
2. System / E2E の正式緑化と OAuth 削除（列 7）
3. R2 登録手順の記述
4. cache
5. mp3 未完了のまま切替

## 5. Contract

1. 疎通 dispatch が test bucket で成功する（System / E2E に載せない）
2. prod bucket に人手移行した完成ペアが約 10 stem ある
3. 本番 runtime の正本が R2（書込・読取とも）になる
4. 公開直 URL / `r2.dev` に依存しない

## 6. Constraints

1. mp3 列（`audio-mp3-cutover-migrate`）完了前に本 Issue を完了扱いにしない
2. 書込だけ / 読取だけを本番に載せない
3. OAuth を本 Issue 完了条件に含めない（次列）

## 7. Acceptance Criteria

1. [ ] 疎通用 `workflow_dispatch` が成功している（System/E2E ではない）
2. [ ] prod に人手移行 ~10 stem の json+mp3 がある
3. [ ] generator と playback が同着で R2 を正本にしている
4. [ ] `DEPLOY.md` が R2 切替後の latest に更新されている（OAuth 削除の最終確認は列 7）

## 8. Verification

1. dispatch ログ
2. R2 Dashboard または List での stem 数観測
3. 本番 hostname での一覧・再生の人手 smoke（正式 E2E は列 7）

## 9. Dependencies

1. `audio-mp3-cutover-migrate` 完了
2. `generator-r2-write-adapter` / `playback-r2-read-adapter` 実装完了
3. 次列: `r2-post-cutover-verify-oauth`

## 10. Notes

旧 lane 名 `R-r2-cutover` の同着切替達成は本 Issue に吸収した。OAuth 削除は意図的に次へ送る。
