---
name: Gemini TTS は Developer API の PCM を WAV に wrap して保存・再生する
date: 2026-08-18T11:17:00
branch: feature/tts-speech-synthesizer
---

## 1. Decision

1. Gemini Developer API（Interactions TTS）の戻り raw PCM（24 kHz / 16-bit / mono）を、Adapter 内部で **WAV** bytes に wrap する。`SpeechSynthesizer.Synthesize` の成功戻りは非空の WAV bytes（`SpeechAudio.Content`）
2. Drive 配置契約の音声拡張子は `{episodeId}.wav`。一覧は `*.json` を stem 列挙（`contracts/drive-layout.md`）
3. mp3 encoder（`shine-mp3` 等）・外部圧縮 lib・Cloud Text-to-Speech API への乗り換えは行わない。Developer API + AgentSecrets proxy を維持
4. `docs/decisions/2026-08-17T17-41-59-feature-tts-speech-synthesizer.md` の Port 1 呼び出し・Adapter 定数・retry・課金方針は維持。戻り形式（mp3）だけ本 decision で上書き

## 2. Reason

1. Rule of Least Power / KISS（design-philosophy §4-2・§2-3）。Gemini Developer API は mp3 を返さず raw PCM のみ。再生可能な枯れた形式へは RIFF/WAV header 追加だけで足り、非可逆圧縮器は不要
2. Orthogonality / SRP。TTS HTTP と MPEG 圧縮を同一 Adapter に同居させない
3. Least Privilege。無名 encoder lib や CGO/LAME/ffmpeg を generator に載せない
4. UNIX 哲学（§4-1）。WAV は自己記述の標準 container。Playback は sample rate を契約に書かず `<audio>` で再生できる
5. YAGNI。配信・帯域圧縮（mp3）要件は今無い

## 3. Rejected

1. 現状どおり shine-mp3 で PCM→mp3（encoder 依存・品質/workaround コスト。philosophy に反する）
2. Cloud Text-to-Speech API で MP3 直出し（Developer API から製品・認証・endpoint が分岐。Least Power に反する）
3. raw PCM を Drive に保存（ブラウザ再生不可。Gemini の 24 kHz が Reader 契約へ漏れる）
4. Port 引数で encoding を選ぶ案（Application へ vendor/形式が逆流する）
