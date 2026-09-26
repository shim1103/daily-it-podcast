package delivery

import (
	"fmt"
	"io"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

// LogWriter は generator CLI の唯一の log 出力面である。
// why: 生 log（log.Print / fmt.Fprintf(os.Stderr, ...)）を各所へ散らさず、
//
//	"generator: key=value" 1 行 format を1点で保証する（architecture/logging-policy）。
type LogWriter struct {
	w io.Writer
}

func NewLogWriter(w io.Writer) *LogWriter { return &LogWriter{w: w} }

// Field は log 1 行へ添える key=value ペアである。
// why: value は語彙固定（識別子・状態名など空白と "=" を含まない語）を呼び出し側の責務とする。
//
//	自由文を載せるなら quote 責務も呼び出し側が負う。oneLine は改行だけを潰す。
type Field struct {
	Key   string
	Value string
}

// Event は "generator: category=<category> event=<name> [k=v ...]" を1行だけ書く。
// why: message/理由に secret を渡さないのは呼び出し側の責務（error-handling/logging §3）。
func (l *LogWriter) Event(category, name string, fields ...Field) {
	if l == nil || l.w == nil {
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "generator: category=%s event=%s", oneLine(category), oneLine(name))
	for _, f := range fields {
		fmt.Fprintf(&b, " %s=%s", oneLine(f.Key), oneLine(f.Value))
	}
	b.WriteByte('\n')
	_, _ = io.WriteString(l.w, b.String())
}

var (
	_ port.ProgressReporter = (*LogWriter)(nil)
	_ port.FallbackReporter = (*LogWriter)(nil)
)

// Start は段階の開始を "category=progress ... phase=start" 行で書く。
func (l *LogWriter) Start(step string) {
	l.Event("progress", step, Field{Key: "phase", Value: "start"})
}

// Done は段階の完了を "category=progress ... phase=done [detail=...]" 行で書く。
func (l *LogWriter) Done(step, detail string) {
	fields := []Field{{Key: "phase", Value: "done"}}
	if detail != "" {
		fields = append(fields, Field{Key: "detail", Value: detail})
	}
	l.Event("progress", step, fields...)
}

// Fallback は取得元切替などの観測イベントを "category=fallback" 行で書く。
func (l *LogWriter) Fallback(event string) {
	l.Event("fallback", event)
}
