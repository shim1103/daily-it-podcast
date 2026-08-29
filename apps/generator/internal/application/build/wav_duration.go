package build

// WavDurationSec は RIFF/WAV header と data から再生尺（秒）を返す。
//
// @require wav は非空。
// @ensure 成功時 durationSec >= 0。
// @invariant vendor 定数を使わず header から読む。失敗時は entities/errors.Error（Op = corrupt_speech_audio）を返す。
func WavDurationSec(wav []byte) (float64, error) {
	panic("wav duration: contract stub; logic is C")
}
