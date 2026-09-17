---
name: TTS の model 切り替え fallback を source 配列化し、部分成功を持ち越す
date: 2026-09-16T11:41:59
branch: feature/generator-tts-fallback
---

## 1. Decision

1. `application/speech`（新設 package）に、TextWriter の `manuscript.TextWriter` と同型の合成 layer `SpeechSynthesizer{sources []port.SpeechSynthesizer, fallback port.FallbackReporter}` を置く。責務は model 切り替え fallback のみ。
2. `port.SpeechSynthesizer.SynthesizeAll` の契約を変える：**失敗時も、それまでに合成できた `[]models.SpeechAudio` を返す**（現行は `nil, err` で部分成功を握り潰していた）。
3. 合成 layer の fallback ループは、ある source が `port.ErrSourceExhausted`（TextWriter と共通の番兵型）を返したら、その source が返した部分成功分を保持し、**未処理の texts だけを次 source へ渡して続きから合成**する。
4. `port.ErrSourceExhausted` は TextWriter と同一の型を共有する。TTS 専用の番兵は新設しない。
5. `SpeechSynthesizer`（Infrastructure 実装）の constructor に、config-name による暗黙分岐ではなく明示的な `Tier`（`TierFree` / `TierPaid`）引数を追加する。
6. TTS の `callGap` / `MaxAttempts` / `SynthesizeBudget` は、AI Studio 実測値（`gemini-3.1-flash-tts-preview`: RPM=10, RPD=10）を根拠に導出し直す。`callGap = 60 / RPM` の式を正本にする。

## 2. Reason

1. TextWriter の fallback 設計（source 配列 + 単一 fallback ループ）と同型にする理由：primary/spare のような「独立した取得元を順に試す」という性質は TextWriter と TTS で同じであり、同じ構造を別 package として再利用する方が、fallback ロジックの正しさを 1 箇所で保証できる。
2. TextWriter とは異なり、TTS の `SynthesizeAll` は複数 segment（`texts []string`）を 1 回の呼び出しで処理する。途中の segment まで合成できた状態で source が枯渇した場合、**その部分成功を捨てて最初からやり直すと、既に消費した quota が無駄になる**。部分成功を返す契約にすることで、次 source は残りの segment だけを処理すればよく、fallback 発火のたびに全 segment を再合成しない。
3. `port.ErrSourceExhausted` を TextWriter と共有するのは、両者とも「この取得元は当面使えない、別の取得元があるなら切り替えてよい」という同じ意味論を持つため。TTS 専用の番兵を新設すると、同じ意味の状態を 2 つの型で表現することになり DRY に反する。
4. `Tier` を明示引数にする理由：どの key が free か paid かを config の変数名（`GEMINI_API_KEY` か `SPARE_GEMINI_API_KEY` か）から暗黙に判定すると、composition 層の結線を見ないと tier が分からない。呼び出し側が明示的に `Tier` を渡すことで、tier の切り替えが constructor の引数として読める。
5. 現行 `callGap=20s` は「無料枠 3 RPM」という前提（Decision `2026-09-02T13-56-00`）で導出されていたが、AI Studio の実測値は RPM=10 であり、前提が古い。`60/RPM` という式を正本にすることで、RPM が変わった時に値を手計算せず済む。

## 3. Rejected

1. **TTS 専用の枯渇番兵を新設する案** — `port.ErrSourceExhausted` と同じ意味論を別型で表現することになり DRY に反する。
2. **部分成功を返さず、失敗時は全 segment を最初から次 source でやり直す案** — 既に消費した quota が無駄になる。TTS は 1 回の呼び出しが有料枠を含む quota を消費するため、無駄な再合成のコストが TextWriter より大きい。
3. **`Tier` を config の変数名から暗黙に判定する案** — composition 層の結線を見ないと、どの key がどの tier か分からない。明示引数の方が読みやすい。
