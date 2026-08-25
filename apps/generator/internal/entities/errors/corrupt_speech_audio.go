package errors

import "fmt"

// CorruptSpeechAudio は WAV が契約（RIFF/L16）を満たせない。
type CorruptSpeechAudio struct {
	Err error
}

func (e *CorruptSpeechAudio) Error() string {
	if e == nil || e.Err == nil {
		return "speech audio is corrupt"
	}
	return fmt.Sprintf("speech audio is corrupt: %v", e.Err)
}

func (e *CorruptSpeechAudio) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
