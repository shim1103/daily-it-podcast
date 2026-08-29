package build

// ConcatWAV は同一 PCM パラメータの WAV を 1 本に結合する。
// 隣接 part の間に entities/constants.SegmentSilenceSec 分の無音 PCM を挿入する。
//
// @require parts は 1 つ以上。各要素は非空 WAV。
// @ensure 成功時は非空の結合 WAV。形式は入力と同一 PCM パラメータに限る。無音 insert 分は durationSec / startSec 算定に含める。
// @invariant vendor 定数を使わず header から読む。失敗時は entities/errors.Error（Op = corrupt_speech_audio）を返す。
func ConcatWAV(parts ...[]byte) ([]byte, error) {
	panic("wav concat: contract stub; logic is C")
}
