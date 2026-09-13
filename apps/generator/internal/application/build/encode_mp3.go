package build

// EncodeWAVToMP3 は WAV バイト列を MP3 へ変換する。
// ConcatWAV 直後・EpisodeWriter.Write 直前で呼ぶ想定。ProduceEpisode への結線は未実施。
// ffmpeg 実装は C（Issue）側。本 stub は契約凍結用の零値返却のみ。
//
// @require wav は呼び出し側が非空を保証する想定（stub は検査しない）。
// @ensure 現 stub は常に (nil, nil) を返す。
func EncodeWAVToMP3(wav []byte) ([]byte, error) {
	_ = wav
	return nil, nil
}
