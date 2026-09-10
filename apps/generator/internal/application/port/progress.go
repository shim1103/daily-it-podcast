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
