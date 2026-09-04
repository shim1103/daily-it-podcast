# System e2e 引き継ぎ

終わっていないことだけを書く。実装完了・decision 済みの経緯は `docs/decisions/` が SSoT、運用方針は `DEPLOY.md` §5 が SSoT。ここには写さない。

## 未完了

1. [ ] `TEST_CURSOR_API_KEY` / `TEST_GEMINI_API_KEY` / `TEST_GOOGLE_OAUTH_*` / `TEST_DRIVE_FOLDER_ID` を repo Secret / Variable へ登録（人手）。未登録だと `generator-system.yml` の e2e 1 回通しは Skip する。
2. [ ] **e2e 1 回通し（`TestProduceEpisodeSystem`）が実 credential でまだ走っていない**。`generator-system.yml` を実 `TEST_*` で 1 度回し、実 3 情報源 → Cursor API 原稿 → Gemini TTS → OAuth+Drive 書込 の疎通と Drive 実到達を確認する。Fetch 窓に SourceItem 0 件だった日は `no_source_items` で PASS 扱い。
3. [ ] **`TestGeminiTTSRate` が実 API でまだ走っていない**。`generator-draft-rate.yml` は実 API dispatch 済み（run 33840526373、3/3 PASS）。TTS 側も同様に 1 度 dispatch して尺帯ごとの PASS 率・所要を台帳化する。
4. [ ] draft 尺の下限マージンが薄い。run 33840526373 の default variant で 1 回が下限 +2 文字。`TestCursorAPIDraftRate` の variant `a` で detail 目安を上げた prompt を A/B し、良ければ `constants.TextWriterBriefPrompt` へ反映する。

## 未決（必要になったら別 Decision）

1. `interactionResponse.Status` は現状未使用。`status != "completed"` の扱い。
