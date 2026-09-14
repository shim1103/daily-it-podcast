---
name: 完成音声の保存・配信を mp3 に一本化し、TTS〜結合は WAV のまま、encode は ffmpeg
date: 2026-09-13T13:40:29
branch: feature/playback-now-playing-audio-listenability
---

## 1. Decision

1. Drive 上の完成音声と playback HTTP 成功 body の形式は **mp3 のみ**とする。拡張子・Content-Type・`EncodeWAVToMP3` の正本は A（`contracts/drive-layout.md` / `apps/playback/contracts` / `build.EncodeWAVToMP3`）。
2. `SpeechSynthesizer`（TTS Port）の成功戻りは **セグメント WAV のまま**。Gemini が返す PCM の wrap・結合・尺計算（`ConcatWAV` / `WavDurationSec`）は Application 非公開 helper が WAV で行う。
3. WAV→mp3 は `EncodeWAVToMP3` を **ConcatWAV 成功直後・WriteEpisode 直前**に 1 回呼ぶ。再生 GET のたびに encode しない。
4. `EncodeWAVToMP3` の実装手段は **ffmpeg の subprocess**（OS / CI 上の `ffmpeg`）。Port にも TTS Adapter にも出さない。
5. 既存 `{episodeId}.wav` は **一括 batch migration** で `{episodeId}.mp3` へ変換・置換する。配信の wav/mp3 二系統は採らない。
6. 先行 Decision `2026-08-18T11-17-00` の「mp3 encoder を行わない」「配信圧縮は YAGNI」を、**保存・配信経路**について本 Decision が置き換える。TTS が PCM→WAV wrap する判断は維持する。

## 2. Reason

1. 本番音声は非圧縮 WAV で数十 MB 級になり、`<audio>` の初回 load が帯域で遅くなる。append-only・ほぼ update なしの配信物に非圧縮を残す前提（先行 YAGNI）が崩れた。
2. 結合・無音挿入・尺は PCM/WAV が可逆で単純。mp3 セグメント結合は再 encode と lossy の重ねになる。Port 出口〜結合までは WAV、配信用圧縮は書込直前の 1 点に閉じる。
3. ffmpeg は speech 向け bitrate とブラウザ互換 mp3 が取れ、GHA ubuntu に載せやすい。CGO+LAME はクロスビルドと CI を重くし、旧「C/無名 encoder を載せない」痛みと同型。pure Go encoder は品質・メンテの不確実性が残る。
4. 一括 migration は契約を一本化し、読取側の拡張子分岐を残さない。lazy 変換は初回 GET を遅くし、Worker に encode を載せる。
5. Opus は効率で勝つことがあるが、MIME・容器・互換の設計が増え、今の「mp3 へ移行」ゴールと契約変更コストが合う。

## 3. Rejected

1. TTS Adapter 内で mp3 化し Port 戻りを mp3 にする案 — 結合・尺が Adapter か再 decode に漏れ、先行 `2026-08-25T22-37-31`（結合を Adapter に閉じない）に反する。
2. 保存は WAV・配信だけ mp3（または両方保存）— 二系統の正本と運用コストが残る。最終ゴール（WAV 廃止）と衝突する。
3. 読取時・Worker 上で都度 encode — access 経路が重く、Range と相性が悪い。
4. CGO + LAME — C toolchain / cgo 依存。generator の単純な Go ビルドを壊す。
5. Cloud TTS を mp3 直出しに乗り換える案 — 認証・endpoint・製品が分岐し Least Power に反する（`2026-08-18T11-17-00` Rejected と同型）。
6. 過去資産を二系統 serve で残す案 — playback の MIME/拡張子分岐が永久に残る。
7. Opus を完成形式にする案 — 互換・MIME・容器の設計が増え、今の移行ゴール（mp3）より長い。効率だけなら将来再検討できる。
