---
name: TextWriter は ManuscriptDraft、SpeechSynthesizer は 1 text 1 WAV のまま薄く保つ
date: 2026-08-25T22:37:29
branch: feature/generator-cmd-usecase-boundary
superseded_by: 2026-08-25T23-02-35-feature-generator-cmd-usecase-boundary.md
---

> **上書き**: TextWriter の戻りを ManuscriptDraft とする結論は `docs/decisions/2026-08-25T23-02-35-feature-generator-cmd-usecase-boundary.md` が正。SpeechSynthesizer の薄さと定型/TTS 順が ProduceEpisode である側面は同 Decision が維持する。

## 1. Decision

1. `TextWriter` の成功戻りは完成 `manuscript.schema.json` ではなく、途中型 `ManuscriptDraft`（Intro、Topics の Title/Preface/Detail、ClosingSummary）とする。Port 署名の正は A（Contract Freeze）側 artifact。
2. `SpeechSynthesizer` は 1 朗読 text → 1 WAV のままとする。segment 配列・結合・定型挨拶は Port に載せない。
3. 定型挨拶と Draft の結合、TTS への渡し順、尺の書き込みは `ProduceEpisode`（Builder）が行う。
4. 原則の正は architecture `backend/application.md`（Port は能力のみ / 途中型 ≠ 完成契約）および既存 `2026-08-17T17-41-59`（Synthesize 入力は本文のみ）。

## 2. Reason

1. 完成 schema は UI・Drive 読取の完成形である。LLM 1 回の戻りに timing・episodeId・定型挨拶まで含めると、生成単位と永続単位が一致し、Port が完成契約を背負う。Draft に落とすと TextWriter の変わり方（生成）と WriteEpisode の変わり方（検査）が分離する。
2. 定型挨拶を Adapter や Port に置くと、台本方針の変更が vendor I/O 境界へ波及する。Application の Builder に置けば Cursor / Gemini を変えずに定型だけ変えられる。
3. TTS 結合を `SpeechSynthesizer` に置くと、1 能力 Port が episode 構造と尺方針を知り、ISP と「自前分割で timestamp」決定（`2026-08-25T06-56-00`）と衝突する。

## 3. Rejected

1. `TextWriter` を完成 manuscript schema と 1:1 にする案 — 既存 `2026-08-18T16-30-00` の Rejected と同型。timing 未確定の生成結果と永続完成形が混ざる。
2. `SpeechSynthesizer` に segment 配列や concat を持たせる案 — Port が朗読方針と結合を知る。
3. 定型挨拶を Cursor Adapter 定数に閉じる案 — 台本方針が vendor Adapter に埋まる。
