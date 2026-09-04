---
name: Playback Worker の rollback 一次手段は wrangler rollback（全量）
date: 2026-09-04T02:04:00
branch: feature/playback-deploy-ops-followup
---

## 1. Decision

1. 悪い deploy を戻す一次手段は、現行 CLI の `wrangler versions list` で直前の Version を特定し、`wrangler rollback` で **全量** 戻すこととする。
2. 二次手段は、既知の good commit を checkout して通常の `wrangler deploy` をやり直すこととする。
3. Variable / Secret / Access は Worker Version と別 lifecycle とし、rollback の対象にしない。手順の正本は `DEPLOY.md`。契約値（Worker `name` 等）は `apps/playback/wrangler.jsonc` を参照し、本 Decision に写さない。

## 2. Reason

1. 現行 wrangler が Version 単位の rollback を提供する。日常の「壊れた直前へ戻す」には、git 履歴を辿り直すより Version ID 指定の方が操作が短い。
2. 本番 hostname 以外の preview / version URL は共有しない（先行 Decision `2026-08-25T17-10-00-feature-playback-worker-deploy.md`）。traffic を複数 Version に割る日常 rollback は、その公開境界と衝突しやすい。
3. OAuth・Drive・Access の設定ミスは code Version を戻しても直らない。対象を混ぜると「rollback したのに直らない」が再現し、Least Astonishment に反する。

## 3. Rejected

1. `wrangler versions deploy` の traffic split を日常 rollback にする案 — version URL / 分割配信が Access・共有方針と噛み合わない。
2. git revert + 再 deploy だけを一次手段にする案 — Version 履歴が Cloudflare 側にあるのに、毎回 local tree 操作を必須にすると手順が長くなる。二次には残す。
3. Secret / Access まで同一 rollback 手順に含める案 — lifecycle が違い、失敗時の切り分けが壊れる。
