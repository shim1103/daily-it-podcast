package models

import "time"

// Post は PostSource が返すオリジナル投稿 1 件。
// 反応数・lang・author profile・Reply/Repost/引用は含めない。
type Post struct {
	ID        string    // tweet id。重複排除 key
	AuthorID  string    // 投稿者の X user id（WatchUserIDs と対応）
	Text      string    // 本文
	CreatedAt time.Time // 投稿時刻
	Permalink string    // 投稿 URL
	URLs      []string  // 本文中リンクの expanded URL。無ければ空
	Media     []Media   // 添付。無ければ空
}

// Media は Post が参照する添付 1 件。
type Media struct {
	Type string // 例: photo / video。vendor 表記を Domain 向けに正規化した値
	URL  string // 取得可能なメディア URL
}
