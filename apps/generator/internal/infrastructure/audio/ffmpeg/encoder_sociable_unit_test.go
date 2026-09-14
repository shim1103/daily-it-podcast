package ffmpeg_test

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/audio/ffmpeg"
)

func TestEncodeWAVToMP3_returnsNonEmptyMP3_whenLookPathAndRunSucceed(t *testing.T) {
	t.Parallel()

	// Given: LookPath 成功・Run が非空 stdout を返す
	wantMP3 := []byte("fake-mp3-bytes")
	var gotName string
	var gotArgs []string
	var gotStdin []byte
	enc := ffmpeg.NewEncoder(
		func(file string) (string, error) {
			if file != "ffmpeg" {
				t.Fatalf("LookPath file = %q, want ffmpeg", file)
			}
			return "/usr/bin/ffmpeg", nil
		},
		func(_ context.Context, name string, args []string, stdin []byte) ([]byte, error) {
			gotName = name
			gotArgs = append([]string(nil), args...)
			gotStdin = append([]byte(nil), stdin...)
			return wantMP3, nil
		},
	)
	wav := []byte("concat-wav-bytes")

	// When: EncodeWAVToMP3 を呼ぶ
	got, err := enc.EncodeWAVToMP3(context.Background(), wav)

	// Then: 非空 MP3。Run へは解決 path・標準 argv・WAV stdin が渡る
	if err != nil {
		t.Fatalf("EncodeWAVToMP3: err = %v, want nil", err)
	}
	if !bytes.Equal(got, wantMP3) {
		t.Fatalf("EncodeWAVToMP3: got = %q, want %q", got, wantMP3)
	}
	if gotName != "/usr/bin/ffmpeg" {
		t.Fatalf("Run name = %q, want /usr/bin/ffmpeg", gotName)
	}
	if !bytes.Equal(gotStdin, wav) {
		t.Fatalf("Run stdin = %q, want %q", gotStdin, wav)
	}
	wantArgs := []string{
		"-hide_banner", "-loglevel", "error",
		"-i", "pipe:0",
		"-f", "mp3", "-codec:a", "libmp3lame", "-b:a", "128k",
		"pipe:1",
	}
	if !slices.Equal(gotArgs, wantArgs) {
		t.Fatalf("Run args = %q, want %q", gotArgs, wantArgs)
	}
}

func TestEncodeWAVToMP3_returnsInfrastructureError_whenLookPathFails(t *testing.T) {
	t.Parallel()

	// Given: LookPath が error（ffmpeg 不在）
	boom := errors.New("executable file not found in $PATH")
	enc := ffmpeg.NewEncoder(
		func(string) (string, error) { return "", boom },
		func(context.Context, string, []string, []byte) ([]byte, error) {
			t.Fatal("Run は呼ばれない想定")
			return nil, nil
		},
	)

	// When: EncodeWAVToMP3 を呼ぶ
	got, err := enc.EncodeWAVToMP3(context.Background(), []byte("wav"))

	// Then: Infrastructure Error。MP3 は空
	if got != nil {
		t.Fatalf("got = %v, want nil", got)
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want wrap of %v", err, boom)
	}
}

func TestEncodeWAVToMP3_returnsInfrastructureError_whenRunFails(t *testing.T) {
	t.Parallel()

	// Given: LookPath 成功・Run が非 0 exit 相当の error
	boom := errors.New("exit status 1")
	enc := ffmpeg.NewEncoder(
		func(string) (string, error) { return "/usr/bin/ffmpeg", nil },
		func(context.Context, string, []string, []byte) ([]byte, error) {
			return nil, boom
		},
	)

	// When: EncodeWAVToMP3 を呼ぶ
	got, err := enc.EncodeWAVToMP3(context.Background(), []byte("wav"))

	// Then: Infrastructure Error。MP3 は空
	if got != nil {
		t.Fatalf("got = %v, want nil", got)
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want wrap of %v", err, boom)
	}
}

func TestEncodeWAVToMP3_returnsInfrastructureError_whenStdoutEmpty(t *testing.T) {
	t.Parallel()

	// Given: Run は成功するが stdout が空
	enc := ffmpeg.NewEncoder(
		func(string) (string, error) { return "/usr/bin/ffmpeg", nil },
		func(context.Context, string, []string, []byte) ([]byte, error) {
			return []byte{}, nil
		},
	)

	// When: EncodeWAVToMP3 を呼ぶ
	got, err := enc.EncodeWAVToMP3(context.Background(), []byte("wav"))

	// Then: Infrastructure Error。MP3 は空
	if got != nil {
		t.Fatalf("got = %v, want nil", got)
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
}
