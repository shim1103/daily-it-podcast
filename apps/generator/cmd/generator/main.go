package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/composition"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
)

// main は generator CLI の Driving Adapter 入口である。
//
// @require process が Interrupt / SIGTERM を届けられる。
// @ensure load または Run の失敗は構造化 stderr へ出し exit 非0。成功は exit 0。
// @invariant infrastructure / application/port を import しない。秘密・env・生成手順を持たない。
func main() {
	// why: os.Exit は defer を実行しない。後始末付き本体を return させてから一度だけ Exit する。
	os.Exit(runCLI())
}

func runCLI() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logw := delivery.NewLogWriter(os.Stderr)

	produceEpisode, err := composition.NewProduceEpisodeFromEnv(logw)
	if err != nil {
		writeExternalError(os.Stderr, err)
		return 1
	}

	return run(ctx, time.Now(), os.Stderr, produceEpisode.Run)
}

func run(ctx context.Context, now time.Time, stderr io.Writer, produce func(context.Context, time.Time) (string, error)) int {
	if _, err := produce(ctx, now); err != nil {
		writeExternalError(stderr, err)
		return 1
	}
	return 0
}

func writeExternalError(stderr io.Writer, err error) {
	// why: 既に失敗経路。stderr 書込失敗をさらに上位へ持ち出す先が無い。
	_, _ = io.WriteString(stderr, delivery.Format(err))
}
