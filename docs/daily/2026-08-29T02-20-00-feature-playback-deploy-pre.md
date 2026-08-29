---
name: Playback deploy 前 toolchain と運用 doc 整備
date: 2026-08-29T02:20:00
session_id: none
branch: feature/playback-deploy-pre
prev: なし
---

## 1. Summary

Playback の初回 deploy 前に、wrangler toolchain・`deploy:dry-run`・運用 doc の Phase 分離を repo 内で固定した。C-03/C-04（runtime config・Access）は shim 手動完了。Phase 1 ゲート AC と Phase 2 deploy は未。

## 2. Changes

1. `apps/playback` に wrangler devDependency、`build` / `types:worker` / `deploy:dry-run` script、Vite `outDir: web/dist`、`worker-configuration.d.ts`（ASSETS binding のみ）を追加した。
2. `web/dist` を `.gitignore` に追加し、build artifact を tracked 対象から外した。
3. `DEPLOY.md` を運用 SSOT のみへ KISS 化し、Phase 手順は `playback-lane.md` と新規 `playback-deploy-pre-gate.md` へ委譲した。
4. `playback-lane.md` で deploy 前 C を完了済みにし、Phase 1–4 index を再編した。
5. `README.md` / `DESIGN.md` に deploy 参照と `deploy:dry-run` を最小追記した。
6. `origin/develop` を merge し、list page design（PR #81）の差分を取り込んだ。
7. `npm run typecheck` と `npm run deploy:dry-run` が exit 0 であることを確認した。

### Commits

- `c9d92e5`
- `4e4c253`
