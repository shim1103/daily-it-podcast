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
// @ensure composition.NewProduceEpisodeFromEnv() の load error は kind 付き構造化 stderr へ出し process exit 非0。
// @ensure ProduceEpisode.Run が nil なら process exit 0、non-nil error なら kind 付き構造化 stderr へ出し process exit 非0。
// @invariant internal/infrastructure と application/port を import しない。秘密・env を読まない。生成手順を持たない。
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	produceEpisode, err := composition.NewProduceEpisodeFromEnv()
	if err != nil {
		writeExternalError(os.Stderr, err)
		os.Exit(1)
	}

	code := run(ctx, time.Now(), os.Stderr, produceEpisode.Run)
	if code != 0 {
		os.Exit(code)
	}
}

// run は ProduceEpisode.Run の結果を process 終了へ写す。
func run(ctx context.Context, now time.Time, stderr io.Writer, produce func(context.Context, time.Time) error) int {
	if err := produce(ctx, now); err != nil {
		writeExternalError(stderr, err)
		return 1
	}
	return 0
}

func writeExternalError(stderr io.Writer, err error) {
	// why: 既に失敗経路。stderr 書込失敗をさらに上位へ持ち出す先が無い。
	_, _ = io.WriteString(stderr, delivery.Format(err))
}
