---
name: 記事source5種はhttpget.GetWithRetry経由でretryを持つためRetryReporter対象に含める
date: 2026-09-26T12:54:24
branch: feature/generator-logging-mature
---

## 1. Decision

1. 先行Decision（`2026-09-25T18-24-04-feature-generator-logging-mature.md`）§1-4「記事source 5種（hackernews/lobsters/publickey/techcrunch/cloudwatch）はretryを持たないため`RetryReporter`をDIしない」を訂正する。実際には5種全部が`httpget.GetWithRetry`という共有ヘルパー経由で「5xx/client.Do errorを1回だけ即再試行」というretryを持っており、この前提は誤りだった。
2. `httpget.GetWithRetry`のsignatureへ`retry port.RetryReporter, step string`を追加し、5source全部のconstructor（`NewListItemSource`）へ`RetryReporter`をDIする。`audio/ffmpeg`は本当にretryを持たない（contract commentに明記済み）ため対象外のまま変更しない。
3. 先行Decision§2-4・§3-2の「記事source・ffmpeg」という並記は「ffmpegのみ」を指すものとして読み替える。本文の書き換えは行わず、本Decisionが訂正の正本となる。

## 2. Reason

1. 先行Decisionの調査時点では`item_source.go`の`getWithRetry`という薄いwrapperメソッド名から「retryを持つように見えるが、実際には確認していない」状態で「retryなし」と判断していた。実際に`httpget.go`を読むと`GetWithRetry`という名の通りretry機構（5xx/Do errorへの1回だけの即時再試行）を持つ共有関数であることが判明した。
2. `r2.retryLoop`と同型の「共有generic/共通skeleton関数に1箇所実装することで複数呼び出し元へ波及させる」設計が可能であり、既に`r2`で採用した設計と対称にできる。

## 3. Rejected

1. 先行Decisionの記述をそのまま残し、本Decisionだけを別問題として扱う案 — `decision.md`§8-4が指摘する「前提が崩れた文言は指示対象を失った空文になる」状態を放置すると、後続の読み手が先行Decisionの§1-4を字面通り適用し「記事sourceは対象外」と誤解する。本Decisionで明示的に訂正範囲を示す。
