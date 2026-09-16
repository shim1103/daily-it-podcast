# feature(storage): R2 疎通・人手移行・本番同着切替を完了する

## 1. Summary

このIssueでは R2 Adapter（列 4・5）実装後に、`workflow_dispatch` による疎通確認、Drive からの人手 ~10 episode 移行、generator 書込と playback 読取の **R2 本番同着切替**までを完了する。完了後、正本 storage は R2 になる（OAuth 削除は次列）。

## 2. Context

1. 事実: 実施順の正は Decision `2026-09-14T12-49-26`（本 Issue は列 6）。
2. 事実: R2 登録は完了済み。登録手順は本 Issue に書かない。
3. 事実: 移送は人手 ~10（script / Super Slurper しない）。移動・移行後の確認とも人手（`2026-09-16T10-45-14`）。
4. 事実: 疎通は System / E2E とは別の dispatch。test bucket の Put/List/Get 往復のみを見る。stem pair 整合・schema 適合は見ない。test/prod 切替引数は持たない（`2026-09-16T10-45-14`）。
5. 事実: generator（`newProduceEpisode`）・playback（本番 route の `createPlaybackControllers` 呼び出し）とも、現行は Drive 呼び出し固定。R2 側 Adapter 本体・factory 関数（`newR2WriteEpisode` 等）・`R2EpisodeRepository` は実装済みだが未結線。本 Issue で Drive 呼び出しを R2 へ完全置換する（`2026-09-16T10-45-14`）。config・OAuth token 取得自体の削除は列7。
6. 運用: release 専用の空 Issue ではない。切替の達成契約を持つ。

## 3. Canonical Sources

1. 順番: `docs/decisions/2026-09-14T12-49-26-docs-generator-r2-write-details.md`
2. 実施の形: `docs/decisions/2026-09-14T11-04-30-docs-generator-r2-write-details.md`
3. error / retry: `docs/decisions/2026-09-14T11-19-21-docs-generator-r2-write-details.md`
4. 結線方式・疎通範囲: `docs/decisions/2026-09-16T10-45-14-feature-r2-smoke-migrate-cutover.md`
5. 依存実装: `generator-r2-write-adapter` / `playback-r2-read-adapter`
6. 配置: `contracts/episode-layout.md`
7. 運用 latest: `DEPLOY.md`（切替時に差し替え開始。値の百科はここに書かない）

## 4. Scope

### In Scope

1. generator の本番結線を Drive から R2 へ完全置換（`newProduceEpisode` 内の `newGoogleDriveCompletedEpisodeLookup` / `newGoogleDriveWriteEpisode` 呼び出しを `newR2CompletedEpisodeLookup` / `newR2WriteEpisode` へ差し替え）
2. playback の本番 route を R2 固定へ完全置換（`app.ts` の `createPlaybackControllers` 呼び出しが R2 mode を選ぶよう変更）
3. `workflow_dispatch` による R2 疎通確認（test bucket Put/List/Get、binding read の最小確認。System/E2E とは別 gate）を **一度実行し PASS を確認する**。常設 gate ではないため、確認後は yml・実装（`smoke.go` 等）を削除する
4. Drive から人手で約 10 episode を prod bucket へ配置（`{id}.json` + `{id}.mp3`）
5. generator 書込と playback 読取を **分けて本番切替しない**同着 deploy
6. `DEPLOY.md` の storage 表を R2 前提へ差し替え開始（OAuth 行の削除は次列と整合）

### Out of Scope

1. R2 Adapter 本体・factory 関数の新規実装（列 4・5 で実装済み。本 Issue は既存 factory の呼び出し側切替のみ）
2. Drive config・OAuth token 取得コードの削除（列 7）
3. System / E2E の正式緑化と OAuth 削除（列 7）
4. R2 登録手順の記述
5. cache
6. mp3 未完了のまま切替
7. Drive→R2 移送・移行後確認の script 化（人手のまま。`2026-09-14T12-49-26` / `2026-09-16T10-45-14`）

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

1. [x] generator の本番結線が R2 呼び出しへ完全置換されている（Drive 呼び出しが `newProduceEpisode` に残っていない）
2. [x] playback の本番 route が R2 mode 固定で `createPlaybackControllers` を呼んでいる
3. [x] 疎通用 `workflow_dispatch` が成功した（System/E2E ではない。test bucket 限定。実行後 yml・実装は削除済み）
4. [ ] prod に人手移行 ~10 stem の json+mp3 がある
5. [ ] generator と playback が同着で R2 を正本にしている
6. [ ] `DEPLOY.md` が R2 切替後の latest に更新されている（OAuth 削除の最終確認は列 7）

## 8. Verification

1. generator/playback の R2 結線 diff（Drive 呼び出しの削除を確認）
2. dispatch 実行ログ（[run 35054835880](https://github.com/shim1103/daily-it-podcast/actions/runs/35054835880)、`TestR2Smoke_roundTripsProbeObject_overPutListGet` PASS）
3. R2 Dashboard または List での stem 数観測
4. 本番 hostname での一覧・再生の人手 smoke（正式 E2E は列 7）

## 9. Dependencies

1. `audio-mp3-cutover-migrate` 完了
2. `generator-r2-write-adapter` / `playback-r2-read-adapter` の Adapter 本体実装完了（結線切替は本 Issue のスコープ）
3. 次列: `r2-post-cutover-verify-oauth`

## 10. Notes

旧 lane 名 `R-r2-cutover` の同着切替達成は本 Issue に吸収した。OAuth 削除は意図的に次へ送る。

Drive→R2 の人手移送は script 化しない（`2026-09-14T12-49-26` / `2026-09-16T10-45-14`）。実行時の具体手順（download 元・一時配置・upload 手段）は本 Issue に固定せず、実行 session の `docs/daily/` に書く。

疎通確認の yml（`generator-r2-smoke.yml`）・実装（`smoke.go` 等）は一度の PASS 確認後に削除した。常設 gate ではなく一過性の疎通確認のため、確認後に残す価値が無い。
