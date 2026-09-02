---
name: System suite は TTS 単体到達 test を主体にし、full produce run は full tag で分離する
date: 2026-09-02T13:57:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `apps/generator/test/system/` に **TTS 単体到達 test** を新規で置く。実 `GEMINI_API_KEY` で `gemini.SpeechSynthesizer` を組み、`build.SpeechTexts` が返す topic+2 本の短い朗読 text（疑似原稿から生成。Cursor / GetX は呼ばない）を順に `Synthesize` する。各戻りが非空 WAV で `build.WavDurationSec` が正の秒数を返し、**無料枠 quota（RPD / RPM）を超えず topic+2 回を完走できる**ことを ensure する。429 到達は「無料枠に収まらなかった」失敗として扱う。
2. 既存の full produce run test（`TestProduceEpisodeSystem_writesJsonAndWavPair_whenSubprocessSucceeds`）は build tag を `system` から **`system && full`** へ移す。既定の System 実行（`scripts/generator/test-system.sh` の `-tags=system`）では compile されず走らない。full は課金枠と潤沢な RPD がある時に手動で `-tags="system full"` 実行する。
3. Google Drive 書込経路の実到達確認は、この System suite ではなく Drive の Narrow Integration（実 OAuth / test Drive folder）で持つ。full run はそこに依存しない前提で分離する。
4. `workflow_dispatch` の実行是非・頻度は本 Decision の対象外（運用判断）。

## 2. Reason

1. これまでの dispatch は毎回 full `ProduceEpisode.Run` を回し、Cursor draft を焼いてから TTS で落ちていた（run 33581258235 / 33582558173）。失敗原因は 2 回とも Gemini TTS step に切り分け済みで、Cursor / GetX / Drive は各 Narrow で既に緑。TTS だけを叩く test なら 1 run の Cursor・GetX 消費がゼロになり、切り分け対象に必要な request だけを発行できる。
2. System / E2E test の目的は「下位 Scope で検証不可能なもの」に限定する（testing-strategy）。TTS が実 secret で `SpeechAudio` を返し、それが尺計算に乗ることは下位で確認できない。一方、全経路を毎回通すのは無料枠 RPD=15 では成立しない。
3. full run test を消さず `full` tag で残すのは、課金枠と quota が揃えば「入口から Drive 出口まで」の到達を後で確認したいため。既定 gate から外すだけにする。
4. Drive を Narrow で持てば、full run が TTS 起因で赤でも Drive 経路の回帰は独立に検出できる。

## 3. Rejected

1. full run test を削除する案 — 入口から出口までの唯一の到達 test を失う。課金枠移行後に必要になる。
2. full run test を `system` tag のまま残し dispatch で毎回回す案 — 無料枠 RPD=15 では 1 run すら安定せず、Cursor / GetX を毎回無駄に焼く。
3. TTS test を Narrow Integration（fake upstream）で済ませる案 — fake では「実 secret で実 Gemini が topic+2 回を quota 内で返す」ことを検証できない。それがこの Scope の存在理由。
4. TTS test で長文原稿を使い output size まで検証する案 — 検証目的は request 回数が無料枠に収まるかであって、audio の尺・品質ではない。短い固定 text で回数だけ回す。
