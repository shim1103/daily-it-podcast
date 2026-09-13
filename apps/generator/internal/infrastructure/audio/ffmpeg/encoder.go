// Package ffmpeg は WAV→MP3 を OS 上の ffmpeg subprocess で行う Driven Adapter である。
package ffmpeg

import (
	"context"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

var _ port.WAVToMP3Encoder = (*Encoder)(nil)

// LookPath は PATH 上の実行 file を探す契約である。
// production では Composition が internal/runtime.LookPath を渡す（config.LookupEnv と同型）。
type LookPath func(file string) (string, error)

// Run は name を起動し stdin を渡し、成功時の stdout を返す契約である。
// production では Composition が internal/runtime.CommandRun を渡す。Adapter は os/exec を import しない。
type Run func(ctx context.Context, name string, args []string, stdin []byte) (stdout []byte, err error)

// Encoder は ffmpeg subprocess で WAV を MP3 へ変換する。
// 本 stub の EncodeWAVToMP3 は零値返却のみ。ffmpeg 本実装は C（Issue）側。
type Encoder struct {
	lookPath LookPath
	run      Run
}

// NewEncoder は LookPath / Run を注入した Encoder を返す。
//
// @require lookPath と run は Composition（または test）が渡す。production は runtime.LookPath / runtime.CommandRun。
// @ensure 戻りは非 nil の *Encoder（port.WAVToMP3Encoder）。
func NewEncoder(lookPath LookPath, run Run) *Encoder {
	return &Encoder{lookPath: lookPath, run: run}
}

// EncodeWAVToMP3 は WAV バイト列を MP3 へ変換する。
//
// @require wav は呼び出し側が非空を保証する想定（stub は検査しない）。
// @ensure 現 stub は常に (nil, nil) を返す。lookPath / run は C 本実装で使う。
func (e *Encoder) EncodeWAVToMP3(ctx context.Context, wav []byte) ([]byte, error) {
	_ = ctx
	_ = wav
	_ = e.lookPath
	_ = e.run
	return nil, nil
}
