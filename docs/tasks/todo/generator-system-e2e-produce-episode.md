# System e2e 引き継ぎ

終わっていないことだけを書く。実装完了・decision 済みの経緯は `docs/decisions/` が SSoT、運用方針は `DEPLOY.md` §5 が SSoT。ここには写さない。

## 完了

1. [x] `TEST_CURSOR_API_KEY` / `TEST_GEMINI_API_KEY` / `TEST_GOOGLE_OAUTH_*` / `TEST_DRIVE_FOLDER_ID` を repo Secret / Variable へ登録。
2. [x] **e2e 1 回通し（`TestProduceEpisodeSystem`）を実 credential で 1 度回した**。run [33857369881](https://github.com/shim1103/daily-it-podcast/actions/runs/33857369881)（`feature/playback-e2e-redeploy-master`、workflow_dispatch）で PASS（503.4s、`DRIVE_FOLDER_ID` 注入・Skip なし・episodeId `8ff4177b-26fe-4036-ab7b-d2a4e9e7639d` で Drive 実到達確認）。実 3 情報源 → Cursor API 原稿 → Gemini TTS → OAuth+Drive 書込 の疎通が緑。

## 未完了

1. [ ] **`TestGeminiTTSRate` が実 API でまだ走っていない**。`generator-draft-rate.yml` は実 API dispatch 済み（run 33840526373、3/3 PASS）。TTS 側も同様に 1 度 dispatch して尺帯ごとの PASS 率・所要を台帳化する。
2. [ ] draft 尺の下限マージンが薄い。run 33840526373 の default variant で 1 回が下限 +2 文字。`TestCursorAPIDraftRate` の variant `a` で detail 目安を上げた prompt を A/B し、良ければ `constants.TextWriterBriefPrompt` へ反映する。

## 未決（必要になったら別 Decision）

1. `interactionResponse.Status` は現状未使用。`status != "completed"` の扱い。
