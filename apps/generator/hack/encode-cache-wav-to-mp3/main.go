// Package main は .cache の wav を本番 Encoder へ渡して mp3 にする使い捨て入口である。
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/audio/ffmpeg"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/runtime"
)

// main は引数 prod|test の .cache/<target> 配下 wav を EncodeWAVToMP3 へ渡す。
//
// @require 引数は prod または test。ffmpeg が PATH にある（Encoder 経由）。
// @ensure 各非空 wav の同 stem mp3 を書く。Drive / secret を読まない。
func main() {
	if len(os.Args) != 2 || (os.Args[1] != "prod" && os.Args[1] != "test") {
		fmt.Fprintf(os.Stderr, "usage: encode-cache-wav-to-mp3 prod|test\n")
		os.Exit(1)
	}
	target := os.Args[1]
	root, err := repoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode-cache: %v\n", err)
		os.Exit(1)
	}
	cacheRoot := os.Getenv("CACHE_ROOT")
	if cacheRoot == "" {
		cacheRoot = filepath.Join(root, ".cache")
	}
	dir := filepath.Join(cacheRoot, target)
	enc := ffmpeg.NewEncoder(runtime.LookPath(), runtime.CommandRun())
	if err := EncodeCacheDir(context.Background(), dir, enc.EncodeWAVToMP3); err != nil {
		fmt.Fprintf(os.Stderr, "encode-cache: %v\n", err)
		os.Exit(1)
	}
}

// EncodeCacheDir は dir 直下の .wav を encode へ渡し、同 stem の .mp3 を書く。
func EncodeCacheDir(ctx context.Context, dir string, encode func(context.Context, []byte) ([]byte, error)) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("encode-cache: dir 無し（skip）: %s\n", dir)
			return nil
		}
		return err
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".wav") {
			continue
		}
		wavPath := filepath.Join(dir, e.Name())
		wav, err := os.ReadFile(wavPath)
		if err != nil {
			return err
		}
		if len(wav) == 0 {
			fmt.Fprintf(os.Stderr, "encode-cache: 空 wav を skip: %s\n", wavPath)
			continue
		}
		mp3, err := encode(ctx, wav)
		if err != nil {
			return fmt.Errorf("%s: %w", wavPath, err)
		}
		if len(mp3) == 0 {
			return fmt.Errorf("%s: encode が空 mp3 を返した", wavPath)
		}
		stem := strings.TrimSuffix(e.Name(), ".wav")
		mp3Path := filepath.Join(dir, stem+".mp3")
		fmt.Printf("encode-cache: %s -> %s\n", wavPath, mp3Path)
		if err := os.WriteFile(mp3Path, mp3, 0o644); err != nil {
			return err
		}
		count++
	}
	if count == 0 {
		fmt.Printf("encode-cache: wav 0 件（skip）: %s\n", dir)
		return nil
	}
	fmt.Printf("encode-cache: done count=%d dir=%s\n", count, dir)
	return nil
}

func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repo root が見つからない")
		}
		dir = parent
	}
}
