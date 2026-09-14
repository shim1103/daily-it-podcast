# chore(audio): mp3 本番同着切替と既存 wav 一括移行・fixture 差し替えを完了する

## 1. Summary

このIssueでは `generator-audio-mp3-encode-write` と `playback-audio-mp3-read` を **同着で本番に載せ**、続けて既存 `{episodeId}.wav` を `{episodeId}.mp3` へ一括置換し、安定 E2E fixture を mp3 にする。完了後、Drive 上の完成ペアと週次 E2E は mp3 のみを見る。

## 2. Context

1. 事実: 実施順の正は Decision `2026-09-14T12-49-26`（本 Issue は列の 3）。
2. 事実: encode / 読取の実装 Issue（列 1・2）が先。本 Issue は **release 専用の空 Issue ではない**（切替＋batch＋fixture の達成契約を持つ）。
3. 事実: 形式方針は `2026-09-13T13-40-29`（一括 migration・二系統 serve なし）。
4. 運用: 1 Issue = 1 PR に限らない場合あり（deploy 手順＋migration）。完了は Acceptance 全体。

## 3. Canonical Sources

1. 順番: `docs/decisions/2026-09-14T12-49-26-docs-generator-r2-write-details.md`
2. 形式: `docs/decisions/2026-09-13T13-40-29-feature-playback-now-playing-audio-listenability.md`
3. 配置: `contracts/episode-layout.md`
4. 依存実装: `generator-audio-mp3-encode-write` / `playback-audio-mp3-read`
5. fixture: `apps/playback/test/e2e/fixtures/stable-episode/README.md`
6. 運用: `DEPLOY.md`（値は写さない）

## 4. Scope

### In Scope

1. encode-write と mp3-read を **分けて本番 deploy しない**同着切替
2. 既存 wav → mp3 の一括変換（script または文書化した ffmpeg 手順）
3. 変換後の stem 一致・非空・旧 `.wav` 非残存
4. 安定 E2E fixture の物理 file と README を `.mp3` に更新
5. 切替後の一覧・再生煙（E2E または人手）

### Out of Scope

1. encode / 読取の本実装（列 1・2）
2. R2 登録・Adapter・OAuth 削除（列 4 以降）
3. cache

## 5. Contract

1. 本番に載った新作 episode の音声は非空 `{id}.mp3`（`audio/mpeg`）
2. 移行後、対象 folder 直下の完成ペアは `{id}.json` + `{id}.mp3` のみ。同一 stem の `.wav` は残らない
3. 安定 fixture が `.mp3` として README・実 file 一致

## 6. Constraints

1. lazy（GET 時変換）を本番経路に入れない
2. encode-write と異なる encoder 品質設定を黙って使わない
3. 読取未追随のまま encode だけを本番に載せない（逆も同様）

## 7. Acceptance Criteria

1. [ ] encode-write と mp3-read が同着で本番に載っている
2. [ ] 対象 folder に移行対象の `.wav` が残っていない
3. [ ] 安定 fixture が `.mp3` として README・実 file 一致
4. [ ] 週次 E2E または人手 smoke で一覧・再生が通る

## 8. Verification

1. 本番 deploy の同着確認
2. Drive（または fixture dir）の file 一覧観測
3. `playback-e2e` または人手再生 smoke

## 9. Dependencies

1. `generator-audio-mp3-encode-write` 完了
2. `playback-audio-mp3-read` 完了
3. 次列: `generator-r2-write-adapter` / `playback-r2-read-adapter`（本 Issue 完了後）

## 10. Notes

batch は不可逆に近い（旧 wav 削除）。実行前に対象 stem 一覧を固定する。旧 lane 名 `R-mp3-cutover` / `R-mp3-migrate` の達成は本 Issue に吸収した。

移行入口（人手 download/upload は Scope 本文。secret 値は書かない）:

1. local encode: `scripts/generator/encode-cache-wav-to-mp3.sh`（`.cache/<prod|test>` → 本番 `Encoder`）
2. Drive purge+verify: `scripts/generator/drive-purge-wav-verify-set.sh` / `generator-drive-purge-wav-verify-set.yml`（`workflow_dispatch`・`target=prod|test`）
