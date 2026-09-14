package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/audio/ffmpeg"
	appruntime "github.com/shim1103/daily-it-podcast/apps/generator/internal/runtime"
)

// newFFmpegWAVToMP3Encoder は ffmpeg Adapter を組み立てる。
//
// @ensure 戻りは port.WAVToMP3Encoder。LookPath / Run は internal/runtime を inject する。
// @invariant Application へ os/exec を渡さない。結線だけ行う。
func newFFmpegWAVToMP3Encoder() port.WAVToMP3Encoder {
	return ffmpeg.NewEncoder(appruntime.LookPath(), appruntime.CommandRun())
}
