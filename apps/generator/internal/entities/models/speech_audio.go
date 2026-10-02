package models

// SpeechAudio は SpeechSynthesizer が返す音声 file の中身。
// TTS は WAV セグメントを返す。完成保存形式は episode-layout（mp3）に従う。path は持たない。
// RIFF 解析・結合手順は Application 層の unexported helper が所有する。
//
// @invariant DurationSec は Content（WAV）から算出した再生尺（秒）。Content が空でない限り >= 0。
type SpeechAudio struct {
	Content     []byte
	DurationSec float64
}
