package models

// SpeechAudio は SpeechSynthesizer が返す音声 file の中身。
// 形式は Drive 配置契約の拡張子に従う WAV バイト列（TTS 決定に準拠）。path は持たない。
// RIFF 解析・結合手順は Application 層の unexported helper が所有する。
type SpeechAudio struct {
	Content []byte
}
