---
name: TextWriter の invalid-draft retry は最後の試行結果を持ち越し、Gemini 分類は fetch 時点で確定する
date: 2026-09-16T13:06:32
branch: feature/generator-textwriter-adapter-fallback
---

## 1. Decision

1. `geminiapi.Write` の invalid-draft retry ループは、**generateContent 自体が retry しない error を返して抜ける場合でも、直前の attempt で得た raw response と buildFn の error（あれば）を持ち越す**。「成功」でも「情報を持たない完全な失敗」でもない、部分情報を保持した失敗状態を戻り値として区別できる形にする（先行 Decision `2026-09-16T11-41-26` §1-2 が定義した `Write` の戻り値契約を、この 1 点だけ部分 supersede する）。
2. invalid-draft の retry 用 brief は、`brief + 固定 prefix + 直前 raw + 固定文言 + buildFn の error` という構造で組み立てる。単純な `brief + err.Error()` の append ではなく、raw response 本文と error メッセージを別々の要素として区別可能な形で埋め込む。
3. retry prompt の固定文言（`"\n\n# Previous attempt rejected\n"` と `"\n上記の検証失敗をすべて解消せよ...\n"`）を定数化する。この文言は `manuscript.TextWriter`（model 切り替え時）と `geminiapi.Write`（同一 model 内の invalid-draft retry）の両方で使うため、どちらか一方が正本を持ち、他方は参照する。
4. TTS 側 `speech/gemini/error.go` の `wrapIfSourceExhausted`（fetchPCM の error メッセージを事後的に文字列走査して枯渇を判定する関数）を廃止する。枯渇か否かの分類は、HTTP response を直接見る `fetchPCM`（`transport.go`）がその場で確定させ、TextWriter 側 `geminiapi.fetchOnce` の `fetchRetryKind` と同型の分類 enum で返す。

## 2. Reason

1. 現行の `geminiapi.Write` は、generateContent が retry しない error（400 等）を返した時点で `models.ManuscriptDraft{}` という空値だけを返し、直前の attempt で得ていた raw response と invalid-draft の理由を完全に捨てていた。TTS 側の `SynthesizeAll` は既に「部分成功（audios）を保持したまま error を返す」設計（先行 Decision `2026-09-16T11-41-59`）にしており、TextWriter だけこの原則を欠いていたのは非対称である。得られた情報を捨てずに呼び出し元へ渡すことで、`manuscript.TextWriter` が次 source へのfallback判断に使える材料が増える。
2. invalid-draft の retry brief を raw + error の両方を区別可能な形で埋め込む理由：error メッセージ（validation の失敗理由）だけでは、次の attempt（または次 source）が「前回何を生成したか」を知らないまま同じ間違いを繰り返す可能性がある。raw response 本文も見せることで、次の生成が自分の前回出力を直接参照して修正できる。
3. 固定文言を定数化する理由：同じ文字列 literal が `manuscript.TextWriter` と `geminiapi.Write` の 2 箇所に手書きで重複しており、文言を変える際にどちらか一方だけ更新される drift のリスクがある（coding-style/naming.md §3 の Magic String 定数化規則）。
4. `wrapIfSourceExhausted` を廃止する理由：この関数は fetchPCM が既に見ている HTTP response の情報を、error メッセージへ変換してから文字列走査で分類し直すという、情報の once-lost-then-reconstructed な設計になっていた。TextWriter 側の `fetchOnce` は既に HTTP response を見た時点で `fetchRetryKind` という分類を確定させており、TTS 側だけ後段で再分類する非対称な設計を採る理由がない。分類を fetch 時点に一元化すれば、文字列走査という不確実な手段（response body の言い回しが変われば判定が壊れる）を避けられる。

## 3. Rejected

1. **invalid-draft retry で得た raw/error を破棄したまま、generateContent の retry しない error だけをそのまま返す案（現行）** — 得られた情報を無駄にする。TTS 側の部分成功保持という既存の設計原則と非対称になる。
2. **retry brief を `brief + err.Error()` の単純 append のままにする案** — raw response 本文が次の attempt へ渡らず、前回の出力を直接参照した修正ができない。
3. **固定文言を `manuscript.TextWriter` と `geminiapi.Write` の両方に手書きで残す案** — 2 箇所に同じ literal が重複し、どちらか一方だけ変更される drift リスクが残る。
4. **`wrapIfSourceExhausted` を維持し、判定条件（quota_exceeded 等の文字列）を精緻化する案** — 事後の文字列走査という設計自体を維持することになり、fetchPCM が既に持っている構造化された response 情報（HTTP status code、response body の code field）を活用しない非効率が残る。
