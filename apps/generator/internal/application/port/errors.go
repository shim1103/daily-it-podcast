package port

import "errors"

// RejectionPrefixText は invalid-draft retry の brief で、前回 attempt の raw response の直前に挟む固定文言。
const RejectionPrefixText = "\n\n# Previous attempt rejected\n"

// RejectionMiddleText は raw response と buildFn の error メッセージの間に挟む境界マーカー。
const RejectionMiddleText = "\n\n# Rejection reason\n"

// RejectionSuffixText は buildFn の error メッセージの直後に付け、次 attempt へ検証失敗の解消を指示する固定文言。
const RejectionSuffixText = "\n上記の検証失敗をすべて解消せよ。topics 件数・各 field 文字数・日本語・末尾句点を満たし、JSON オブジェクトのみを出力せよ。\n"

// BuildRejectionBrief は brief へ raw response と rejection 理由を
// RejectionPrefixText/RejectionMiddleText/RejectionSuffixText で挟んで埋め込む。
// invalid-draft retry 用 brief の組み立ての SSOT である（Decision 2026-09-16T13-06-32 §1-3）。
func BuildRejectionBrief(brief, raw, rejectionReason string) string {
	return brief + RejectionPrefixText + raw + RejectionMiddleText + rejectionReason + RejectionSuffixText
}

// LastAttempt は invalid-draft retry の直前 attempt が残した raw response と buildFn の error を、
// retry を抜ける error の chain で次の取得元へ持ち越す（Decision 2026-09-16T13-06-32 §1-1）。
// TextWriter 実装は error の chain に含め、呼び出し側は errors.As で取り出す。
//
// @invariant Raw と BuildErr は「直前 attempt で generateContent が成功し buildFn が invalid と
// 判定した」ときにのみ両方が非ゼロ値になる。それ以外では、この型を chain へ含めない。
type LastAttempt struct {
	// Raw は直前 attempt で generateContent が返した response 文字列。
	Raw string
	// BuildErr は直前 attempt で buildFn が返した validation error。
	BuildErr error
}

// Error は LastAttempt を error chain の末端要素として使えるようにする。
//
// @ensure BuildErr が nil のとき panic せず固定文言を返す。
func (a LastAttempt) Error() string {
	if a.BuildErr == nil {
		return "previous attempt rejected: (no build error)"
	}
	return "previous attempt rejected: " + a.BuildErr.Error()
}

// ErrSourceExhausted は TextWriter 実装が「この取得元は当面使えない。別の取得元があるなら切り替えてよい」と
// 示す番兵である。原因（利用枠喪失・認証断・429・5xx・通信断など）と取得元 vendor を問わず、
// 利用枠が尽きた場合に限らない。
//
// 実装は fmt.Errorf("%w: %w", port.ErrSourceExhausted, <infra error>) の形で wrap して返す。
// 呼び出し側は errors.Is で切り替え可否だけを見る。
var ErrSourceExhausted = errors.New("text writer source exhausted")

// ErrDraftRejected は TextWriter 実装が「invalid-draft retry を使い切ってもなお
// buildFn が valid と認める draft が得られなかった」と示す番兵である。
// ErrSourceExhausted と違い、取得元は使えるが応答内容が繰り返し invalid だった状態を表す。
//
// 実装は fmt.Errorf("%w: %w", port.ErrDraftRejected, <last build error>) の形で wrap して返す。
// 呼び出し側は errors.Is で判定する。
var ErrDraftRejected = errors.New("text writer draft rejected after max attempts")
