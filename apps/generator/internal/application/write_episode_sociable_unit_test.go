package application_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

const validManuscript = `{
	"episodeId": "ep-1",
	"date": "2026-08-20",
	"title": "今日の IT",
	"durationSec": 12,
	"body": {
		"opening": "こんにちは",
		"topics": [{
			"title": "話題",
			"preface": "前置き",
			"detail": "詳細",
			"startSec": 0
		}],
		"closing": "また明日"
	}
}`

type fakeEpisodeWriter struct {
	calls      int
	episodeID  string
	manuscript []byte
	audio      models.SpeechAudio
	err        error
}

func (f *fakeEpisodeWriter) Write(_ context.Context, episodeID string, manuscript []byte, audio models.SpeechAudio) error {
	f.calls++
	f.episodeID = episodeID
	f.manuscript = manuscript
	f.audio = audio
	return f.err
}

func TestWriteEpisode_writesValidatedEpisode_whenAllInputsAreValid(t *testing.T) {
	// Given: schema に適合する原稿と非空 WAV
	fake := &fakeEpisodeWriter{}
	uc := application.NewWriteEpisode(fake)
	audio := models.SpeechAudio{Content: []byte("RIFFWAV")}

	// When: Write を呼ぶ
	err := uc.Run(context.Background(), "ep-1", []byte(validManuscript), audio)

	// Then: validation 後に fake が 1 回呼ばれ、入力がそのまま渡る
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if fake.calls != 1 {
		t.Fatalf("Write calls = %d, want 1", fake.calls)
	}
	if fake.episodeID != "ep-1" {
		t.Fatalf("episodeID = %q, want %q", fake.episodeID, "ep-1")
	}
	if string(fake.manuscript) != validManuscript {
		t.Fatalf("manuscript was changed")
	}
	if string(fake.audio.Content) != string(audio.Content) {
		t.Fatalf("audio was changed")
	}
}

func TestWriteEpisode_returnsSchemaErrorWithoutWriting_whenManuscriptIsInvalid(t *testing.T) {
	// Given: schema に適合しない原稿
	fake := &fakeEpisodeWriter{}
	uc := application.NewWriteEpisode(fake)

	// When: 必須 field がない原稿で Write を呼ぶ
	err := uc.Run(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: schema Domain Error。fake は呼ばれない
	if err == nil {
		t.Fatal("expected error")
	}
	var schemaErr *domainerrors.InvalidManuscript
	if !errors.As(err, &schemaErr) {
		t.Fatalf("error type %T (%v), want *errors.InvalidManuscript", err, err)
	}
	if fake.calls != 0 {
		t.Fatalf("Write calls = %d, want 0", fake.calls)
	}
}

func TestWriteEpisode_returnsEpisodeIDMismatchWithoutWriting_whenManuscriptStemDiffers(t *testing.T) {
	// Given: 原稿内 episodeId が stem と異なる
	fake := &fakeEpisodeWriter{}
	uc := application.NewWriteEpisode(fake)

	// When: 異なる episodeID で Write を呼ぶ
	err := uc.Run(context.Background(), "ep-2", []byte(validManuscript), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: stem 不一致 Domain Error。fake は呼ばれない
	if err == nil {
		t.Fatal("expected error")
	}
	var mismatch *domainerrors.EpisodeIDMismatch
	if !errors.As(err, &mismatch) {
		t.Fatalf("error type %T (%v), want *errors.EpisodeIDMismatch", err, err)
	}
	if fake.calls != 0 {
		t.Fatalf("Write calls = %d, want 0", fake.calls)
	}
}

func TestWriteEpisode_returnsEmptyEpisodeIDWithoutWriting_whenEpisodeIDIsEmpty(t *testing.T) {
	// Given: episodeID が空
	fake := &fakeEpisodeWriter{}
	uc := application.NewWriteEpisode(fake)

	// When: 空 episodeID で Write を呼ぶ
	err := uc.Run(context.Background(), "", []byte(validManuscript), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: Domain Error。fake は呼ばれない
	if err == nil {
		t.Fatal("expected error")
	}
	var emptyID *domainerrors.EmptyEpisodeID
	if !errors.As(err, &emptyID) {
		t.Fatalf("error type %T (%v), want *errors.EmptyEpisodeID", err, err)
	}
	if fake.calls != 0 {
		t.Fatalf("Write calls = %d, want 0", fake.calls)
	}
}

func TestWriteEpisode_returnsEmptyAudioWithoutWriting_whenAudioIsEmpty(t *testing.T) {
	// Given: WAV Content が空
	fake := &fakeEpisodeWriter{}
	uc := application.NewWriteEpisode(fake)

	// When: 空 WAV で Write を呼ぶ
	err := uc.Run(context.Background(), "ep-1", []byte(validManuscript), models.SpeechAudio{})

	// Then: Domain Error。fake は呼ばれない
	if err == nil {
		t.Fatal("expected error")
	}
	var emptyAudio *domainerrors.EmptyAudio
	if !errors.As(err, &emptyAudio) {
		t.Fatalf("error type %T (%v), want *errors.EmptyAudio", err, err)
	}
	if fake.calls != 0 {
		t.Fatalf("Write calls = %d, want 0", fake.calls)
	}
}

func TestWriteEpisode_returnsSchemaErrorWithoutWriting_whenManuscriptIsMalformedJSON(t *testing.T) {
	// Given: JSON として壊れた原稿
	fake := &fakeEpisodeWriter{}
	uc := application.NewWriteEpisode(fake)

	// When: 壊れた JSON で Write を呼ぶ
	err := uc.Run(context.Background(), "ep-1", []byte(`{"episodeId":`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: schema Domain Error。fake は呼ばれない
	var schemaErr *domainerrors.InvalidManuscript
	if !errors.As(err, &schemaErr) {
		t.Fatalf("error type %T (%v), want *errors.InvalidManuscript", err, err)
	}
	if fake.calls != 0 {
		t.Fatalf("Write calls = %d, want 0", fake.calls)
	}
}

func TestWriteEpisode_returnsSchemaErrorWithoutWriting_whenManuscriptHasTrailingJSON(t *testing.T) {
	// Given: JSON の後ろに別の値がある原稿
	fake := &fakeEpisodeWriter{}
	uc := application.NewWriteEpisode(fake)

	// When: trailing JSON 付きで Write を呼ぶ
	err := uc.Run(context.Background(), "ep-1", []byte(validManuscript+` {}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: schema Domain Error。fake は呼ばれない
	var schemaErr *domainerrors.InvalidManuscript
	if !errors.As(err, &schemaErr) {
		t.Fatalf("error type %T (%v), want *errors.InvalidManuscript", err, err)
	}
	if fake.calls != 0 {
		t.Fatalf("Write calls = %d, want 0", fake.calls)
	}
}

func TestWriteEpisode_returnsWriterError_whenWriterFails(t *testing.T) {
	// Given: writer が error を返す
	boom := fmt.Errorf("write failed")
	fake := &fakeEpisodeWriter{err: boom}
	uc := application.NewWriteEpisode(fake)

	// When: valid episode を Write する
	err := uc.Run(context.Background(), "ep-1", []byte(validManuscript), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: writer の error をそのまま返し、writer は 1 回呼ばれる
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if fake.calls != 1 {
		t.Fatalf("Write calls = %d, want 1", fake.calls)
	}
}
