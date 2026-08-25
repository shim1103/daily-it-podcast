package errors_test

import (
	"errors"
	"fmt"
	"testing"

	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
)

func TestCorruptSpeechAudio_ErrorAndUnwrap(t *testing.T) {
	t.Parallel()
	cause := fmt.Errorf("bad riff")
	err := &domainerrors.CorruptSpeechAudio{Err: cause}
	if got := err.Error(); got != "speech audio is corrupt: bad riff" {
		t.Fatalf("Error() = %q", got)
	}
	if !errors.Is(err, cause) {
		t.Fatal("Unwrap/Is failed")
	}
	if (&domainerrors.CorruptSpeechAudio{}).Error() != "speech audio is corrupt" {
		t.Fatal("nil cause message")
	}
}

func TestInvalidManuscriptDraft_ErrorAndUnwrap(t *testing.T) {
	t.Parallel()
	cause := fmt.Errorf("missing topics")
	err := &domainerrors.InvalidManuscriptDraft{Err: cause}
	if got := err.Error(); got != "manuscript draft is invalid: missing topics" {
		t.Fatalf("Error() = %q", got)
	}
	if !errors.Is(err, cause) {
		t.Fatal("Unwrap/Is failed")
	}
	if (&domainerrors.InvalidManuscriptDraft{}).Error() != "manuscript draft is invalid" {
		t.Fatal("nil cause message")
	}
}
