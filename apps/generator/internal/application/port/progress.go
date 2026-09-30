package port

// ProgressReporter は produce の各段階の開始と完了を観測面へ知らせる。
// Start は段階に入った時（結果はまだ無い）、Done は段階を抜けた時（detail = 結果サマリ、空可）。
// 失敗時は Done を呼ばない。Start だけが log に残り、どの段階で落ちたかを示す。
type ProgressReporter interface {
	Start(step string)
	Done(step, detail string)
}

// FallbackReporter は取得元切替などの観測イベントを 1 件知らせる。
type FallbackReporter interface {
	Fallback(event string)
}

// RetryReporter は retry ループ内で 1 attempt が失敗し次 attempt へ進む直前の
// 途中経過を知らせる。成功時・全 attempt 失敗時は呼ばない（成功は
// ProgressReporter.Done、最終失敗は呼び出し元の error 返却で表現する）。
//
// @require retry != nil（Composition Root の結線責務）。
type RetryReporter interface {
	Retry(step string, attempt, max int, reason string)
}
