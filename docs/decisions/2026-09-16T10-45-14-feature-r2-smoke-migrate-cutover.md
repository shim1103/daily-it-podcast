---
name: 列6の本番切替はDrive呼び出しをR2へ完全置換。疎通dispatchはtest bucketのPut/List/Get往復のみを見る
date: 2026-09-16T10:45:14
branch: feature/r2-smoke-migrate-cutover
---

## 1. Decision

1. **結線方式**: generator（`newProduceEpisode`）・playback（本番 route の `createPlaybackControllers` 呼び出し）とも、Drive 呼び出しを **R2 へ完全置換**する。mode 切替パラメータで Drive/R2 を並行に呼べる状態は残さない。config・OAuth token 取得（`cfg.Drive` / `DriveConfig`）自体の削除は列7（`r2-post-cutover-verify-oauth`）の責務とし、本 Issue（列6）では composition の呼び出し側だけを置換する。
2. **疎通 `workflow_dispatch` の検証範囲**: test bucket に対する Put/List/Get の往復と、playback 側 `EPISODES` binding 経由の get/list が通ることだけを見る。stem pair の整合・不完全ペア許容・schema 適合は見ない。test/prod を引数で切り替える口は持たない（test 固定）。

## 2. Reason

1. 先行 Decision `2026-09-14T11-04-30` §5 が「storage の二次系 fallback / Drive+R2 二重正本は採らない」と既に定めている。mode 切替パラメータを本番経路に残すと、列7 の OAuth 削除までの間 Drive/R2 二重正本の状態が生き続け、この先行判断と衝突する。切替後に問題が出た場合の復旧は `git revert` + `wrangler rollback`（`DEPLOY.md` §7 既定手順）で足りる。
2. 先行 Decision `2026-09-14T12-49-26` §4 が「疎通確認は Issue 6 内の `workflow_dispatch` 専用とし、`generator-system` / `playback-e2e` とは別にする。System / E2E は Issue 7 が正」と既に定めている。stem pair 整合や不完全ペア許容の意味論検証は Application 層（`episode-layout.md` の読み/書き契約）と System/E2E の責務であり、疎通（生 I/O 呼び出しの経路確認）がそこへ踏み込むと責務が重複する。
3. test bucket には人手移行対象の実 episode データが無い（人手移行は Scope②で prod bucket が対象）。test bucket 上で pair 件数や整合を検証しようとしても検証対象が存在しない。
4. test/prod を 1 本の dispatch で切り替え可能にすると、疎通目的の workflow が prod credential にも触れる経路を持つことになる。Decision `2026-09-14T11-04-30` §2 の「test/prod は別 bucket」という隔離方針に対し、疎通目的の範囲を超えて prod への到達手段を増やす。prod 側の確認は Scope⑥（R2 Dashboard/List 目視 + 本番 hostname の人手 smoke）に既に別経路がある。

## 3. Rejected

1. **Drive/R2 を mode パラメータで並行に呼べる結線を残す案** — 障害時に単一 deploy で Drive へ戻せる利点はあるが、Constraints②「書込だけ/読取だけを本番に載せない」との整合を保つコストが増え、列7 の OAuth 削除まで二重正本状態を維持する理由が無い。
2. **疎通 dispatch で test bucket 上の stem pair 整合まで検証する案** — test bucket に実データが無く検証対象が存在しない。意味論検証は System/E2E（列7）の役割であり、先行 Decision `12-49-26` の責務分離と衝突する。
3. **疎通 dispatch に test/prod 切替引数を持たせる案** — 疎通目的に対して prod credential への到達経路を増やすのは過剰（YAGNI）。prod 確認は Scope⑥の人手 smoke で足りる。
