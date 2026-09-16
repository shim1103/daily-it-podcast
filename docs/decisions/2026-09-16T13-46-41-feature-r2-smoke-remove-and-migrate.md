---
name: R2疎通確認のyml・実装は一度のPASS確認後に削除する。常設gateとして残さない
date: 2026-09-16T13:46:41
branch: feature/r2-smoke-remove-and-migrate
---

## 1. Decision

疎通 `workflow_dispatch`（`generator-r2-smoke.yml`・`smoke.go` 等）は、PASS を1度確認した後、実装一式を削除する。常設 gate として repo に残さない。

## 2. Reason

1. 疎通確認の目的は「R2 S3 互換 API への経路が生きているか」の一度の確認であり（Decision `2026-09-16T10-45-14` §1-2）、継続的な回帰検出ではない。目的を達成した後も実装を残すと、目的を持たない dead code が repo に残る。
2. 本番結線（generator/playback の composition）自体は Sociable Unit / Narrow Integration / Broad Integration で継続的に検証される（既存 test 群）。疎通確認はそれらでは検証できない「実 R2 credential での生 I/O 往復」という一過性の環境確認であり、Adapter の振る舞い保証は疎通確認に依存しない。
3. 常設で残す場合、cron 化しない dispatch 専用 test を repo に維持するコストが生じる（credential 登録・build tag 分離・実行のtriage）。1度確認できれば十分な性質のものにこのコストを払う理由が無い。

## 3. Rejected

1. **常設 dispatch として repo に残す案** — 再実行の必要が生じた際に再実装するコストより、常設維持コストの方が高い。疎通確認は生 I/O 経路が変わった時（S3 互換 API の endpoint 変更等）以外に再実行する動機が薄く、その時点で最小構成を再実装すれば足りる。
2. **削除せず既定 build から除外したまま放置する案**（`r2smoke` タグのみで凍結） — 動かないcode が repo に残ると、次に読む人が「なぜ実行されないtestがあるか」を都度調査するコストを払う。目的を終えた実装は消す方が読み手にとって単純（KISS）。
