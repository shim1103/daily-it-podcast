# System e2e 引き継ぎ

## 現状
- 既定 System gate は run 33610705667 で緑（`-tags=system`）だった。真因は Gemini のレスポンス parse バグ（`output_audio.data` を読んでいたが実際は `steps[].content[].data`。Decision `2026-09-02T18-01-00`）。これまで「Gemini preview の Limitation」「無料枠 日次 quota」と記録していた System 失敗の主因はこれ。
- 続く run 33612303963 で「Gemini 以外 full」が `invalid_manuscript_draft: total rune count 2858`（Cursor 原稿が尺下限未満）で落ちた。flaky。対策として brief prompt に尺の逆算誘導と Self-check を入れた（Decision `2026-09-02T18-26-00`）。run 33627209650 で 3 連続 PASS（3825 / 3620 / 3473 文字、topic 5 件・下限 3360 クリア）。
- **再編（Decision `2026-09-03T14-45-00` / `14-46-00` / `14-47-00` / `16-30-00`。executor 実装済み・未 commit）**:
  - `generator-system.yml` の全 credential を `TEST_*` env へ。`test_run` パターンと inline Cursor smoke を撤去。
  - **cron 週次 + dispatch は system test を 1 回ずつ通すだけ**（`speech_synthesis` / `cursorcli_draft` 1 回版）。N 回ループも `-count=N` も入れない。壊れは 1 回で落ちる。
  - `gemini_excluded_full_system_test.go` / `produce_episode_system_test.go`(full) / `drive_observe.go` を削除。`//go:build system && full` 消滅。
  - PASS 率・所要の計測は dispatch 専用 test（`tts_rate` / `draft_rate`、`system && ratemeasure`）+ 専用 workflow に残す。**1 回通しが落ちた後の切り分け用**。env は `TEST_GEMINI_API_KEY` / `TEST_CURSOR_API_KEY` 直読み。
  - Cursor CLI install を `scripts/generator/install-cursor-cli.sh` へ切り出し。workflow の長い集計 shell を `scripts/generator/*-summary.sh` へ切り出し。

## 実装済（Decision 正）
1. Gemini audio 応答を `steps[].content[].data` から取る（`2026-09-02T18-01-00`、commit `eb72d8d`）。
2. TTS text を topic+2 束へ（`2026-09-02T13-55-00`）。`SpeechTexts` = `1 + topics + 1` 本。Domain 定数不変。
3. Gemini retry「同種 error 2 連続で打ち切り」+ 1 セグメント上限 `MaxAttempts` 3 / 既定 `callGap` 20s（`2026-09-02T13-56-00`）。`callGap` / `retryBackoffBase` / `retryBackoffMax` は `SpeechSynthesizer` の field 化し `NewSpeechSynthesizerWithTuning` で注入可（既定 constructor は既定値、挙動不変。`2026-09-03T14-46-00`）。**port を `SynthesizeAll(ctx, texts []string) ([]SpeechAudio, error)` へ変更**（`8ab057e`）。retry 予算は Adapter が束ね、1 度の `SynthesizeAll` 合計を `SynthesizeBudget` 15 回以内へ抑える二段構え（内側 `MaxAttempts` / 外側 `SynthesizeBudget`）。極小 PCM（`minPCMBytes` 未満）は `decode_pcm` 相当の retryable。`ProduceEpisode.Run` は for loop を 1 回の `SynthesizeAll` へ、`Timeline` / `ConcatWAV` は据え置き。gemini package は責務ごとに file 分割（`bd5bb9b`、挙動不変）。
4. audio 欠落 error に応答本文 snippet（400 byte）+ トップレベルキー一覧。
5. `speech_synthesis_system_test.go`（`system`）= 実 GEMINI で topic+2 束を 1 回 `SynthesizeAll` へ渡し通す。`cursorcli_draft_system_test.go`（`system`）= 固定擬似ソース → 実 Cursor `Write` → draft parse を 1 回（回帰確認。rate 計測は `draft_rate` へ分離）。
6. `build.ComposeBriefWithTemplate(items, template)` を分離（`ComposeBrief` は既定 template へ委譲、出力不変。`2026-09-03T14-47-00`）。
7. brief prompt に # Length strategy と # Self-check + parse 注意（`2026-09-02T18-26-00`）。Domain 定数不変。prompt `const` 本文の改善は `draft_rate` の variant（`testdata/brief_prompt_variant_*.txt`）で A/B。

## 次
1. `TEST_CURSOR_API_KEY` / `TEST_GEMINI_API_KEY` の repo Secret 登録（人手）。未登録だと TEST key 化後の実 API 部分が Skip / 失敗。
2. feature → develop PR。cron で走るのは `speech_synthesis` / `cursorcli_draft`(1 回版) の 1 回通しのみ。
3. **`generator-system` の e2e 前通し（`ProduceEpisode.Run` = topic 5 束を 1 回 `SynthesizeAll` → `Timeline` → `ConcatWAV` → Drive 書込）は port 変更後まだ実 dispatch していない**。`generator-tts-rate` は TTS 単体で確認済み（下記）。
4. cron が赤／不安定なとき rate 計測を dispatch: `gh workflow run generator-tts-rate.yml [-f runs= -f call_gap= -f double=min|tgt|max]` / `gh workflow run generator-draft-rate.yml [-f runs= -f prompt_variant=]`。結果は `generator-system-pass-rate.md` へ記録。
5. 3/3 の全体 3473 は下限 3360 まで 113 文字マージン。再発が続くなら `draft_rate` の variant で detail 目安を上げた prompt を検証してから `const` へ反映。

## rate 計測の実 dispatch 実績（このセッション）
- run 33780794360: `double` 導入前・先頭 1 束（greeting+intro、~21s）× 10 → PASS 10/10、平均 20.9s、429/retry 0。
- run 33828810131: 本番 topic 束 Max（~141s / 986 rune）× 10 → PASS 10/10、平均 71.8s、429/retry/MAX_TOKENS 0。長尺特有の失敗は 146s 圏内では観測されず。所要は 30〜104s と 3 倍以上ぶれる（synthesize 自体が非線形）。

## Drive の実到達
実 OAuth + 実 Drive list / write（create+upload）/ read / delete は旧「Gemini 以外 full」test で緑化（run 33610705667、Decision `2026-09-03T16-30-00` で当該 test は削除）。write 経路は Decision `2026-08-30T23-31-00` / `23-32-00` どおり（同 stem upsert・json→wav 公開順・補償なし）。以後の Drive 実到達確認は system test の 1 回通し（`speech_synthesis` は Drive を触らないため、通し経路での Drive 到達確認は develop 以降の別 test か本番 Run に委ねる）。

## memo
- `//go:build system` でビルドされる test file は `speech_synthesis_system_test.go` / `cursorcli_draft_system_test.go` の 2 本。`system && ratemeasure` で `tts_rate` / `draft_rate` が加わる。
- `tts_rate_system_test.go` の計測対象は `TTS_DOUBLE`（空/min/tgt/max、既定 max）で選ぶ本番 topic 束。未知値は `t.Fatalf`。`runs` 等の運用既定は `generator-tts-rate.yml` の `inputs.default` が SSOT、test 側の第 3 引数は env 未設定時のローカル最小フォールバック。
- gemini package の実装 file は責務ごとに `synthesizer.go` / `constructor.go` / `retry.go` / `backoff.go` / `transport.go`。各 file に対応 test file 1 本、共有 fake は `fake_test.go`。
- `interactionResponse.Status` フィールドは現状未使用。`status != "completed"` の扱いは必要になったら別 Decision。
