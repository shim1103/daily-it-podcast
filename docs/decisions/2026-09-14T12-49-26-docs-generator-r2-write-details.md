---
name: storage 切替の実施順は mp3 完了→R2 Adapter→疎通→人手移行→R2 cutover→System/E2E→OAuth 削除。登録は先行済み
date: 2026-09-14T12:49:26
branch: docs/generator-r2-write-details
---

## 1. Decision

1. 音声 storage まわりの **実施順**を次の Issue file 単位に固定する（release 専用 Issue は作らない）。
   1. `generator-audio-mp3-encode-write`
   2. `playback-audio-mp3-read`
   3. `audio-mp3-cutover-migrate`（旧 lane の `R-mp3-cutover` + `R-mp3-migrate` + batch 達成を 1 Issue に束ねる）
   4. `generator-r2-write-adapter`
   5. `playback-r2-read-adapter`
   6. `r2-smoke-migrate-cutover`（`workflow_dispatch` 疎通・人手 ~10 移行・R2 本番同着切替）
   7. `r2-post-cutover-verify-oauth`（System / E2E 緑化・OAuth / `DRIVE_*` 削除）
2. **R2 登録（旧 S0）**は本 Decision 時点で完了済みとする。登録手順の百科は docs に書かない。credential の latest 置き場は `DEPLOY.md`（切替後に差し替え）。
3. 既存 object の Drive→R2 移送は **人手 ~10 episode** とする。script / Super Slurper は採らない（YAGNI）。
4. **疎通確認**は Issue 6 内の `workflow_dispatch` 専用とし、`generator-system` / `playback-e2e` とは別にする。System / E2E は Issue 7（R2 本番切替後）が正。
5. 着手順の大枠（mp3 → R2 → cache）は先行 `2026-09-13T13-41-00` を維持する。本 file はその **Issue 分割と前後関係**だけを具体化する。
6. 進捗 index の正は `docs/tasks/todo/*-lane.md`。lane は本 Decision の Issue 列を掲載し、手順本文は各 Issue へ委譲する。

## 2. Reason

1. mp3 未完了のまま R2 に載せると形式と storage の二変数が同時に動き、失敗切り分けが壊れる（先行着手順と同旨）。
2. release 専用 Issue は達成契約が薄く、Adapter Issue と二重になる。cutover / migrate の完了条件は「何を達成するか」を持つ Issue に載せる。
3. 件数 ~10・手動 download が容易なら移送専用実装は過剰。手段を決めて D を閉じる。
4. System / E2E を cutover 前に R2 へ向けると、未切替の赤/緑が運用を誤らせる。dispatch 疎通と正式 gate を分ける。
5. OAuth 削除を Adapter PR と同着にすると、切替検証前に credential を失う。検証緑の後に削除する。

## 3. Rejected

1. release だけの Issue（`R-mp3-*` / `R-r2-*` を空の達成契約にする案）— 契約が無く進捗だけが増える。
2. Super Slurper / 移送 script を必須にする案 — Drive 非対応または過剰（YAGNI）。
3. System / E2E を R2 疎通の唯一手段にする案 — cutover 前の正式 gate と混ざる。
4. OAuth を Adapter merge 直後に消す案 — 切替失敗時に Drive へ戻せない。
5. 登録手順を Decision / Issue に百科する案 — 運用 latest と手順が二重になる。登録は完了済み前提。
