package models

import "time"

// SourceBody は Purpose 側（Detail=事実 / Discourse=議論）の本文と、その Purpose 用 URL。
// Text と Links がともに空なら、その Purpose 経路は無い。
type SourceBody struct {
	Text  string
	Links []string
}

// Empty は Text も Links も無いとき true。
func (b SourceBody) Empty() bool {
	return b.Text == "" && len(b.Links) == 0
}

// SourceItem は情報源 1 件。粒度は Adapter が決める。
// OccurredAt は情報源データに付いている発生時刻（UTC）であり、取得時刻でも呼び出し now でもない。
type SourceItem struct {
	SourceID   string
	OccurredAt time.Time
	Summary    string     // 短い要約・題名側。item があるなら非空は Adapter @ensure
	Detail     SourceBody // purpose1（事実）側。記事 URL は Links
	Discourse  SourceBody // purpose2（議論）側。議論ページ URL は Links。Text は 1 塊
	Meta       string     // item_id / actor 等の帰属のみ。空可
}
