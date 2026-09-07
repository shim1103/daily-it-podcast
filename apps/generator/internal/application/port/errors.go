package port

import "errors"

// ErrSourceExhausted は TextWriter 実装が「この取得元は当面使えない（利用枠喪失・認証断）。
// 別の取得元があるなら切り替えてよい」と示す番兵である。原稿の取得元 vendor を問わない。
//
// 実装は自分の失敗を fmt.Errorf("%w: %w", port.ErrSourceExhausted, <infra error>) の形で
// wrap して返す。呼び出し側は errors.Is(err, port.ErrSourceExhausted) で切り替え可否だけを見る。
var ErrSourceExhausted = errors.New("text writer source exhausted")
