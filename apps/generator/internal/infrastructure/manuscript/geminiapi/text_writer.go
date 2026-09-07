// Package geminiapi は Gemini generateContent を使う原稿 TextWriter Adapter を提供する。
// Cursor Cloud Agents の利用枠喪失時に manuscript UseCase が secondary として使う。
package geminiapi

import (
	"context"
	"net/http"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

var _ port.TextWriter = (*TextWriter)(nil)

// TextWriter は Gemini generateContent を叩く原稿 Adapter。
type TextWriter struct {
	client         *http.Client
	apiKey         string
	backoffSleepFn func(context.Context, time.Duration) // why: test の並列実行と共存するため package global に置かない
}

// NewTextWriter は Gemini generateContent 用 TextWriter を組み立てる。
//
// @require apiKey は Composition で検証済み。
// @ensure 戻りは port.TextWriter。apiKey は APIKeyHeader にだけ使う。
// @ensure client == nil のとき Write は geminiErr("build_request") を返す。
func NewTextWriter(client *http.Client, apiKey string) *TextWriter {
	return newTextWriter(client, apiKey, ctxSleep)
}

// newTextWriter は backoff の sleep 関数を差し込める内部 constructor。
// why: NewTextWriter は本番の ctxSleep を固定し、test は待ちを観測する fake を渡す（cursorapi と同型）。
func newTextWriter(client *http.Client, apiKey string, backoffSleepFn func(context.Context, time.Duration)) *TextWriter {
	if backoffSleepFn == nil {
		backoffSleepFn = ctxSleep
	}
	return &TextWriter{client: client, apiKey: apiKey, backoffSleepFn: backoffSleepFn}
}

// ctxSleep は ctx が先に切れたらそちらを優先して待ちを中断する。
func ctxSleep(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// Write は brief から generateContent 1 回で原稿断片を得る。
//
// @require brief は trim 後に非空。
// @ensure 成功時は非空 text 断片を返す。失敗時は *geminiapi.Error、断片は空。
// @invariant generateContent は idempotent（同 body は同じ生成試行・副作用なし）。client.Do error / 5xx を 1 回、429 を MaxAttempts まで backoff で再試行する。401 / 403 / その他 4xx、finishReason が STOP 以外、空 text、parse 失敗は再試行しない。secret 実値を error へ出さない。model は ModelID 固定。
//
// TODO(generator-text-writer-fallback): 本体は C で実装する。現状は stub。
func (w *TextWriter) Write(ctx context.Context, brief string) (string, error) {
	return "", nil
}
