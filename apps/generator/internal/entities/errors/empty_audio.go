package errors

// EmptyAudio は音声 bytes が空である。
type EmptyAudio struct{}

func (e *EmptyAudio) Error() string {
	return "speech audio content is empty"
}
