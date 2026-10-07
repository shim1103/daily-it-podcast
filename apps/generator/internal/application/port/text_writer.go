package port

import (
	"context"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// TextWriter は brief（本文の要約/依頼文）から ManuscriptDraft を生成する。
//
// @require brief は trim 後に非空。生成対象は brief のみ。モデル/voice 等の vendor 設定を渡さない。
// @require buildFn は非 nil。生の応答文字列を models.ManuscriptDraft へ解釈する関数を呼び出し側（Application 層）がDI で渡す。
//
// @ensure 成功時は buildFn が返す非 nil な models.ManuscriptDraft を返す。
// @ensure この取得元が当面使えない時は ErrSourceExhausted を、buildFn が valid と認める draft を得られなかった時は
// ErrDraftRejected を wrap して返す。どちらも別の取得元へ切り替えてよい合図で、後者は直前の失敗を LastAttempt として chain へ含める。
// @invariant 再試行するかどうか、再試行の条件は実装の裁量である。
type TextWriter interface {
	Write(ctx context.Context, brief string, buildFn func(string) (models.ManuscriptDraft, error)) (models.ManuscriptDraft, error)
}
