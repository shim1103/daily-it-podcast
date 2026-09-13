---
name: production の OS/HTTP 既定値は internal/runtime。ffmpeg Encoder は infra。Composition は結線のみ
date: 2026-09-13T17:38:37
branch: feature/generator-audio-mp3-encode-write
---

## 1. Decision

1. production の外側 primitive 工場（`*http.Client`・`LookupEnv`・`LookPath`・subprocess `Run`・表示 `*time.Location`）の置き場は **`internal/runtime`** とする。Composition Root はここから受け取って Adapter / config / UseCase へ inject するだけにする。
2. **`infrastructure/audio/ffmpeg` は Port（`WAVToMP3Encoder`）の Driven Adapter のまま**とする。encode の意味（成功条件・失敗の Infrastructure Error・将来の ffmpeg argv）を持つ。OS/`os/exec` 実装は持たない。
3. **`infra/runtime` は作らない。** OS 工場と Port Adapter 層を同名で混ぜない。
4. ffmpeg Encoder を `runtime` や Application helper へ移さない。外側 tool（PATH 上の binary）依存は Adapter、純 `[]byte` 変換だけが Application helper。
5. 先行 Decision `2026-09-13T16-32-57` の「OS 実装は Composition/`runtime.go` に閉じる」のうち **置き場が `composition/runtime.go` である部分**を本 Decision が supersede する。inject の向き（Composition が差し、Application は `os`/`exec` を持たない、ffmpeg は別 Port）は維持する。
6. 先行 Decision `2026-08-29T13-48-58` の「`runtime.go` は vendor 非依存の production 既定値だけ」の趣旨は維持し、その file の所属を `internal/runtime` へ移す（path の正本は本 Decision）。

## 2. Reason

1. Composition Root の責務は結線である。`sharedCommandRun` のように buffer・stderr を扱う実装が `composition/runtime.go` に厚く居ると、結線と Drivers 工場が 1 package に混ざる。工場を `internal/runtime` へ出せば Composition は「渡す」だけに戻る。
2. `audio/ffmpeg` は `WAVToMP3Encoder` を実装する。これは UseCase 語彙の外側能力であり、HTTP の Gemini Adapter と同型の Driven Adapter である。依存する外側は SaaS ではなく OS 上の `ffmpeg` binary だが、環の外であることに変わりはない。Application helper（`ConcatWAV` 等）は外側 I/O が無いものに限る。
3. `infra/runtime` に工場を置くと、「Port を実装する Adapter」と「Client/`exec` を生成する工場」が同じ `infrastructure` 名の下に並び、読み手がどちらも Adapter だと誤読する。工場は Composition が使う最外 helper であり、Port 実装ではない。
4. HTTP を Composition 側工場が持つのは「標準ライブラリだから」ではない。`*http.Client` の production 既定を誰が選ぶかの問題であり、vendor 話し方は引き続き各 infra Adapter が持つ。

## 3. Rejected

1. ffmpeg Encoder を `internal/runtime` へ移す案 — Port 実装と OS 工場が同居し、encode 意味が Drivers 工場に漏れる。
2. `infrastructure/runtime` に工場を置く案 — Adapter 層と工場が同名空間で混ざる。
3. 工場を `composition/runtime.go` に残し file 分割だけする案 — 短期の整理には足りるが、subprocess `Run` を単独検証・Broad から再利用する口が Composition の unexported に閉じたまま残る。exported な `internal/runtime` の方が結線特権と工場の境界がはっきりする。
4. encode を Application `build` helper のまま `exec` する案 — 先行 `2026-09-13T16-32-57` Rejected と同型。
