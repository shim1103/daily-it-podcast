---
name: Gemini TTS は Developer API から WAV を直接取得して保存・再生する
date: 2026-08-18T11:17:00
branch: feature/tts-speech-synthesizer
---

## 1. Decision

1. Gemini Developer API（Interactions TTS `gemini-3.8-flash-tts`）は unary 応答で標準 **WAV**（`audio/wav`）を直接返す。Adapter 内部での raw PCM から WAV への wrap 処理を廃止し、取得した WAV bytes（`SpeechAudio.Content`）を直接返す
2. Drive 配置契約の音声拡張子は `{episodeId}.wav`。一覧は `*.json` を stem 列挙（`contracts/drive-layout.md`）
3. mp3 encoder（`shine-mp3` 等）・外部圧縮 lib・Cloud Text-to-Speech API への乗り換えは行わない。Developer API + AgentSecrets proxy を維持
4. `docs/decisions/2026-08-17T17-41-59-feature-tts-speech-synthesizer.md` の Port 1 呼び出し・Adapter 定数・retry・課金方針は維持。戻り形式（mp3）だけ本 decision で上書き

## 2. Reason

1. Rule of Least Power / KISS（design-philosophy §4-2・§2-3）。Gemini 3.8 Flash TTS は最初から標準 WAV を返す。client 側での header 組立や PCM サンプルレート等のフォーマット管理は一切不要となり、受け取った WAV を検証してそのまま流すだけで足りる
2. Orthogonality / SRP。TTS HTTP と音声形式変換を同一 Adapter に同居させない
3. Least Privilege。無名 encoder lib や CGO/LAME/ffmpeg を generator に載せない
4. UNIX 哲学（§4-1）。WAV は自己記述の標準 container。Playback は sample rate を契約に書かず `<audio>` で再生できる
5. YAGNI。配信・帯域圧縮（mp3）要件は今無い

## 3. Rejected

1. 旧答え: Adapter 内部で raw PCM を WAV bytes へ wrap する案（Gemini 3.8 Flash が直接標準 WAV を返すため、client 側の PCM 変換コードおよび PCM 前提定数が不要になり、過剰な責務となったため廃止）
2. 現状どおり shine-mp3 で PCM→mp3（encoder 依存・品質/workaround コスト。philosophy に反する）
3. Cloud Text-to-Speech API で MP3 直出し（Developer API から製品・認証・endpoint が分岐。Least Power に反する）
4. raw PCM を Drive に保存（ブラウザ再生不可。Gemini の 24 kHz が Reader 契約へ漏れる）
5. Port 引数で encoding を選ぶ案（Application へ vendor/形式が逆流する）
