---
name: TTS へ渡す朗読 text を「greeting+intro / topic ごと preface+detail / summary+farewell」の topic+2 束にする
date: 2026-09-02T13:55:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `build.SpeechTexts` が Gemini TTS へ渡す朗読 text 列を、従来の `2 + 2×topicCount + 2` 本から **`1 + topicCount + 1` 本**へ束ねる。
   1. 先頭 1 本 = greeting と intro を連結。
   2. 中間 topicCount 本 = 各 topic の preface と detail を連結（topic 順）。
   3. 末尾 1 本 = closingSummary と farewell を連結。
2. `build.Timeline` の期待 segment 本数と `topicStartSecs` の意味を束ね後の単位へ合わせる。`topicStartSecs[i]` は i 番目 topic 束（preface+detail 連結）の開始累積秒。
3. 連結 delimiter は改行 1 個とする。text 内容・原稿検証・`manuscript.schema.json` は変更しない。
4. 原稿 Domain 定数（`entities/constants` の `Draft*`。文字数・秒数・topic 数・`CharsPerSecond`）は本 Decision の対象外で、変更しない。

## 2. Reason

1. System e2e の緑化 blocker は Gemini TTS の無料枠 **RPD=15**（`gemini-2.5-flash-preview-tts` の実測: 3 RPM / 10,000 TPM / 15 RPD）。1 run が発行する `Synthesize` 回数は `SpeechTexts` の本数がそのまま決める。topic 5 なら旧 14 本で、retry 前に既に RPD の大半を食う。
2. 束ね後は topic 5 で 7 本、topic 3 で 5 本。1 run の最小 request 数を半減以上でき、有料枠移行後も無料枠 quota 内で回せるかを検証しやすくなる。
3. 尺計算（`WavDurationSec` → `Timeline`）が要求する単位は「1 回の Synthesize が返す 1 本の WAV」である。preface と detail を 1 本へ束ねても、detail 単独の尺は元々 `startSec` に含めない仕様なので、`topicStartSecs` は束の先頭（＝従来の preface 開始秒）と一致し、timeline の外形は変わらない。greeting+intro / summary+farewell も同様に「開始秒を持たない固定 segment」なので束ねても timeline に影響しない。
4. delimiter を改行にするのは、TTS が 2 文を続けて読む自然さを保ちつつ、束ね境界を原稿本文へ埋め込まないため。空文字連結だと文が繋がって読まれる。

## 3. Rejected

1. farewell だけ削って `2 + 2×topicCount + 1` にする案 — 削減幅が 1 本だけで RPD=15 に対して効果が薄い。挨拶の締めが消えて product の空気感（Opening/Closing template で確定済み）も崩れる。
2. topic 数の Domain 定数（`DraftTopicCountMin` 等）を下げて segment を減らす案 — System 都合で製品尺の意図を変えることになり、尺モデル Decision（`2026-08-30T03-06-53`）と衝突する。shim の明示許可が無い定数変更にあたる。
3. System test 実行時だけ env で segment 上限を絞る override を入れる案 — `ProduceEpisode` に test 専用分岐が生まれる。束ね自体が本番でも request 削減として妥当なので、分岐を足さず全経路で束ねる方が単純。
4. TTS を 1 request 1 本のまま `Synthesize` 間に長い sleep を入れて RPM だけ回避する案 — RPD=15 は sleep では超えられない。request 総数を減らさない限り無料枠に収まらない。
