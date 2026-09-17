package port

import "errors"

// RejectionPrefixText は invalid-draft retry の brief に前回 attempt の raw response を
// 埋め込む直前に挟む固定文言。manuscript.TextWriter（model 切り替え時）と geminiapi.Write
// （同一 model 内の invalid-draft retry）の両方が参照する SSOT（Decision 2026-09-16T13-06-32 §1-3）。
const RejectionPrefixText = "\n\n# Previous attempt rejected\n"

// RejectionMiddleText は前回 attempt の raw response と buildFn の error メッセージの間に挟む
// 固定文言。raw 本文と error メッセージを区別可能な形で埋め込むための境界マーカー。
const RejectionMiddleText = "\n\n# Rejection reason\n"

// RejectionSuffixText は buildFn の error メッセージの直後に付ける固定文言。次 attempt へ
// 検証失敗の解消を指示する。
const RejectionSuffixText = "\n上記の検証失敗をすべて解消せよ。topics 件数・各 field 文字数・日本語・末尾句点を満たし、JSON オブジェクトのみを出力せよ。\n"

// BuildRejectionBrief は brief へ raw response と rejection 理由を
// RejectionPrefixText/RejectionMiddleText/RejectionSuffixText で挟んで埋め込む。
// manuscript.TextWriter と geminiapi.Write の両方が invalid-draft retry 用 brief の組み立てに使う SSOT。
func BuildRejectionBrief(brief, raw, rejectionReason string) string {
	return brief + RejectionPrefixText + raw + RejectionMiddleText + rejectionReason + RejectionSuffixText
}

// BuildRejectionBriefWithoutRaw は BuildRejectionBrief と同じ
// Prefix/Middle/Suffix 構造を、直前 attempt の raw response が無い場合向けに使う。
// raw を持たないぶん RejectionMiddleText の直前に raw を挟まない。
func BuildRejectionBriefWithoutRaw(brief, rejectionReason string) string {
	return brief + RejectionPrefixText + RejectionMiddleText + rejectionReason + RejectionSuffixText
}

// LastAttempt は invalid-draft retry ループの直前 attempt が残した部分情報である。
// generateContent 自体が retry しない error を返してループを抜ける場合でも、直前 attempt の
// raw response と buildFn の error（あれば）を破棄せず持ち越すために使う
// （Decision 2026-09-16T13-06-32 §1-1）。TextWriter 実装は fmt.Errorf("%w: %w", cause, LastAttempt{...})
// の形で返す error の chain に含め、呼び出し側は errors.As で取り出す。
//
// @invariant Raw と BuildErr は「直前 attempt で generateContent が成功し buildFn が invalid と
//
//	判定した」ときにのみ両方が非ゼロ値になる。直前 attempt が無い、または直前 attempt の
//	generateContent 自体が失敗していた場合は zero value のまま（この型を chain へ含めない）。
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

// ErrSourceExhausted は TextWriter 実装が「この取得元は当面使えない（利用枠喪失・認証断）。
// 別の取得元があるなら切り替えてよい」と示す番兵である。原稿の取得元 vendor を問わない。
//
// 実装は自分の失敗を fmt.Errorf("%w: %w", port.ErrSourceExhausted, <infra error>) の形で
// wrap して返す。呼び出し側は errors.Is(err, port.ErrSourceExhausted) で切り替え可否だけを見る。
var ErrSourceExhausted = errors.New("text writer source exhausted")

// ErrDraftRejected は TextWriter 実装が「invalid-draft retry を使い切ってもなお
// buildFn が valid と認める draft が得られなかった」と示す番兵である。
//
// ErrSourceExhausted（取得元そのものが使えない）とは意味が異なる：この番兵は取得元は生きているが
// 応答内容が繰り返し invalid だった状態を示す。呼び出し側は次の取得元へ渡す brief を、
// error chain から errors.As で LastAttempt を取り出せるかどうかで決める（どちらの番兵でも
// LastAttempt が取れれば rejection 理由を織り込み、取れなければ素の brief のまま渡す）。
// 実装は自分の失敗を fmt.Errorf("%w: %w", port.ErrDraftRejected, <last build error>) の形で
// wrap して返す。呼び出し側は errors.Is(err, port.ErrDraftRejected) で判定する。
var ErrDraftRejected = errors.New("text writer draft rejected after max attempts")
