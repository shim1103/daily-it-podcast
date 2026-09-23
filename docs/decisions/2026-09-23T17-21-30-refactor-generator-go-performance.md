---
name: WAV segment の尺計算は port.SpeechSynthesizer の契約に含め、ConcatWAV 側の再計算を廃止する
date: 2026-09-23T17:21:30
branch: refactor/generator-go-performance
---

## 1. Decision

1. WAV segment の尺（duration）計算は `port.SpeechSynthesizer.SynthesizeAll` の戻り値契約に含める（`port.SpeechSynthesizer` interface 自体を変更する）。infrastructure 側の実装（Gemini 等）が音声 byte 列と併せて尺情報を返す。
2. `ProduceEpisode.Run` は `SynthesizeAll` から尺を直接受け取り、独自に `build.WavDurationSec` を呼ぶ処理（現行 L114-123）を廃止する。
3. `ConcatWAV`（`build.ConcatWAV`）は「WAV 結合」のみの責務を維持する。尺計算の重複呼び出し排除を目的に `ConcatWAV` 自体の入出力契約は変更しない。
4. `build.Timeline` の責務・存在は変更しない。segment 毎の秒数列を受け取り、topic 開始秒・ending 開始秒・全体 duration という podcast 構造上の値へ変換する役割を維持する。

## 2. Reason

### 2-1. 前提事実（判断時に確認済みの現状）

1. `ProduceEpisode.Run`（`internal/application/produce_episode.go` L114-123）は、`uc.speech.SynthesizeAll` が返す `audios`（`[]models.SpeechAudio`、各要素が WAV byte 列）を受け取った直後、各要素に対し独自に `build.WavDurationSec` を呼び、`segmentDurations` を算出していた。この loop は既存の `uc.progress.Start`/`Done` で囲まれておらず、`synthesize_speech` の Done と `build_timeline` の Start の間に裸で挟まっていた。
2. `build.WavDurationSec`（`internal/application/build/wav_duration.go`）は内部で `parseWAV`（RIFF/WAVE header 解析）を呼ぶ。`build.ConcatWAV`（`internal/application/build/wav_concat.go`）も各 part に対し同じ `parseWAV` を呼んでいる。つまり `Run` が呼ぶ `WavDurationSec` と、後続の `ConcatWAV` 内部の parse は、同一 WAV byte 列に対する重複した parse だった。
3. `port.SpeechSynthesizer.SynthesizeAll`（`internal/application/port/speech_synthesizer.go`）は既存契約として `[]models.SpeechAudio`（WAV byte 列を持つ）を返す。`models.SpeechAudio`（`internal/entities/models/speech_audio.go`）は変更前 `Content []byte` のみを持ち、尺情報は持っていなかった。
4. `build.ConcatWAV` の既存 doc comment は「同一 PCM パラメータの WAV を 1 本に結合する」とのみ書かれており、topic・ending 等の podcast 構造を扱う責務は持っていない。
5. `build.Timeline`（呼び出し箇所: `Run` L126）は `segmentDurations` と topic 数を受け取り、topic 開始秒・ending 開始秒・全体 duration を算出する、`WavDurationSec`/`ConcatWAV` とは別の既存関数。

### 2-2. 選択の理由

1. **`port.SpeechSynthesizer` 契約自体を変更する理由（4a 採用）**：2-1-3 の通り、WAV を返すことは `SpeechSynthesizer` の既存の共通契約であり、尺情報も同じ箇所（infrastructure の実装が WAV を組み立てる末端）で一緒に返す方が自然。`application/speech` 層に尺計算専用の wrapper 関数を別途新設する案（4b）は、`speech` package 内に「fallback chain 本体」と「尺計算 wrapper」という 2 つの入口を作ってしまい、責務が分散する。
2. **`ProduceEpisode.Run` から尺計算 loop を除去する理由**：2-1-1・2-1-2 の通り、`Run`（application 層の use case）が `audios` という音声実装由来の生 byte 列を直接 parse しており、「組み立ての知識を持つ」という `Run` の責務からも外れた重複処理になっていた。`ConcatWAV` 内部でも同じ `parseWAV` 相当の処理が実行されており、二重計算になっている。
3. **`ConcatWAV` 自体は変更しない理由**：2-1-4 の通り、`ConcatWAV` の責務は「同一 PCM パラメータの WAV を 1 本に結合する」ことのみであり、doc comment 上も topic・ending の区別を知るべきでない。尺計算は `ConcatWAV` の目的（結合）にとって本質的な処理ではなく、責務混在を避けるため `ConcatWAV` 自体は変更対象から外す。重複計算の解消は、呼び出し元（`SynthesizeAll`）が先に尺を確定させ、以降の呼び出し（`Timeline`）へ渡す形で行う。
4. **`build.Timeline` を維持する理由**：2-1-5 の通り、`WavDurationSec`（個々の segment が何秒かを求める処理）と `Timeline`（segment 毎の秒数列を、podcast 構造上意味のある値へ変換する処理）は別の抽象度の処理であり、尺計算の呼び出し元を変えることは `Timeline` の存在理由に影響しない。

## 3. Rejected

1. **`application/speech` 層に尺計算専用の wrapper 関数を新設する案（4b）** — `port.SpeechSynthesizer` の interface は変更しないため infrastructure 層（Gemini 等）への変更波及は避けられるが、`speech` package 内に入口が 2 つできる。WAV を返す契約と尺を返す契約を分離する理由がないため、4a（契約自体への統合）を採用した。
2. **`ProduceEpisode.Run` 内に尺計算専用の中間 step（`ParseSegments` 等）を新設する案** — `Run` に新しい謎 step が増えるだけで、「音声 file 列 → 結合」というユーザ視点の流れを分かりにくくする。`SynthesizeAll` の戻り値に統合する方が素直なため見送った。
3. **`ConcatWAV` が尺情報を副産物として返す案** — `ConcatWAV` は topic・ending の区別を知らない責務のままであるべきで、尺計算という別の関心をここに持たせると責務が混在する。
