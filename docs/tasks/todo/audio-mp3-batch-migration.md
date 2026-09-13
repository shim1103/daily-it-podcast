# chore(audio): 既存 WAV を一括 mp3 化し安定 fixture を差し替える

## 1. Summary

このIssueでは Drive（および repo 内安定 fixture）上の既存 `{episodeId}.wav` を一括で `{episodeId}.mp3` に変換・置換し、旧 wav を残さない状態にする。完了後、本番読取と週次 E2E が mp3 のみを見る。

## 2. Context

1. 事実: Decision は一括 batch migration（二系統 serve なし）（`2026-09-13T13-40-29`）。
2. 事実: encode 手段は ffmpeg（encode-write Issue と同じ）。
3. 運用: 1 Issue = 1 PR。release は lane `R-mp3-migrate`。

## 3. Canonical Sources

1. 判断: `docs/decisions/2026-09-13T13-40-29-feature-playback-now-playing-audio-listenability.md`
2. 配置: `contracts/drive-layout.md`
3. fixture: `apps/playback/test/e2e/fixtures/stable-episode/README.md`
4. 運用: `DEPLOY.md`（値は写さない）

## 4. Scope

### In Scope

1. 既存 wav → mp3 の一括変換（script または文書化した ffmpeg 手順）
2. 変換後の stem 一致・非空・再生煙（またはマジックバイト）の確認
3. 旧 `.wav` を残さない
4. 安定 E2E fixture の物理 file と README を `.mp3` に更新

### Out of Scope

1. encode 本実装（encode-write Issue）
2. R2 / cache（Decision のみ。本 Issue にしない）

## 5. Contract

1. 移行後、対象 folder 直下の完成ペアは `{id}.json` + `{id}.mp3` のみ
2. 同一 stem の `.wav` は残らない

## 6. Constraints

1. lazy（GET 時変換）を本番経路に入れない
2. encode-write と異なる encoder 品質設定を黙って使わない

## 7. Acceptance Criteria

1. [ ] 対象 folder に移行対象の `.wav` が残っていない
2. [ ] 安定 fixture が `.mp3` として README・実 file 一致
3. [ ] 週次 E2E または人手 smoke で一覧・再生が通る

## 8. Verification

1. Drive（または fixture dir）の file 一覧観測
2. 可能なら `playback-e2e` / 人手再生 smoke

## 9. Dependencies

1. `generator-audio-mp3-encode-write`（ffmpeg encode が使える）
2. `playback-audio-mp3-read`（読取が `.mp3`）
3. 両者が同じ release に載った後に本番 folder を触る

## 10. Notes

batch は不可逆に近い（旧 wav 削除）。実行前に対象 stem 一覧を固定する。
