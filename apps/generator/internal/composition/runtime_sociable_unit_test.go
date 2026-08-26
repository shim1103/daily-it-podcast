package composition

import (
	"path/filepath"
	"testing"

	commandlaunchagentsecrets "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/agentsecrets"
)

func TestCursorCommandProjectDir_resolvesCursorDedicatedProject_whenHomeIsSet(t *testing.T) {
	// Given: HOME
	home := t.TempDir()
	t.Setenv("HOME", home)

	// When: Composition 所有の Cursor project dir を取る
	got := CursorCommandProjectDir()

	// Then: AgentSecrets の Cursor 専用 project 規約に従う絶対 path
	want := commandlaunchagentsecrets.DefaultProjectDir(commandlaunchagentsecrets.CursorProjectName)
	if got != want {
		t.Fatalf("CursorCommandProjectDir() = %q, want %q", got, want)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("CursorCommandProjectDir() = %q, want absolute", got)
	}
	if filepath.Base(got) != commandlaunchagentsecrets.CursorProjectName {
		t.Fatalf("project base = %q, want %q", filepath.Base(got), commandlaunchagentsecrets.CursorProjectName)
	}
}

func TestCursorCommandInheritedEnvNameAllow_returnsCopy_withoutMutatingSSoT(t *testing.T) {
	t.Parallel()

	// Given: 公開 allowlist
	first := CursorCommandInheritedEnvNameAllow()
	if len(first) == 0 {
		t.Fatal("CursorCommandInheritedEnvNameAllow() = empty, want non-empty")
	}
	original := first[0]

	// When: 呼び出し側が戻りを書き換える
	first[0] = "MUTATED"

	// Then: SSoT は壊れない
	second := CursorCommandInheritedEnvNameAllow()
	if second[0] != original {
		t.Fatalf("SSoT mutated: second[0] = %q, want %q", second[0], original)
	}
}
