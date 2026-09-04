---
name: SpeechTexts 束境界は改行3個、topic.detail 内の段落改行は最大1個
date: 2026-09-04T17:05:00
branch: feature/playback-e2e-redeploy-master
---

## 1. Decision

1. `build.SpeechTexts` の束境界 delimiter（greeting↔intro / preface↔detail / closingSummary↔farewell）を **改行 3 個**（`"\n\n\n"`）にする。先行 Decision `2026-09-02T13-55-00` の「改行 1 個」を supersede する。
2. Cursor brief prompt（`TextWriterBriefPrompt`）は `topic.detail` について次を指導する。段落分けが必要なときだけ **改行は 1 個まで**。それ以外は段落分けせず、改行や文間の無意味な空白は入れない。
3. 上記の detail 改行規約は **validation で強制しない**（YAGNI）。prompt 指導のみ。

## 2. Reason

1. TTS へ渡す transcript は「本文だけ」方針を保ちつつ、field 境界（greeting / intro 等）をはっきりさせたい。改行を増やすと境界の間隔が読み取りやすく、TTS の間にも寄与しうる。
2. Cursor が `detail` 内で入れる段落改行（最大 1 個）と、draft 組み立てが入れる field 境界（改行 3 個）を数えで区別できる。同じ 1 個だと境界か段落かが曖昧になる。
3. detail の連続改行や文間空白を機械検証しても、失敗時の再試行コストに見合う再発防止が今は薄い。prompt で足りる間は validation を足さない。

## 3. Rejected

1. 束境界を半角 space や句読点メタで埋める案 — 本文方針を壊し、TTS が記号を読み上げるリスクがある。
2. detail の改行数を Domain validation で落とす案 — YAGNI。prompt で誘導できるうちは検証を増やさない。
3. 束境界を改行 2 個のままにする案 — detail 内の「段落＝改行 1」と差が小さく、空行（改行 2）と紛らわしい。
