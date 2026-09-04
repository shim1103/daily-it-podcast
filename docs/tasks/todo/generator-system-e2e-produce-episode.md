# System e2e 引き継ぎ

終わっていないことだけを書く。実装完了・decision 済みの経緯は `docs/decisions/` が SSoT、運用方針は `DEPLOY.md` §5 が SSoT。ここには写さない。

## 未完了

1. [ ] `TEST_CURSOR_API_KEY` / `TEST_GEMINI_API_KEY` を repo Secret へ登録（人手）。未登録だと `generator-system.yml` の実 API 部分が Skip / 失敗する。
2. [ ] feature → develop PR。
3. [ ] **`generator-system` の e2e 前通しが未 dispatch**。`ProduceEpisode.Run`（topic 束を 1 回 `SynthesizeAll` → `Timeline` → `ConcatWAV` → Drive 書込）を実 credential で通した実績が port 変更後まだない。`generator-tts-rate` は TTS 単体でしか確認できていない。
4. [ ] **通し経路での Drive 実到達確認が未**。`speech_synthesis_system_test.go` は Drive を触らない。旧 full test は削除済みなので、通し経路の Drive 到達は §3 の e2e 前通しか本番 Run に委ねている。
5. [ ] **cursorapi 版の system 網羅が未着手**。Cursor CLI → Cloud Agents HTTP API 移行（Decision `2026-09-03T17-03-33`）で旧 `cursorcli_draft_system_test.go` / `draft_rate_system_test.go` / `generator-draft-rate.yml` を削除した。HTTP API 版の system 回帰 test（1 回通し）と draft rate 計測（dispatch 専用 workflow）は別途起票する。台帳は `generator-lane.md` 未完了 4。
6. [ ] 全体尺の下限マージンが薄い（実績 3473 / 下限 3360 = 113 文字）。再発が続くなら prompt `const`（`constants.TextWriterBriefPrompt`）の detail 目安を上げる。

## 未決（必要になったら別 Decision）

1. `interactionResponse.Status` は現状未使用。`status != "completed"` の扱い。
