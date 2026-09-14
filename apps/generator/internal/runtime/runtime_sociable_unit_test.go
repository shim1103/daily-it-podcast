package runtime_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	appruntime "github.com/shim1103/daily-it-podcast/apps/generator/internal/runtime"
)

func TestHTTPClient_hasSharedTimeout(t *testing.T) {
	t.Parallel()

	// Given: （前提なし。工場の既定値だけを見る）
	// When: 短時間用 HTTP Client を作る
	c := appruntime.HTTPClient()

	// Then: 非 nil かつ Timeout は 30s
	if c == nil {
		t.Fatal("HTTPClient: nil")
	}
	if c.Timeout != 30*time.Second {
		t.Fatalf("Timeout = %v, want 30s", c.Timeout)
	}
}

func TestHTTPClientWithoutTimeout_hasNoClientTimeout(t *testing.T) {
	t.Parallel()

	// Given: （前提なし）
	// When: Timeout 無し Client を作る
	c := appruntime.HTTPClientWithoutTimeout()

	// Then: 非 nil かつ Client.Timeout は 0
	if c == nil {
		t.Fatal("HTTPClientWithoutTimeout: nil")
	}
	if c.Timeout != 0 {
		t.Fatalf("Timeout = %v, want 0", c.Timeout)
	}
}

func TestLookupEnv_matchesOsLookupEnv(t *testing.T) {
	// Given: probe key を環境へ置き、LookupEnv を得る（t.Setenv と Parallel は両立しない）
	t.Setenv("GENERATOR_RUNTIME_LOOKUP_PROBE", "1")
	lookup := appruntime.LookupEnv()

	// When: probe key を読む
	got, ok := lookup("GENERATOR_RUNTIME_LOOKUP_PROBE")

	// Then: 置いた値が返る
	if !ok || got != "1" {
		t.Fatalf("LookupEnv = (%q, %v), want (\"1\", true)", got, ok)
	}
}

func TestLookPath_findsExecutableOnPATH(t *testing.T) {
	t.Parallel()

	// Given: PATH 上にある実行 file 名
	name := "go"

	// When: LookPath で解決する
	path, err := appruntime.LookPath()(name)

	// Then: 非空 path と nil error
	if err != nil {
		t.Fatalf("LookPath(%s): %v", name, err)
	}
	if path == "" {
		t.Fatalf("LookPath(%s): empty path", name)
	}
}

func TestCommandRun_returnsStdout_whenProcessSucceeds(t *testing.T) {
	t.Parallel()

	// Given: CommandRun と stdin 本文
	run := appruntime.CommandRun()
	stdin := []byte("hello-runtime")

	// When: cat に stdin を渡して起動する
	out, err := run(context.Background(), "cat", nil, stdin)

	// Then: stdout が stdin と同じで error なし
	if err != nil {
		t.Fatalf("CommandRun: %v", err)
	}
	if string(out) != string(stdin) {
		t.Fatalf("stdout = %q, want %q", out, stdin)
	}
}

func TestCommandRun_returnsError_whenProcessFails(t *testing.T) {
	t.Parallel()

	// Given: 常に非 0 で終わる command
	run := appruntime.CommandRun()

	// When: false を起動する
	_, err := run(context.Background(), "false", nil, nil)

	// Then: error が返る
	if err == nil {
		t.Fatal("CommandRun(false): err = nil, want error")
	}
}

func TestCommandRun_includesStderr_whenProcessFailsWithStderr(t *testing.T) {
	t.Parallel()

	// Given: 非 0 exit かつ stderr に文言を出す shell
	run := appruntime.CommandRun()
	args := []string{"-c", "echo boom >&2; exit 1"}

	// When: その shell を起動する
	_, err := run(context.Background(), "sh", args, nil)

	// Then: error に stderr 文言が含まれる
	if err == nil {
		t.Fatal("CommandRun: err = nil, want error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("CommandRun err = %v, want stderr snippet", err)
	}
}

func TestDisplayLocation_resolvesConfiguredZone(t *testing.T) {
	t.Parallel()

	// Given: 定数 DisplayTimeZone が指す Location
	want, err := time.LoadLocation(constants.DisplayTimeZone)
	if err != nil {
		t.Fatalf("LoadLocation(%s): %v", constants.DisplayTimeZone, err)
	}

	// When: DisplayLocation を呼ぶ
	loc := appruntime.DisplayLocation()

	// Then: 同じ zone の非 nil Location
	if loc == nil {
		t.Fatal("DisplayLocation: nil")
	}
	if loc.String() != want.String() {
		t.Fatalf("DisplayLocation = %q, want %q", loc, want)
	}
}
