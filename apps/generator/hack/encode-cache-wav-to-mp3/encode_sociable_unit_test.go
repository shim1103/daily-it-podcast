package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestEncodeCacheDir_writesMp3BesideWav_whenEncodeSucceeds(t *testing.T) {
	// Given: cache dir に非空 wav。encode は固定 mp3 bytes を返す
	dir := t.TempDir()
	wavPath := filepath.Join(dir, "ep-1.wav")
	if err := os.WriteFile(wavPath, []byte("wav-bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile wav: %v", err)
	}
	wantMP3 := []byte("mp3-bytes")
	encode := func(context.Context, []byte) ([]byte, error) {
		return wantMP3, nil
	}

	// When: EncodeCacheDir を呼ぶ
	err := EncodeCacheDir(context.Background(), dir, encode)

	// Then: err なし。同 stem の mp3 が wantMP3。wav は残る
	if err != nil {
		t.Fatalf("EncodeCacheDir: err = %v, want nil", err)
	}
	got, readErr := os.ReadFile(filepath.Join(dir, "ep-1.mp3"))
	if readErr != nil {
		t.Fatalf("ReadFile mp3: %v", readErr)
	}
	if !bytes.Equal(got, wantMP3) {
		t.Fatalf("mp3 = %q, want %q", got, wantMP3)
	}
	if _, statErr := os.Stat(wavPath); statErr != nil {
		t.Fatalf("wav が消えている: %v", statErr)
	}
}

func TestEncodeCacheDir_returnsNil_whenNoWav(t *testing.T) {
	// Given: wav が無い空 dir。encode が呼ばれたら Fail
	dir := t.TempDir()
	encode := func(context.Context, []byte) ([]byte, error) {
		t.Fatal("encode は呼ばれない想定")
		return nil, nil
	}

	// When: EncodeCacheDir を呼ぶ
	err := EncodeCacheDir(context.Background(), dir, encode)

	// Then: err なし
	if err != nil {
		t.Fatalf("EncodeCacheDir: err = %v, want nil", err)
	}
}

func TestEncodeCacheDir_returnsError_whenEncodeReturnsEmpty(t *testing.T) {
	// Given: 非空 wav。encode は空 bytes を返す
	dir := t.TempDir()
	wavPath := filepath.Join(dir, "ep-1.wav")
	if err := os.WriteFile(wavPath, []byte("wav-bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile wav: %v", err)
	}
	encode := func(context.Context, []byte) ([]byte, error) {
		return []byte{}, nil
	}

	// When: EncodeCacheDir を呼ぶ
	err := EncodeCacheDir(context.Background(), dir, encode)

	// Then: 空 mp3 を示す error。mp3 file は作られない
	if err == nil {
		t.Fatal("EncodeCacheDir: err = nil, want 空 mp3 error")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "ep-1.mp3")); !os.IsNotExist(statErr) {
		t.Fatalf("mp3 Stat = %v, want not exist", statErr)
	}
}
