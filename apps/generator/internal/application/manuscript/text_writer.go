// Package manuscript は原稿取得の UseCase を提供する。
// primary の TextWriter が「取得元の枯渇」を示したときだけ secondary へ 1 回だけ切り替える。
package manuscript

import (
	"context"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

var _ port.TextWriter = (*TextWriter)(nil)

// TextWriter は primary → secondary の順で原稿断片を確保する UseCase である。
//
// primary が port.ErrSourceExhausted を wrap した error を返したときだけ secondary へ 1 回切り替える。
// primary / secondary の vendor は知らない。切り替え可否は番兵 error だけで判断する。
type TextWriter struct {
	primary    port.TextWriter
	secondary  port.TextWriter
	onFallback func()
}

// NewTextWriter は primary / secondary と、切り替え発生時の通知関数を束ねる。
//
// @require primary != nil かつ secondary != nil かつ onFallback != nil。
// @ensure 戻りは *TextWriter（port.TextWriter を満たす）。
func NewTextWriter(primary, secondary port.TextWriter, onFallback func()) *TextWriter {
	return &TextWriter{primary: primary, secondary: secondary, onFallback: onFallback}
}

// Write は primary を試し、port.ErrSourceExhausted のときだけ secondary を 1 回だけ呼ぶ。
//
// @require brief は trim 後に非空。
// @ensure primary が成功したらその断片を返し、secondary を呼ばない。
// @ensure primary の error が errors.Is(err, port.ErrSourceExhausted)==true のとき、onFallback を 1 回呼んでから secondary.Write を 1 回だけ呼び、その戻り（成功・失敗を問わず）を返す。
// @ensure primary の error が errors.Is(err, port.ErrSourceExhausted)==false のとき、その error をそのまま返し secondary を呼ばない。
// @invariant 切り替えは高々 1 回（secondary が再び port.ErrSourceExhausted を返しても 3 つ目は無い）。secondary の error を別型でラップしない。secondary へ渡す brief は primary へ渡したものと同一。
//
// TODO(generator-text-writer-fallback): 本体は C で実装する。現状は stub。
func (w *TextWriter) Write(ctx context.Context, brief string) (string, error) {
	return "", nil
}
