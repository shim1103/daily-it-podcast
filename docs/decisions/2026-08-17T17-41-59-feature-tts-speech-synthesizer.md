---
name: 読み上げは Port 1 呼び出し。本文以外は Adapter 定数
date: 2026-08-17T17:41:59
branch: feature/tts-speech-synthesizer
---

## 1. Decision

1. 読み上げの単位は Application 所有の Port `SpeechSynthesizer.Synthesize`。入力は朗読本文だけ。戻りは WAV の中身（`SpeechAudio`）。形式の正は `2026-08-18T11-17-00`。呼び出し入口（CLI / UseCase pipeline）は決めない
2. 本文以外（model / voice / envelope / endpoint）は Gemini Adapter 定数へ閉じる。今は空文字。仕様決定後に同じ定数へ埋める
3. retry は Adapter 内部。無限回を防ぐため上限定数 `MaxAttempts` を置く。回数の正は constants.go
4. 課金は安い plan。学習利用は許容する。無料枠の実行時間制約は agent が見ない（archive `docs/human/MEMO.md` §4 を継承）。Free / Paid の切替は実行設定であり code に持たない

## 2. Reason

1. DIP。Application は Gemini の HTTP・PCM・voice 名を知らない
2. KISS / YAGNI。今の caller は声や口調を選ばない。空定数は `WatchUserIDs` dummy と同じく置き場の固定
3. 公式 TTS の稀な 500 と 429 / 503 はどちらも有限 retry。上限が無いと打ち切れない
4. archive は金額を文書へ固定しない。今の公式は Free が最安かつ学習利用あり。収まらないかは rate limit の実行時判断

## 3. Rejected

1. Port 引数に voice / language / speed を置く案（vendor 値が Application へ逆流する）
2. 戻りを filesystem path にする案
3. 素通し UseCase を今作る案（手順が無い）
4. model 名や価格表を README / DESIGN へ写す案（DRY。契約は Port、値は constants、秘密名は README）
5. Free / Paid を Adapter 定数や Port で切る案
