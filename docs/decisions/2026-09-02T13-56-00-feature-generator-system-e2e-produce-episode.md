---
name: Gemini retry は「同種 error が 2 回連続したら打ち切り」、MaxAttempts=3・callGap=20s
date: 2026-09-02T13:56:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `gemini.SpeechSynthesizer.Synthesize` の retry loop に **同種連続打ち切り**を入れる。直前の失敗と今回の失敗の Op（`*gemini.Error.Op`）が一致し、かつ retryable なら「同種の連続失敗」と数え、**2 回連続で loop を打ち切って最後の error を返す**。Op が変われば連続数はリセットする。
2. `gemini.MaxAttempts` を 6 から **3** へ下げる。異なる Op が交互に来る最悪ケースの上限として残す。
3. `gemini.callGap`（成功・失敗を問わない `client.Do` 間の最小間隔）を 5s から **20s** へ上げる。無料枠 3 RPM = 20s 間隔に合わせる。
4. `retryBackoffBase`（60s）・`retryBackoffMax`（3m）・`parseRetryAfter`・`suggestedWait` 優先は変更しない。
5. retry を無効化・削除はしない。

## 2. Reason

1. 現状の loop は retryable フラグが立てば中身に関わらず `MaxAttempts`（6）回まで回す。`decode_pcm: output audio is missing`（Gemini TTS preview の既知 Limitation。応答が時々 audio 無しで返る）が 1 セグメントで 6 request を焼き、無料枠 RPD=15 を単独で使い切る。
2. `output audio is missing` が 2 回連続で出るなら、その本文に対しては決定論的に失敗していると見なして良い。一過性なら 1〜2 回目で成功するか、別 Op（429 / 503）が挟まって連続数がリセットされる。よって「同種 2 連続で打ち切り」が一過性と決定的を分ける基準になる。
3. 無料枠 RPD=15 由来の 429 は日次リセット（太平洋時間 0:00 = JST 17:00）まで回復しない。CI job は 50 分なので、429 が連続したときに回復を待つ意味は無く、他の retryable と同じく 2 連続で打ち切ってよい。RPM 由来の 429 は 60s backoff 1 回で回復することが多く、2 回試せば足りる。429 を特別扱いする分岐は不要になる。
4. `callGap=5s` は 3 RPM（20s 間隔）を満たさず、連続セグメントで即 429 を誘発していた。20s なら無料枠の RPM を守れ、有料枠でも安全側。
5. `MaxAttempts=3` は、同種連続打ち切りが入れば実効上限は「2 種類の Op が交互 = 3 回」程度で、6 は過剰。

## 3. Rejected

1. `MaxAttempts` だけ下げて連続打ち切りを入れない案 — 3 回でも `output audio missing` が 3 連続すれば 3 request を無駄に焼く。一過性と決定的を区別できない。
2. 429 の Op を retry 対象から外す案 — RPM 由来の一過性 429 まで即失敗になる。連続打ち切りがあれば区別せず扱える。
3. response body の `quota_metric`（PerDay / PerMinute）を parse して RPD 429 だけ即打ち切る案 — body 形式への依存が増える。連続打ち切りで同じ効果が得られ、body parse を持ち込まない方が単純。
4. `callGap` を有料枠 RPM（10 RPM = 6s）基準にする案 — 検証目的が「無料枠 quota を超えないこと」なので、緩い方（20s）に固定して 3 RPM を守る。
