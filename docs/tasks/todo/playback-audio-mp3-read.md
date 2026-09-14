# feature(playback): 音声読取・HTTP を mp3 契約に揃える

## 1. Summary

このIssueでは playback が Drive 上の `{episodeId}.mp3` を読み、`Content-Type: audio/mpeg` で Range 対応のまま返す状態を、実装・test で完了させる。完了後、一覧の `audioRef` 経由の再生が mp3 契約で成立する。

## 2. Context

1. 事実: A が `episodeAudioContentType` / `episodeAudioFileExtension` と読取拡張子を mp3 側へ固定済み。
2. 事実: 安定 E2E fixture の物理 file はまだ `.wav`（差し替えは batch Issue）。
3. 事実: HTTP Range は維持（Decision `2026-09-04T23-16-00`）。
4. 運用: 1 Issue = 1 PR。本番同着切替と fixture 物理差し替えは列 3（`audio-mp3-cutover-migrate`）。

## 3. Canonical Sources

1. 契約: `contracts/episode-layout.md` / `apps/playback/contracts/http.ts`
2. 判断: `docs/decisions/2026-09-13T13-40-29-feature-playback-now-playing-audio-listenability.md`
3. Range: `docs/decisions/2026-09-04T23-16-00-feature-playback-e2e-redeploy-master.md`
4. test 方針: testing-strategy（再掲しない）

## 4. Scope

### In Scope

1. Drive 読取・HTTP・dev middleware / integration / SU が mp3 契約と一致していることの確認と不足修正
2. silent / fake 音声を契約ヘッダまたは再生に必要な形へ必要なら更新
3. A 足場に対する behavior の穴埋め

### Out of Scope

1. generator encode（`generator-audio-mp3-encode-write`）
2. 本番 Drive 上一括 wav→mp3 と同着切替（`audio-mp3-cutover-migrate`）
3. R2（列 4 以降。Decision `2026-09-14T12-49-26`）

## 5. Contract

1. `GET /episodes/:episodeId/audio` 成功時 `Content-Type` は `audio/mpeg`
2. 読取対象 file 名は `{episodeId}.mp3`
3. Range / `Accept-Ranges: bytes` は変えない

## 6. Constraints

1. wav/mp3 二系統 serve を入れない
2. Access / auth を変えない

## 7. Acceptance Criteria

1. [ ] worker / web の音声経路 test が `audio/mpeg` と `.mp3` で緑
2. [ ] Range の既存 SU が緑のまま
3. [ ] playback coverage gate が緑

## 8. Verification

1. `apps/playback` の lint / typecheck / vitest（project 既定）

## 9. Dependencies

1. mp3 A/B 済み
2. 順番: Decision `2026-09-14T12-49-26`（本 Issue は列 2）
3. 本番同着切替・batch・fixture は `audio-mp3-cutover-migrate`（列 3）

## 10. Notes

encode 未マージのまま本変更だけを本番に載せると拡張子と中身が食い違う。同着は列 3。
