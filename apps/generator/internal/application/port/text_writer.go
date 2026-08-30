package port

import "context"

// TextWriter は brief（本文の要約/依頼文）からテキスト断片を生成する。
//
// @require brief は trim 後に非空。生成対象は brief のみ。モデル/voice 等の vendor 設定を渡さない。
// @ensure 成功時は非空の text 断片を返す。戻りの byte 一致は保証しない。
// @invariant ManuscriptDraft や完成 manuscript.schema.json を露出しない。vendor 固有の CLI envelope、
//
//	stdout/stderr 形式、exit code を露出しない。method は Write のみ。
type TextWriter interface {
	Write(ctx context.Context, brief string) (string, error)
}
