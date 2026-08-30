---
name: Composition は Gate UseCase または Port を返し、全 Port の UseCase 包みを強制しない
date: 2026-08-25T22:37:30
branch: feature/generator-cmd-usecase-boundary
---

## 1. Decision

1. Composition factory の戻りは、永続ゲートがあるとき Application UseCase（例: `NewGoogleDriveWriteEpisode` → `WriteEpisode`）、包む方針が無いとき Port（例: `NewCursorTextWriter` → `TextWriter`、`NewGeminiSpeechSynthesizer` → `SpeechSynthesizer`）とする。
2. 生成方針の結線入口は `NewProduceEpisode` とし、Cursor / Gemini をそれぞれ太い原稿・音声 UseCase で包んで対称化しない。
3. factory 名に vendor を付けてよいのは Composition のみ。Application 型名には付けない。
4. 原則の正は architecture `backend/composition-root.md` の結線単位の戻り。

## 2. Reason

1. `WriteEpisode` を Composition が返すのは、検証を迂回できる生 `EpisodeWriter` を入口に出さないためである（lesson 109）。同じ理由が TextWriter / SpeechSynthesizer には無い。包む Application 方針は `ProduceEpisode` が既に持つ。
2. 全 Port を UseCase で包む対称化は、見た目の揃いのために方針を二重化するか、上位を空洞にする。Composition の選ぶ軸は「迂回禁止のゲートがあるか」であり「Infra 名が付くか」ではない。
3. Composition に生成手順を書くと、結線点にビジネス方針が混ざり、Application と Composition の境界が再び曖昧になる。

## 3. Rejected

1. `NewCursor…Manuscript` / `NewGemini…EpisodeAudio` を必ず新設し、`ProduceEpisode` を順送りだけにする案 — 対称化のための分割。方針が分散する。
2. Composition 内で brief 組み立てや TTS ループを書く案 — 結線点に手順が混ざる。
3. Application の型名に GoogleDrive / Cursor / Gemini を含める案 — 内側が vendor 結線を名前で知ることになる。
