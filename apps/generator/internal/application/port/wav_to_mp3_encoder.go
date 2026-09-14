package port

import "context"

// WAVToMP3Encoder は完成 WAV バイト列を配信用 MP3 へ変換する。
// ffmpeg の起動・PATH・exit code は Infrastructure に閉じる。
//
// @require wav は呼び出し側が非空を保証する想定（結合済み完成 WAV）。
// @ensure 成功時は非空 MP3 bytes。ffmpeg 不在・非 0 exit・空出力は error。
// @invariant vendor / OS command / path / bitrate を露出しない。method は EncodeWAVToMP3 のみ。
// @invariant TTS Port とは別境界。セグメント合成・尺計算は行わない。
type WAVToMP3Encoder interface {
	EncodeWAVToMP3(ctx context.Context, wav []byte) ([]byte, error)
}
