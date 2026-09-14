---
name: WAV→mp3 の ffmpeg/exec は Infrastructure Adapter が持ち、Composition が Port として inject する
date: 2026-09-13T16:32:57
branch: feature/generator-audio-mp3-encode-write
---

## 1. Decision

1. WAV→mp3 は **Application Port**（UseCase 語彙）経由とする。契約の正本は A（`application/port` の encoder interface と Infrastructure の ffmpeg Adapter）。`ProduceEpisode` は ConcatWAV 成功直後・WriteEpisode 直前に 1 回呼ぶ（呼び出し位置は先行 `2026-09-13T13-40-29` を維持）。
2. **ffmpeg 変換ロジックは Infrastructure Adapter が持つ。** `os` / `os/exec` の production 実装は **`internal/runtime`** が提供し、Composition が Adapter ctor へ差し、Adapter は受け取った `LookPath` / `Run` だけを使う。Application（UseCase / `build`）は `os` / `net/http` / `os/exec` を持たない。置き場の正本は Decision `2026-09-13T17-38-37`。
3. **Composition** が `internal/runtime` の production 既定値を Adapter / config へ渡し、Adapter を Port として `ProduceEpisode` へ inject する（向きは `sharedHTTPClient` / `sharedLookupEnv` 時代と同型。工場の所属は `internal/runtime`）。
4. **TTS Port / Gemini Adapter には encode を載せない**（先行維持）。
5. Adapter は **retry しない**。`ffmpeg` 不在・非 0 exit・空出力は即 error。
6. 先行 Decision `2026-09-13T13-40-29` の「**Port にも出さない**」だけを本 Decision が **supersede** する。手段が ffmpeg であること、TTS 戻りがセグメント WAV であること、完成形式が mp3 であることは維持する（本文へ再掲しない）。

## 2. Reason

1. Composition が `*http.Client` や `LookupEnv` を握って Adapter / config へ渡すのは「外側が外側の道具を差し込む」形である。UseCase や `build` が `exec.Command` を直持ちするのはその逆で、内側が Frameworks & Drivers を知る環違反になる。
2. 先行の「Port に出さない」は TTS Port / Gemini Adapter への encode 混入を避ける主旨だった。encode を **別 Port** にすれば TTS を汚さず、かつ Application から OS 依存を除去できる。
3. `build` の package 関数が `exec` を直呼びする案や、Application 内の unexported OS seam は、test 差し替えには足りても環の説明にならない。Sociable Unit の fake は他 Port と同様、Port DI で足りる。
4. local `ffmpeg` の失敗は決定論的であり、HTTP 一過性向けの retry を載せると失敗を隠すだけになる。

## 3. Rejected

1. `build.EncodeWAVToMP3` が `os/exec` を直呼びする案 — Application が Drivers を知る。
2. Application package 内の unexported OS seam で凌ぐ案 — 環を守らない妥協。Composition の inject 政策と非対称。
3. TTS Port / Gemini Adapter に encode を載せる案 — 結合・尺が Adapter 側へ漏れ、先行 `2026-08-25T22-37-31` および `2026-09-13T13-40-29` Rejected と同型。
4. encode を Port にせず、Composition が mechanism（runner 関数）を UseCase へ直接渡す案 — UseCase 語彙の境界が無く、`*http.Client` を UseCase に渡すのと同型の mechanism 漏れ。
5. ffmpeg 失敗への重装 retry — transient 前提が無く、決定論的失敗を粘るだけ。
