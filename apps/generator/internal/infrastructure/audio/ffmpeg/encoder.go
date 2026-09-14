// Package ffmpeg は WAV→MP3 を OS 上の ffmpeg subprocess で行う Driven Adapter である。
package ffmpeg

import (
	"context"
	"fmt"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

var _ port.WAVToMP3Encoder = (*Encoder)(nil)

const ffmpegBin = "ffmpeg"

// encodeArgs は stdin WAV → stdout MP3（libmp3lame 128k）の固定 argv。
var encodeArgs = []string{
	"-hide_banner", "-loglevel", "error",
	"-i", "pipe:0",
	"-f", "mp3", "-codec:a", "libmp3lame", "-b:a", "128k",
	"pipe:1",
}

// LookPath は PATH 上の実行 file を探す契約である。
// production では Composition が internal/runtime.LookPath を渡す（config.LookupEnv と同型）。
type LookPath func(file string) (string, error)

// Run は name を起動し stdin を渡し、成功時の stdout を返す契約である。
// production では Composition が internal/runtime.CommandRun を渡す。Adapter は os/exec を import しない。
type Run func(ctx context.Context, name string, args []string, stdin []byte) (stdout []byte, err error)

// Encoder は ffmpeg subprocess で WAV を MP3 へ変換する。
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
// @require wav は呼び出し側が非空を保証する想定。lookPath / run は NewEncoder で注入済み。
// @ensure 成功時は非空 MP3 bytes。ffmpeg 不在・非 0 exit・空出力は *adaptererror.Error。retry しない。
// @invariant os/exec を import しない。argv / bitrate は本 Adapter に閉じる。
func (e *Encoder) EncodeWAVToMP3(ctx context.Context, wav []byte) ([]byte, error) {
	path, err := e.lookPath(ffmpegBin)
	if err != nil {
		return nil, infraErr("look_path", err)
	}
	stdout, err := e.run(ctx, path, encodeArgs, wav)
	if err != nil {
		return nil, infraErr("run", err)
	}
	if len(stdout) == 0 {
		return nil, infraErr("empty_output", fmt.Errorf("ffmpeg stdout is empty"))
	}
	return stdout, nil
}
