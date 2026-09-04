---
name: playback 再 deploy・generator system 実測・opening/ending 朗読全文契約・SpeechTexts 境界改行
date: 2026-09-04T17:40:00
session_id: none
branch: feature/playback-e2e-redeploy-master
prev: なし
---

## 1. Summary

playback Worker を再 deploy し、`playback-e2e` を `develop` 上で緑確認した。ローカル `~/projects/daily-it-podcast` の master は rebase-merge 途中で `git pull` 不能になっていたため、rebase 継続・conflict 解消まで行い、local と `origin/master` の divergence 解消は shim 手元の判断に残した。generator system e2e（`TestProduceEpisodeSystem`）は本体 PASS（約 587s）だが、summary script が fail=0 でも exit 1 になり workflow 全体が赤になった。Drive の `TEST_DRIVE_FOLDER_ID` へ episode が残り、shim が本番へ move（`episodeId` `4d22034f-66a5-42bc-81e1-775e5e82af2f` / `date` `2026-09-04`）。その後 Actions を Node24 major へ上げ、system-summary を fail=0 で exit 0 に直し、manuscript の `body.opening` / `body.ending` を定型込みの朗読全文（Drive SSoT）へ直し、e2e fixture を当該 episode へ差し替えた。Gemini TTS の息継ぎ・同一 voice は non-edit で整理し、Cursor brief の `topic.detail` 改行規約（最大1個・validation なし）と SpeechTexts 束境界の改行3個を実装した。

## 2. Changes

1. playback: `npm run build` + `npx wrangler deploy` → `https://daily-it-podcast.shim1103thy.workers.dev`（Version `c02a4915-3b67-448b-b73d-ebbf60c44e5d`）。`playback-e2e.yml` dispatch on `develop` → run [33838411352](https://github.com/shim1103/daily-it-podcast/actions/runs/33838411352) success（Playwright 3 passed）。
2. generator system: `generator-system.yml` dispatch on `feature/generator-system-e2e-produce-episode` → run [33844002378](https://github.com/shim1103/daily-it-podcast/actions/runs/33844002378)。test 本体 PASS、workflow は `scripts/generator/system-summary.sh` の pipefail 下 exit 1 で overall failure。Drive 成果物は invariant どおり削除しない。
3. shim 指摘: Drive 原稿は TTS が読む全文の SSoT。`body.opening` = greeting+intro、`body.ending` = closingSummary+farewell。schema 文言も「Generator が TTS へ渡す」ではなく「TTS が読み上げる原稿そのもの」へ。
4. Gemini TTS AskQuestion（non-edit）: 自然な息は prompt、確実な間は segment 間 PCM 無音、同一声は `voice` 固定。create-agent 再接続は Interactions の `previous_interaction_id` とは別物。
5. Decision `2026-09-04T16-00-00`（opening/ending 朗読全文）と `2026-09-04T17-05-00`（束境界改行3 / detail 改行最大1）を記録済み。完了した `ci-actions-node24-bump` todo を削除。
6. 未実施のまま: 当該 episode の `body.opening` が greeting のみの可能性がある再 produce、本 branch の PR 作成、shim 手元 master の `origin/master` 揃え。

### Commits

- `093f3d9`
- `ed26e30`
- `a673945`
- `f0363e3`
- `e2b493e`
- `f8ac2b0`
- `5f45f86`
