---
name: Playback Integration gate と E2E の収集・定時・分類語
date: 2026-08-30T16:20:00
branch: docs/playback-integration-e2e-plan
---

## 1. Decision

1. Playback の Integration gate（pre-push / GHA Integration / `scripts/playback/test-integration.sh`）が実行してよいのは、**credential / secret 値を使わない Narrow Integration と Broad Integration** とする。browser E2E は載せない。
2. browser E2E の収集・実行入口は Integration と分け、必須 Unit / Integration gate に載せない。入口の正は code（`scripts/playback/test-e2e.sh`・`playback-e2e.yml`）と `DEPLOY.md`。
3. E2E workflow は generator System と同型に、**月曜 07:00 Asia/Tokyo**（cron UTC は `DEPLOY.md`）+ `workflow_dispatch` とする。
4. Playback browser E2E の file 名分類語は **`e2e`**。Generator の `system` と混線させない。Vitest Integration project は `system_e2e` を収集しない。

## 2. Reason

1. secret なし Narrow / Broad は Repeatable であり、配線回帰を PR 前に落とせる。E2E を混ぜると Access / session 起因の赤と配線赤が同じ gate で判別不能になる（generator `2026-08-30T11-56-00` / `11-56-01` と同型）。
2. 定時を毎 PR に載せると Actions 分と Access session 依存が増える。契約回帰に日次は必須ではない（generator `2026-08-30T12-49-01` と同型）。
3. Generator が `system` を選んだ主因は playback ブラウザ E2E との語衝突回避である。playback 側は `e2e` に固定すれば双方が一意になる。

## 3. Rejected

1. E2E を Integration gate / pre-push に相乗りさせる案 — Scope と失敗原因が混ざる。
2. Broad を重さ理由で pre-push から外す案 — 未実測の重さで収集境界を割り、pre-push 緑と GHA 緑の意味がずれる。
3. 毎 PR で E2E を必須にする案 — Access session 依存を merge 条件に固定する。
4. playback でも `system_e2e` を使う案 — Generator `system` との混線を残す。
