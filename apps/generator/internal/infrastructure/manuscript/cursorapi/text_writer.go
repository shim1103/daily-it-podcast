package cursorapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

var _ port.TextWriter = (*TextWriter)(nil)

// TextWriter は Cursor Cloud Agents API を使う原稿 Adapter。
type TextWriter struct {
	client *http.Client
	apiKey string
}

// NewTextWriter は Cursor Cloud Agents API 用 TextWriter を組み立てる。
//
// @require client と apiKey は Composition で検証済み。
// @ensure 戻りは port.TextWriter。
func NewTextWriter(client *http.Client, apiKey string) *TextWriter {
	return &TextWriter{client: client, apiKey: apiKey}
}

// Write は port.TextWriter の実装。
//
// @invariant Cloud Agents の create、SSE、retry は未実装。
func (w *TextWriter) Write(_ context.Context, brief string) (string, error) {
	if strings.TrimSpace(brief) == "" {
		return "", infraErr("validate_brief", fmt.Errorf("brief is empty after trim"))
	}
	return "", infraErr("not_implemented", fmt.Errorf("Cursor Cloud Agents API TextWriter is not implemented"))
}
