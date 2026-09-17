---
name: TextWriterのmodel切り替えfallback実装とsystem-test topic数動的化
date: 2026-09-17T10:45:26
session_id: none
branch: feature/generator-textwriter-fallback
prev: なし
---

## 1. Summary

TextWriterのGemini呼び出しに、TTS（PR #177）と同型のmodel切り替えfallback機構を実装した。`port.TextWriter.Write`のcontractを、生文字列を返す形から呼び出し側がDIするbuildFn（raw→ManuscriptDraft変換）を受け取り検証済みdraftを返す形へ変更し、invalid-draft時のretryをTextWriter実装側（cursorapi/geminiapi）へ移管した。これにより`GEMINI_API_KEY`(free)→`CURSOR_API_KEY`→`SPARE_GEMINI_API_KEY`(paid, final)の3段fallbackをcomposition層で結線できるようにした。

あわせて、system-testが`SYSTEM_TEST_TOPIC_COUNT`環境変数でtopic数を任意指定できるようにし、topic数依存の秒数・文字数畳み込み計算式を任意topic数を受け取る関数へ一般化した。source取得件数（各ItemSource Adapterのmaxitems）もtopic数に連動して絞れるようにした。

このPRはfeature/generator-textwriter-adapter-fallback branchで進めていた一連の作業を、develop向けstacked PR（#177 TTS fallbackをbaseにした3番目のPR）へ分割する過程で切り出したもの。作業中、developのGoogle Drive撤去（R2完全移行）に追従していなかった複数ファイル（narrow integration test、system test、fixture）の乖離を発見し、あわせて修正した。

## 2. Changes

1. `port.TextWriter.Write`のsignatureをbuildFn DI形式へ変更し、`application/manuscript.TextWriter`（sources配列を順に試すfallback合成layer）を新設
2. `infrastructure/manuscript/cursorapi`・`geminiapi`にinvalid-draft retryループを実装（TextWriterMaxAttempts相当をAdapter側へ移管）
3. `application/build`のtopic数検証・prompt埋め込み関数（`ManuscriptDraftFromWriterOutput`等）をtopicCount引数化
4. `entities/constants`の秒数・文字数畳み込み定数を任意topic数向けの関数（`TotalTgtSecFor`等）へ一般化。`DraftTopicCountMin/Max`廃止漏れのcontract test参照を修正
5. `writer_output_example.json`のfixtureとprompt文言（`TOPIC_COUNT_MIN/MAX`）を、topic数固定値化後の現行文字数range・placeholder集合へ追従
6. 各ItemSource Adapter（HackerNews/Lobsters/Publickey/TechCrunch/クラウドWatch）にmaxitems引数を追加（`<=0`で既存既定上限へフォールバック、本番挙動は不変）
7. `composition/produce_episode.go`をTextWriter 3要素fallback配列・topicCount引数化・ItemSource maxItems配線へ更新。`NewProduceEpisodeFromEnvWithTopicCount`を新設し本番経路はこれへ委譲
8. developのGoogle Drive撤去に追従していなかった`test/integration_support_test.go`・`test/system/produce_episode_system_test.go`・`internal/composition/produce_episode_topic_count_sociable_unit_test.go`をR2ベースの現行credential契約へ揃えた
9. system-test topic数動的化のdecisionを追加
10. PR作成予定（base: feature/generator-tts-fallback）

### Commits

- `38535b9`
- `505a997`
- `a3c5b13`
- `942beb0`
- `16a37e9`
- `28574ac`
- `4e0de28`
