---
name: masterへのCD gate配線とplaybackのsmoke/e2e責務分担を再設計し、workflow ymlを先行反映した
date: 2026-09-15T08:56:33
session_id: cd-release-protect-smoke-e2e-a0
branch: chore/playback-smoke-e2e-redesign
prev: なし
---

## 1. Summary

masterにbranch protectionが無く、e2e/system-testもPR/pushに未連動だった問題を起点に、CD gate全体の配線とplaybackのdeploy前検証（smoke）・deploy後検証（e2e）の責務分担を設計し直した。smokeの設計は複数回の誤りと訂正を経て、最終的に「TEST専用credentialでVite dev server上のHono appへ実composition経路を通す、UI操作込みの1本のPlaywright test」に収束した。workflow yml関連の2 commitをmasterへ先行PR（#164）し、shim承認によりmerge済み。実装branchはdevelop最新へrebaseし、`gh workflow run playback-smoke.yml --ref <branch>`でCI上のsmoke実行が3 test全てpassすることを確認した。

## 2. Changes

- `master`のbranch protection新設（`static-and-unit`・`integration`をrequired、review不要、admin対象外）。
- `generator-produce-episode.yml`にffmpeg installを追加（run 34866427635の失敗原因を修正）。
- `generator-system.yml`・`playback-e2e.yml`をweekly cronからPR/deploy後発火へ切り替え。`playback-deploy.yml`（masterへのpush契機の自動`wrangler deploy`）・`playback-smoke.yml`（masterへのPR契機）を新設。`playback-e2e.yml`は`playback-deploy`成功後に`workflow_run`で発火し、edge反映待ちの30秒sleepを挟む。
- playback smoke-testの設計は3段階で訂正した。(1) credential無しのdummy backend fixtureでUI操作のみ確認 → shimの指摘で「Access以外の本番相当経路が生きているか」という核心を満たさないと判明し撤回。(2) TEST専用credentialでのNode直接API疎通のみ（UIを経由しない） → 実browser（Playwright）でしか検証できない領域（実Integrationはhappy-domでaudio要素もFake化）を落としていたと判明し撤回。(3) 最終形：`web/vite.smoke.config.ts`が`createApp()`（override無し、本番と同じcomposition経路）をTEST専用credential入り`env`で呼ぶVite dev serverを提供し、そこへPlaywrightで実browserアクセスして一覧・原稿・再生を1本のtestで確認する。
- `web/dev-backend-proxy.ts`を新設し、既存`web/vite.config.ts`（fixture固定のdev用dummy backend）と新設`vite.smoke.config.ts`が共有する透過中継middlewareへ統合。当初Range/HEADを独自実装していたが、本番Hono app（`routes/audio-response.ts`）が既にRange対応済みと判明し、Request headerの転送漏れが原因と特定。透過中継のみへ簡素化した。
- `authenticated_playback.e2e.spec.ts`は4 test（一覧・原稿・再生・seek）から1 test（Access session経由での本番Drive到達確認）へ削減。UI機能はsmoke側が担うため。
- `apps/playback/tsconfig.json`の`types`へ`node`を追加し、Vite config file群が使う`process`のTypeScript型エラーを解消。
- README/DESIGN/DEPLOYへscope/non-scope宣言を追加し、README側にあったDESIGN.mdとの重複記述（Technology choices表）を削除してDRY化。
- Decision 3件をCREATE：`2026-09-15T03-51-41`（CD gate配線）、`2026-09-15T05-04-17`（CD自動deploy採用、既存`DEPLOY.md`§8の「CD不採用」方針をこの1点だけ転換）、`2026-09-15T05-36-32`（smoke最終設計）。session内で確定前に作った中間decision（4件）はcommit・公開前に削除・書き直した（ADR immutable原則はcommit済みのみに適用）。
- workflow ymlのみを`origin/master`起点の新branch（`chore/cd-gate-workflows`）へcherry-pickしPR #164として先行反映。`gh workflow run <yml> --ref <branch>`はmasterにyml file名さえあれば、別branch側の実装を使って検証できることを確認した。
- 実装branchを`chore/cd-release-protect-smoke-e2e`から`chore/playback-smoke-e2e-redesign`へrenameし、develop最新（649be69、PR #161まで反映）へrebase。既存unit/integration（59 files/419 tests）全pass、typecheck通過を確認。
- `gh workflow run playback-smoke.yml --ref chore/playback-smoke-e2e-redesign`を実行し、CI上でTEST_*credential・実Google OAuth/Drive疎通・UI操作込みの3 test全passを確認した（手元では確認できなかったGreen経路を初めて確定）。

### Commits

- `135f797`（ffmpeg fix、rebase後SHA）
- `c7d9057`（CD gate配線、rebase後SHA）
- `0d5bf56`（smoke再設計・docs整理・playback実装本体、rebase後SHA）

## 3. 次に必要な作業

- developへの本PR（実装＋docs、`chore/playback-smoke-e2e-redesign` → `develop`）は未作成。
- non-scope扱いで未実行：`generator-system`（system-test）・`playback-e2e`（deploy後e2e）・`playback-deploy`（本番deploy）のCI実行確認。
