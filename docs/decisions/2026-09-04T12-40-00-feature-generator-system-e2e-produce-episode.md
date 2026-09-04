---
name: TTS 呼び出し群の retry 予算を Adapter に集約し、SpeechSynthesizer port を複数 text 対応にする
date: 2026-09-04T12:40:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `port.SpeechSynthesizer` の method を `Synthesize(ctx, text string) (SpeechAudio, error)` から **`SynthesizeAll(ctx, texts []string) ([]SpeechAudio, error)`** へ変える。返すのはセグメント単位の WAV 列で、結合しない。
2. retry 予算・callGap・無料枠 RPD quota は Gemini Adapter が「1 度の `SynthesizeAll` 呼び出し（= 1 episode 分の TTS 呼び出し群）」で束ねて管理する。二段構え:
   1. 内側 = `MaxAttempts`（先行 Decision `2026-09-02T13-56-00`）。1 セグメントが連続で消費してよい上限。
   2. 外側 = `SynthesizeBudget`。1 度の `SynthesizeAll` 全体で許す Gemini 呼び出しの合計上限。各セグメントは `min(MaxAttempts, 残予算)` 回まで。合計が `SynthesizeBudget` へ達したら以降のセグメントは Client を叩かず即 error。
3. `build.Timeline`（各セグメント尺 → topic 開始秒）と `build.ConcatWAV`（WAV 列 → 1 本）は `ProduceEpisode.Run`（application）に残す。Adapter へ移さない。
4. Gemini が HTTP 200 で返す極小 PCM（最小尺閾値 `minPCMBytes` 未満）を `decode_pcm` 相当の retryable として Adapter のループ内で retry する。これにより Adapter が「非空かつ最小尺の WAV」を contract として保証し、上位（計測 test 等）は `err == nil` だけを成功条件にできる。
5. `MaxAttempts` / `callGap` / 同種 2 連続打ち切り（`2026-09-02T13-56-00`）と、topic+2 束の分割（`2026-09-02T13-55-00`）は変更しない。`SynthesizeBudget` の具体値は code の const が正本。

## 2. Reason

1. 従来は `ProduceEpisode.Run` が topic+2 束を for loop で 1 セグメントずつ `Synthesize` していた。`MaxAttempts` はセグメント内の上限でしかなく、セグメント数 × `MaxAttempts` が 1 episode の実効 retry 上限になる。topic 5 なら最悪 7 × 3 = 21 で、無料枠 RPD=15 を 1 episode で焼き切る。`callGap` は Adapter インスタンスの `lastCallAt` でセグメントを跨いで効いているのに、retry 予算だけが application 側に漏れて跨げていない非対称があった。
2. retry 予算・callGap・RPD quota はいずれも vendor（Gemini 無料枠）固有の制約で、対応する定数はすべて Adapter package にある。「1 episode 分の呼び出し群」で束ねて管理する主体は Adapter が自然で、application は「この原稿群を音声にせよ」だけ知ればよい。port を複数 text 対応にしても port の invariant（vendor 固有型・PCM・path・voice 名を露出しない）は保てる。
3. `Timeline` と `ConcatWAV` は同じ WAV 列を入力に取る双子で、両方 application の episode 組み立てロジック（vendor 非依存の純粋関数）。`ConcatWAV` を Adapter へ移すと、Adapter が `SpeechTexts` の束ね規約（greeting+intro 束 / topic 束 / closing+farewell 束）を知る必要が出て、application の原稿構造知識が infra へ漏れる。片方だけ切り出すと WAV 列の所有が 2 箇所に割れる。
4. 極小 PCM は「非空だが実質無音」で、Gemini preview の audio 欠落（`2026-09-02T13-56-00` Reason 2）と同じ一過性劣化。application 側で検知しても retry する術がない（`Synthesize` を呼び直すと callGap 待ち + attempt リセット、しかも「なぜ短かったか」は Adapter に閉じている）。retryable として Adapter のループ内で扱うのが既存の retryable 判定表（client.Do / 429 / 503 / 5xx / decode_pcm）と一貫する。

## 3. Rejected

1. `Synthesize`（単数）を保ち、application が 1 セグメントずつ呼ぶ現状維持 — retry 予算をセグメント跨ぎで管理できず、RPD 焼き切りリスクが残る。callGap だけ跨げて retry 予算は跨げない非対称も解消しない。
2. `Synthesize` の責務を拡張し、「texts を受け取り中で Synthesize と ConcatWAV をやって 1 本の WAV を返す」案 — Adapter が `Timeline` の segment 本数規約と原稿の束ね構造を知ることになり、application の知識が infra へ流出する層違反。
3. `Timeline` + `ConcatWAV` を「texts → 1 本 WAV」の新 application UseCase へ切り出す案 — `Timeline` は topic 境界計算に WAV 列を要るので、新 UseCase が 1 本にまとめた後 `ProduceEpisode` は尺を取れず、`Timeline` も新 UseCase へ道連れになり、結局その UseCase が原稿の topic 構造を知る。再利用需要も今は無い（YAGNI）。`ProduceEpisode.Run` は既に「WAV 列を受け取り Timeline → ConcatWAV」をやっており、port を複数 text 対応にする以外の構造変更を要しない案 3 の方が最小。
4. 極小 PCM を application 側の尺チェック（`WavDurationSec` の結果が閾値未満なら error）で弾く案 — retry できず即失敗になる。Adapter が最小尺を保証しないなら、実 Gemini を叩く計測 test も毎回「非空 WAV / 尺 > 0」を再確認する必要が残り、経路ごとに検査の有無の非対称が生まれる。
