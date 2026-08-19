package models

import "time"

// SourceItem は情報源 1 件。粒度は Adapter が決める。
// OccurredAt は情報源データに付いている発生時刻（UTC）であり、取得時刻でも呼び出し now でもない。
type SourceItem struct {
	SourceID   string
	OccurredAt time.Time
	Context    string
}
