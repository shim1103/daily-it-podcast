# System e2e 引き継ぎ（途中）

## blocker
Gemini TTS 無料枠 = `gemini-2.5-flash-preview-tts` 相当で **3 RPM / 10,000 TPM / RPD=15**。1 run の `Synthesize` 回数は `build.SpeechTexts` の本数で決まり、retry と 429 はいずれも RPD を消費する。無料枠では 1 run すら安定しない（run 33581258235 = decode_pcm 尽き、33582558173 = 429）。有料枠移行が前提。

## 実装済（Decision 正）
1. TTS text を topic+2 束へ（`2026-09-02T13-55-00`）。`SpeechTexts` = `1 + topics + 1` 本。`Timeline` / Run コメント追随。`MarshalManuscript` は preface/detail 分離のまま。Domain 定数不変。
2. Gemini retry「同種 error 2 連続で打ち切り」+ `MaxAttempts` 3 / `callGap` 20s（`2026-09-02T13-56-00`）。retry は維持。
3. `decode_pcm` の audio 欠落 error に応答本文 snippet（commit `41a751a`）。
4. System suite を TTS 単体到達 test 主体へ（`2026-09-02T13-57-00`）。`test/system/speech_synthesis_system_test.go` は `GEMINI_API_KEY` のみ要求。full run（`produce_episode_system_test.go`）は `//go:build system && full` で既定 gate 外。
5. 「Gemini 以外 full」test（`2026-09-02T16-57-00`）。`test/system/gemini_excluded_full_system_test.go`（`//go:build system`）。GetX / Cursor / OAuth / Drive を実 secret で通し、speech だけ無音 WAV fake。Drive 実書込→実読取→実削除まで到達。各区間の経過秒を Logf。`GEMINI_API_KEY` は要らない。

## 次
1. `generator-system.yml` を `workflow_dispatch`（既定 `-tags=system`）。今の gate で走るのは TTS 単体 test と「Gemini 以外 full」test。
   1. 「Gemini 以外 full」= GetX / Cursor / OAuth / Drive の実到達と所要時間の実測（Gemini quota を消費しない）。
   2. TTS 単体 = 有料枠へ切替後に「topic+2 回が無料枠 RPD 内で回るか」を確認。
2. full run（入口→出口すべて実物）を確認するなら `-tags="system full"` を手動実行。課金枠と潤沢な RPD が要る。
3. 緑なら feature → develop PR。

## Drive の実到達
実 OAuth（refresh→access）+ 実 Drive list は過去 run で到達済み。実 Drive **write**（create + upload media）は「Gemini 以外 full」test で初到達を確認する（それまで未検証）。write 経路は Decision `2026-08-30T23-31-00` / `23-32-00` どおり実装済み（同 stem upsert・json→wav 公開順・補償なし）。

## memo
- `test/system/drive_observe.go` は `//go:build system`（full なし）のまま。`system` のみビルドで helper が未使用になるが無害。将来 `system && full` へ寄せると最小。
