package port

import "context"

// CompletedEpisodeLookup は表示用 date に対する完成 episode ペアの有無を照会する。
// Drive 配置・file id・MIME は Infrastructure に閉じる。
//
// @require date は YYYY-MM-DD 暦日文字列。
// @ensure 同一 stem の {stem}.json と {stem}.wav が両方あり、かつ json の date field が date と一致するとき true。
// @ensure 片方だけ・無し・date 不一致は false（error ではない）。
// @invariant stem・Drive file id・MIME・vendor 型を露出しない。method は HasPair のみ。
type CompletedEpisodeLookup interface {
	HasPair(ctx context.Context, date string) (bool, error)
}
