// Scope: Narrow Integration
// 実物境界: cursorcli.TextWriter → commandlaunch.Launcher(=processenv.Launcher 実物) → exec が起動する実 child process
// Double: PATH 上に置いた fake `agent`（= cursorcli.BinaryName）script のみ。real Cursor CLI は使わない。
// Double 種別: Fake（stdin の brief を受けて JSON envelope を返す簡略実装）。
// secret: dummy 値を processenv.NewSecretEnvLauncherFactory へ直接注入。認証付き実 service は observable にしない。
// @require factory へ dummy secret を直接注入する。child は controllable な fake script。
// @ensure 成功 fake は実起動して stdin で brief を受け取り、非空 fragment を返す。失敗 fake は *cursorcli.Error と空 fragment を返す。
// @invariant dummy secret 実値は戻り error message へ出ない。
package test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorcli"
)

const (
	cursorNarrowDummySecretValue = "cursorcli-narrow-integration-dummy-api-key"
	cursorNarrowBrief            = "今日の IT ニュースの導入を書いて"
)

// installFakeCursorCLI は t.TempDir() へ cursorcli.BinaryName 名の fake script を置き、PATH の先頭へ加える。
// script は起動到達の marker を touch し、受け取った stdin を stdinSink へ反射してから body を実行する。
// marker path と stdinSink path を返す。
func installFakeCursorCLI(t *testing.T, body string) (marker string, stdinSink string) {
	t.Helper()
	dir := t.TempDir()
	marker = filepath.Join(dir, "started")
	stdinSink = filepath.Join(dir, "stdin")
	script := "#!/bin/sh\ntouch '" + marker + "'\ncat > '" + stdinSink + "'\n" + body + "\n"
	program := filepath.Join(dir, cursorcli.BinaryName)
	if err := os.WriteFile(program, []byte(script), 0o755); err != nil {
		t.Fatalf("os.WriteFile() error = %v, want nil", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return marker, stdinSink
}

func newCursorNarrowWriter(t *testing.T) *cursorcli.TextWriter {
	t.Helper()
	factory := processenv.NewSecretEnvLauncherFactory(cursorNarrowDummySecretValue, os.LookupEnv)
	return cursorcli.NewTextWriter(factory)
}

func requireChildStarted(t *testing.T, marker string) {
	t.Helper()
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("os.Stat(%q) error = %v, want fake child started", marker, err)
	}
}

func TestCursorTextWriter_returnsFragment_whenFakeAgentEmitsSuccessEnvelope(t *testing.T) {
	// Given: 成功 envelope を printf する fake agent（stdin は helper が stdinSink へ反射する）
	marker, stdinSink := installFakeCursorCLI(t, `printf '%s' '{"type":"result","subtype":"success","is_error":false,"result":"本日の IT ニュースをお届けします。"}'`)
	writer := newCursorNarrowWriter(t)

	// When: brief を渡して Write する
	got, err := writer.Write(context.Background(), "  "+cursorNarrowBrief+"  ")

	// Then: fake child が実起動し、brief が exec stdin パイプを上って fake へ届き、envelope の result が fragment として返る
	requireChildStarted(t, marker)
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	gotStdin, readErr := os.ReadFile(stdinSink)
	if readErr != nil {
		t.Fatalf("os.ReadFile(%q) error = %v, want nil", stdinSink, readErr)
	}
	if string(gotStdin) != cursorNarrowBrief {
		t.Fatalf("fake agent stdin = %q, want trimmed brief %q", string(gotStdin), cursorNarrowBrief)
	}
	if got != "本日の IT ニュースをお届けします。" {
		t.Fatalf("Write() = %q, want envelope result fragment", got)
	}
}

func TestCursorTextWriter_returnsCursorcliError_whenFakeAgentWritesStderrAndExitsOne(t *testing.T) {
	// Given: stderr へ書いて exit 1 する fake agent
	marker, _ := installFakeCursorCLI(t, `printf '%s' 'fake agent failure detail' 1>&2
exit 1`)
	writer := newCursorNarrowWriter(t)

	// When: brief を渡して Write する
	got, err := writer.Write(context.Background(), "今日の IT ニュースの導入を書いて")

	// Then: fake child は実起動し、*cursorcli.Error と空 fragment が返る
	requireChildStarted(t, marker)
	var cursorErr *cursorcli.Error
	if !errors.As(err, &cursorErr) {
		t.Fatalf("errors.As(err, &cursorErr) = false, want true; type = %T (%v)", err, err)
	}
	if got != "" {
		t.Fatalf("Write() = %q, want empty fragment", got)
	}
}

func TestCursorTextWriter_errorOmitsDummySecretValue_whenFakeAgentExitsNonZero(t *testing.T) {
	// Given: 何もせず exit 3 する fake agent
	marker, _ := installFakeCursorCLI(t, `exit 3`)
	writer := newCursorNarrowWriter(t)

	// When: brief を渡して Write する
	_, err := writer.Write(context.Background(), "今日の IT ニュースの導入を書いて")

	// Then: fake child は実起動し、error は返るが dummy secret 実値を含まない
	requireChildStarted(t, marker)
	if err == nil {
		t.Fatal("Write() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), cursorNarrowDummySecretValue) {
		t.Fatalf("error message %q contains dummy secret value", err.Error())
	}
}
