package application

// wavDurationSec は RIFF/WAV header と data から再生尺（秒）を返す。
// ProduceEpisode が各朗読 WAV の startSec / durationSec 確定に使う unexported helper。
//
// @require wav は非空。
// @ensure 成功時 durationSec >= 0。
// @invariant vendor 定数を使わず header から読む。失敗時は entities/errors.CorruptSpeechAudio を返す。
func wavDurationSec(wav []byte) (float64, error) {
	panic("wav duration: contract stub; logic is C")
}
