package application

// concatWAV は同一 PCM パラメータの WAV を 1 本に結合する。
// ProduceEpisode が episode 全体の SpeechAudio.Content 生成に使う unexported helper。
//
// @require parts は 1 つ以上。各要素は非空 WAV。
// @ensure 成功時は非空の結合 WAV。形式は入力と同一 PCM パラメータに限る。
// @invariant vendor 定数を使わず header から読む。失敗時は entities/errors.Error（Op = corrupt_speech_audio）を返す。
func concatWAV(parts ...[]byte) ([]byte, error) {
	panic("wav concat: contract stub; logic is C")
}
