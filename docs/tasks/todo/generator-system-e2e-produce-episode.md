# System e2e 引き継ぎ

終わっていないことだけを書く。実装完了・decision 済みの経緯は `docs/decisions/` が SSoT、運用方針は `DEPLOY.md` §5 が SSoT。ここには写さない。

## 未完了

1. [ ] `TEST_CURSOR_API_KEY` / `TEST_GEMINI_API_KEY` / `TEST_GOOGLE_OAUTH_*` / `TEST_DRIVE_FOLDER_ID` を repo Secret / Variable へ登録（人手）。未登録だと `generator-system.yml` の e2e 1 回通しは Skip する。
2. [ ] feature → develop PR。
3. [ ] **e2e 1 回通し（`TestProduceEpisodeSystem`）が実 credential でまだ走っていない**。`composition.NewProduceEpisodeFromEnv` → `Run` を 1 度通す system test を新設済み（`//go:build system`、compile + env 無し Skip は手元確認済み）。`generator-system.yml` が `-tags=system` で拾う。実 3 情報源 → Cursor API 原稿 → Gemini TTS → OAuth+Drive 書込 の疎通と、通し経路での Drive 実到達がこの 1 本で確認できる。Fetch 窓に SourceItem 0 件だった日は `no_source_items` Domain Error で PASS 扱い。
4. [ ] **rate 計測 2 本が実 API でまだ走っていない**（dispatch 専用・`system && ratemeasure`、compile + key 無し Skip は手元確認済み）:
   - `TestGeminiTTSRate`（`generator-tts-rate.yml`）— 本番 topic 束を `runs` 回 `SynthesizeAll`。
   - `TestCursorAPIDraftRate`（`generator-draft-rate.yml`）— 固定擬似ソース → `Write` → draft parse を `runs` 回。環境要因の分母除外は `*cursorapi.Error` `Op=="do"`。prompt variant A/B は `testdata/brief_prompt_variant_a.txt`。
5. [ ] 全体尺の下限マージンが薄い（実績 3473 / 下限 3360 = 113 文字）。再発が続くなら `TestCursorAPIDraftRate` の variant で detail 目安を上げた prompt を検証してから `constants.TextWriterBriefPrompt` へ反映する。

## 未決（必要になったら別 Decision）

1. `interactionResponse.Status` は現状未使用。`status != "completed"` の扱い。
