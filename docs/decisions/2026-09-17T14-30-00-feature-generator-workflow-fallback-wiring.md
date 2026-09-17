---
name: master push 契機の CD 連鎖を generator-system → playback-deploy → playback-e2e に変える
date: 2026-09-17T14:30:00
branch: feature/generator-workflow-fallback-wiring
---

## 1. Decision

1. `generator-system.yml` の起動条件を、`pull_request: branches:[master]` から `push: branches:[master]` へ変える。master 向け PR 時の起動は廃止する。
2. `playback-deploy.yml` の起動条件を、`push: branches:[master], paths:["apps/playback/**"]` から `workflow_run: workflows:[generator-system], branches:[master], types:[completed]` へ変える。`generator-system.yml` が成功した時だけ deploy する（`if: github.event.workflow_run.conclusion == 'success'`）。`paths` によるファイル差分絞り込みは廃止し、generator 側の変更を含む master push でも deploy 対象にする。
3. `playback-e2e.yml` の起動条件（`workflow_run: workflows:[playback-deploy]`）は変更しない。CD 連鎖全体は `master push → generator-system → playback-deploy → playback-e2e` の順になる。

## 2. Reason

1. `generator-system.yml` は credential 付き・実 API 疎通を伴う gate 外 workflow であり、master 向け PR のたびに実行すると運用負荷が高い（DEPLOY.md に既存の一時無効化記録がある）。PR gate ではなく master push（= リリース）契機に絞ることで、実行頻度をリリース単位に落とす。
2. `playback-deploy.yml` が `apps/playback/**` の path 差分だけを見ていると、generator 側の変更（TextWriter/TTS fallback 実装等）が master へ merge されても playback は deploy されない。CD 連鎖の起点を `generator-system` の成功に一本化することで、「generator が壊れていないことを確認してから playback を出す」という順序を、path 差分の有無に関わらず一貫させる。
3. `playback-deploy` を `generator-system` の成功にだけ依存させる理由：`generator-system` が失敗する master push（regression 混入）で playback を自動 deploy すると、壊れた generator と結びついた playback が本番に出てしまう。deploy 前に system 疎通を確認するゲートとして機能させる。

## 3. Rejected

1. **`generator-system.yml` の `pull_request` trigger を残したまま `push` も追加する案（PR 時・push 時の両方で起動）** — 実行頻度が減らず、運用負荷軽減という目的を果たせない。
2. **`playback-deploy.yml` の `paths` 絞り込みを維持したまま `workflow_run` を追加する案（両方の条件を OR で満たす）** — `generator-system` 成功だけを唯一のゲートにするという意図が薄まり、path 差分次第で連鎖が発火しないケースが残る。
3. **`playback-deploy.yml` を `generator-system` の結果を待たず push 直後に即時実行する案（現行）** — generator 側の regression が playback deploy をブロックしない。deploy 前の疎通確認ゲートとして機能しない。
