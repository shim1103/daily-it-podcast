package models

// SpeechAudio は SpeechSynthesizer が返す音声 file の中身。
// 形式は Drive 配置契約の mp3。path は持たない。
type SpeechAudio struct {
	Content []byte
}
