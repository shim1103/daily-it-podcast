# System e2e 引き継ぎ

## 現状: 既定 System gate は緑（run 33610705667, `-tags=system`）
真因は Gemini のレスポンス parse バグ（`output_audio.data` を読んでいたが実際は `steps[].content[].data`。Decision `2026-09-02T18-01-00`）。これまで「Gemini preview の Limitation」「無料枠 日次 quota」と記録していた System 失敗の主因はこれ。修正後、TTS 単体 test と「Gemini 以外 full」test が両方緑（go test 本体 403s）。

## 実装済（Decision 正）
1. Gemini audio 応答を `steps[].content[].data` から取る（`2026-09-02T18-01-00`、commit `eb72d8d`）。
2. TTS text を topic+2 束へ（`2026-09-02T13-55-00`）。`SpeechTexts` = `1 + topics + 1` 本。`MarshalManuscript` は preface/detail 分離のまま。Domain 定数不変。
3. Gemini retry「同種 error 2 連続で打ち切り」+ `MaxAttempts` 3 / `callGap` 20s（`2026-09-02T13-56-00`）。retry は維持。
4. audio 欠落 error に応答本文 snippet（400 byte）+ トップレベルキー一覧（`41a751a` / `c776db6` / `92eb16e`）。
5. System suite 分割（`2026-09-02T13-57-00` / `16-57-00`）:
   - `speech_synthesis_system_test.go`（`system`）= 実 GEMINI で topic+2 回を無料枠 RPD 内で回す。
   - `gemini_excluded_full_system_test.go`（`system`）= GetX / Cursor / OAuth / Drive を実 secret で通し speech だけ fake。Drive 実書込→読取→削除まで到達。
   - `produce_episode_system_test.go`（`system && full`）= 入口→出口すべて実物。既定 gate 外。課金枠と潤沢 RPD 時に手動 `-tags="system full"`。
6. `test-system.sh` に `-v`（区間別 `t.Logf` を CI ログへ）。

## 次
1. feature → develop PR。
2. `-v` 付きで再 dispatch すれば区間別の所要秒（GetX fetch / Cursor draft / Drive write+read+delete）が取れる。
3. full run（`system && full`）は課金枠移行後に手動確認。

## Drive の実到達
実 OAuth + 実 Drive list / write（create+upload）/ read / delete は「Gemini 以外 full」test で緑化（run 33610705667）。write 経路は Decision `2026-08-30T23-31-00` / `23-32-00` どおり（同 stem upsert・json→wav 公開順・補償なし）。

## memo
- `drive_observe.go` は `//go:build system`。`system` のみビルドで一部 helper が未使用になるが無害。
- `interactionResponse.Status` フィールドは現状未使用。`status != "completed"` の扱いは必要になったら別 Decision。
