package models

// SpeechAudio は SpeechSynthesizer が返す音声 file の中身。
// TTS は WAV セグメントを返す。Drive 保存形式は drive-layout（mp3）に従う。path は持たない。
// RIFF 解析・結合手順は Application 層の unexported helper が所有する。
type SpeechAudio struct {
	Content []byte
}
