---
name: PlaybackはCD（master push契機の自動wrangler deploy）を採用する。DEPLOY.md §8の不採用方針をこの1点だけ転換する
date: 2026-09-15T05:04:17
branch: chore/cd-release-protect-smoke-e2e
---

## 1. Decision

1. `DEPLOY.md` §8「Playbackで採用しないもの」のうち、**CD・git hook自動deployのみ**を不採用listから外す。custom domain / Pages+別Worker / app内Access JWT / Service Token・WARP / preview URL共有 / DAST / Dependabot・Renovateは引き続き不採用のまま維持する。
2. `playback-deploy.yml`を新設する。`apps/playback/**`に変更があるmasterへのpush契機（+ `workflow_dispatch`）で`wrangler deploy`を自動実行する。
3. `playback-e2e.yml`は、`push: {branches: [master]}`から**`workflow_run`（`playback-deploy`の`completed`・`conclusion == 'success'`）**へ変更する。deploy成功を条件にし、deploy失敗時にe2eを走らせない。
4. `playback-e2e.yml`のjob内に、Cloudflare edgeへのdeploy反映を待つ固定30秒の待機stepを、`workflow_run`起点の実行時のみ挟む（`workflow_dispatch`起点の手動実行では待たない）。

## 2. Reason

1. shimの明示指示（本session）によりCD自動deploy不採用の既存方針を転換する。
2. `DEPLOY.md` §8が挙げていた根拠decisionを読み直すと、CD自動deploy不採用の直接的な理由は薄かった。`2026-09-04T02-04-02`は「ある運用後続タスクの完了条件にCD自動deployを含めない」という**限定scopeの決定**であり、Reason（「CD・DAST・Dependabotは軸が独立し、tool未選定のまま混ぜるとcheckboxが閉じない」）は未着手の整理であって恒久的禁止の理由ではない。`2026-08-25T17-10-00`はAccess/公開境界の話でCD自動deployには言及していない。したがって「不採用」という強い書き方は根拠decisionの実態より強く、今回のCD gate配線（`2026-09-15T03-51-41`）の前提（masterへのpush後に自動deployし、その後e2eを走らせる）と整合させるためこの1点を転換する。
3. DAST・Dependabot・custom domain等の他項目は、CD自動deployとは独立した軸であり、今回の転換理由（gate配線の整合）が及ばない。維持する。
4. shimの指摘「playback-e2eは相変わらずpush branch masterになっているけど、test-e2e.shの前にdeployはしているん？」の通り、単純な`push: {branches: [master]}`のままではdeploy自体が走る保証が無い（deploy workflowが存在しなかった）。`workflow_run`でdeploy成功を明示的な前提条件にすることで、「masterにpushされた」ことと「本番Workerに実際に反映された」ことを区別する。
5. shimの指摘「deploy後反映まで少し時間を待ってから実行させた方がいい」の通り、`wrangler deploy`のexit 0とCloudflare edgeへの反映完了は同時とは限らない。固定30秒の待機を挟み、deploy未反映によるE2Eのfalse negativeを避ける。`workflow_dispatch`（手動実行、既にdeploy済みの前提で叩く）では待機しない。

## 3. Rejected

1. **§8全体を撤回する案** — DAST・Dependabot・custom domain等はCD自動deployと無関係の軸であり、今回shimから転換の指示があったのはCD自動deployの1点のみ。他項目まで撤回する根拠が無い。
2. **`playback-e2e`を`workflow_dispatch`専用に戻し、手動deploy運用を維持する案** — 今回のCD gate配線の目的（deploy後に自動でe2eが走り、回帰を早期検知する）を満たせない。shimの指示（"update policy"）は自動化の方向を選んでいる。
3. **`playback-e2e`を`push: {branches: [master]}`のまま維持し、deploy workflowだけ並行で新設する案** — pushとdeploy成功のtimingが保証されず、deployが失敗または遅延した状態でE2Eが本番URLへ到達し、意味の無い赤（またはdeploy未反映のfalse negative）を出す。`workflow_run`で明示的にdeploy成功を前提条件にする。
4. **待機時間を挟まない案** — `wrangler deploy`のexit 0とedge反映完了が同時とは限らないというshimの指摘に対応できない。
