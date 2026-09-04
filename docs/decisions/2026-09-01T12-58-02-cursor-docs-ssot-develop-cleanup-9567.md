---
name: DEPLOY の実装到達済み caveat は残さない
date: 2026-09-01T12:58:02
branch: cursor-docs-ssot-develop-cleanup-9567
---

## 1. Decision

運用 SSOT（`DEPLOY.md`）に書いた **「未実装のため失敗しうる」等の一時 caveat** は、実装・登録・定時運用が到達したら削除する。lane の済み要約と矛盾する文言は DEPLOY に残さない。

## 2. Reason

1. `DEPLOY.md` は継続運用の latest SSOT（`2026-08-25T16-57-00`）。`ProduceEpisode.Run` 未完の注記は実装前の注意であり、Run / Broad / produce workflow 到達後は誤読を生む。
2. 進捗の済み/未は `docs/tasks/todo/*-lane.md` が index。DEPLOY に完了ナレーションや未完了 caveat を二重に書くと、どちらが正かがずれる（`2026-08-30T16-20-04` が進捗二重化を却下したのと同型）。
3. cron 暦日や `workflow_dispatch` の前提（default branch が develop）は運用上必要なので残す。消すのは **実装フェーズに紐づく一時警告のみ**。

## 3. Rejected

1. caveat を「念のため」残し続ける案 — 定時 produce が通る状態で毎回読者が無視する注意になり、本当の障害と区別できなくなる。
2. 未完了 caveat を lane へ移すだけで DEPLOY に残す案 — 運用文書に未実装フラグが残り、GHA 定時の期待と矛盾する。
3. Secret 登録状況の完了表を DEPLOY に新設する案 — inventory の latest は必要なら `DEPLOY` §4–5 の登録名で足りる。完了日付ナレーションは code / lane と二重になる（今回の doc 整理方針 KISS に反する）。
