---
name: segment 間無音は concatWAV helper と定数秒数で挿入する
date: 2026-08-29T14:15:00
branch: docs/produce-episode-run-spec
---

## 1. Decision

1. 複数 segment の WAV を episode 1 本にする際、**隣接 segment の間**に `entities/constants` で固定した **無音秒数**を挿入する。実装は **`concatWAV`（Application 非公開 helper）** が所有する。
2. 無音長の正本は定数 `SegmentSilenceSec` とする。TTS HTTP 入力へ pause メタ指示を書かない。
3. `durationSec`（原稿 JSON）と `startSec` 算定は、**無音 insert 分を含めた**累積尺を正とする。

## 2. Reason

1. TTS Port は朗読 text のみ。pause を transcript に書くと Gemini envelope 外の演出文として読まれるリスクがある（既存 Adapter は Transcript ラベル以降を読む）。
2. segment 分割（別 Decision）と組み合わせると、API 呼び出し境界だけでは物理的な 1 秒 gap を保証できない。PCM 無音挿入が mechanism として最小（KISS）。
3. RIFF 操作は Application 非公開 helper に置く方針（`2026-08-25T22-37-31`）と一致する。

## 3. Rejected

1. `SpeechSynthesizer` Port に pause 引数を足す案 — vendor 演出が Application 境界へ逆流する。
2. 文字列中に「（1 秒休む）」等を埋める案 — 朗読される。TTS 入力は speakable 原稿のみに限定すべき。
3. 無音 insert なしで WAV を直結する案 — segment 分割の目的（間の pause）が失われる。
