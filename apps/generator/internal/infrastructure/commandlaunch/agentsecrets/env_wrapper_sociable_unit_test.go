package agentsecrets_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

func TestCommand_placesChildArgvAfterEnvSubcommandAndSeparator_whenChildArgvIsGiven(t *testing.T) {
	t.Parallel()

	// Given: project dir と依存秘密名を宣言した wrapper
	wrapper := agentsecrets.EnvWrapper{
		ProjectDir: "/path/to/project",
		SecretKeys: []string{secretnames.CursorAPIKeyName},
	}

	// When: 子 command の argv を包む
	name, args := wrapper.Command([]string{"agent", "-p", "--trust"})

	// Then: exec する program は wrapper の binary で、argv は env と -- を経て子 argv へ続く
	if name != agentsecrets.EnvBinary {
		t.Fatalf("name = %q, want %q", name, agentsecrets.EnvBinary)
	}
	want := []string{agentsecrets.EnvSubcommand, agentsecrets.ArgSeparator, "agent", "-p", "--trust"}
	if !slices.Equal(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
}

func TestCommand_returnsOnlyWrapperArgv_whenChildArgvIsEmpty(t *testing.T) {
	t.Parallel()

	// Given: 子 argv を持たない wrapper
	wrapper := agentsecrets.EnvWrapper{ProjectDir: "/path/to/project"}

	// When: 空の子 argv を包む
	_, args := wrapper.Command(nil)

	// Then: separator までは維持され、後続要素は無い
	want := []string{agentsecrets.EnvSubcommand, agentsecrets.ArgSeparator}
	if !slices.Equal(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
}

func TestCommand_omitsSecretKeysFromArgv_whenSecretKeysAreDeclared(t *testing.T) {
	t.Parallel()

	// Given: 依存する秘密名を宣言した wrapper
	// why: `agentsecrets env` は key 名を受け取る flag を持たない。宣言が argv へ漏れると
	// 存在しない flag として wrapper の起動が壊れる。
	wrapper := agentsecrets.EnvWrapper{
		ProjectDir: "/path/to/project",
		SecretKeys: []string{secretnames.CursorAPIKeyName},
	}

	// When: 子 command の argv を包む
	_, args := wrapper.Command([]string{"agent"})

	// Then: 宣言した key 名は argv のどこにも現れない
	if joined := strings.Join(args, " "); strings.Contains(joined, secretnames.CursorAPIKeyName) {
		t.Fatalf("args %q contains declared secret key name", joined)
	}
}

func TestValidate_returnsNil_whenProjectDirIsSet(t *testing.T) {
	t.Parallel()

	// Given: project dir を持つ wrapper
	wrapper := agentsecrets.EnvWrapper{ProjectDir: "/path/to/project"}

	// When: 妥当性を確認する
	err := wrapper.Validate()

	// Then: error は無い
	if err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestValidate_returnsError_whenProjectDirIsBlank(t *testing.T) {
	t.Parallel()

	// Given: project dir が空白だけの wrapper
	// why: project dir が決まらないと active project も決まらず、全 secret が子 process へ渡る。
	wrapper := agentsecrets.EnvWrapper{ProjectDir: "  \t "}

	// When: 妥当性を確認する
	err := wrapper.Validate()

	// Then: error が返る
	if err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}

func TestValidate_returnsError_whenProjectDirIsRelative(t *testing.T) {
	t.Parallel()

	// Given: 相対 path を project dir に持つ wrapper
	// why: 相対 path は親 process の cwd を基準に解決されるため、起動位置が変わると
	// 別 project、または未紐付けの状態で wrapper が動き、渡る秘密の範囲が黙って変わる。
	wrapper := agentsecrets.EnvWrapper{ProjectDir: "secrets/cursor"}

	// When: 妥当性を確認する
	err := wrapper.Validate()

	// Then: error が返る
	if err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}

func TestProjectDir_joinsHomeWithProjectsRoot_whenHomeAndNameAreGiven(t *testing.T) {
	t.Parallel()

	// Given: home と project 名
	// When: project dir を組み立てる
	got := agentsecrets.ProjectDir("/Users/example", "cursor")

	// Then: AgentSecrets の project 配置規約に従った絶対 path が返る
	want := "/Users/example/" + agentsecrets.ProjectsRootName + "/cursor"
	if got != want {
		t.Fatalf("ProjectDir() = %q, want %q", got, want)
	}
}

func TestProjectDir_returnsRelativePath_whenHomeIsEmpty(t *testing.T) {
	t.Parallel()

	// Given: home が空（環境不備）
	// When: project dir を組み立てる
	got := agentsecrets.ProjectDir("", "cursor")

	// Then: 絶対 path にならず、Validate が実行時 error として弾ける形で返る
	if filepath.IsAbs(got) {
		t.Fatalf("ProjectDir() = %q, want non-absolute so Validate rejects it", got)
	}
	if err := (agentsecrets.EnvWrapper{ProjectDir: got}).Validate(); err == nil {
		t.Fatal("Validate() = nil, want error for path built from empty home")
	}
}

func TestDefaultProjectDir_resolvesFromEnvironment_whenHomeIsSet(t *testing.T) {
	// Given: HOME が設定された環境
	t.Setenv("HOME", "/Users/example")

	// When: project 名だけを渡して既定 dir を解決する
	got := agentsecrets.DefaultProjectDir("cursor")

	// Then: 呼び出し側が HOME を読まずに絶対 path を得る
	want := "/Users/example/" + agentsecrets.ProjectsRootName + "/cursor"
	if got != want {
		t.Fatalf("DefaultProjectDir() = %q, want %q", got, want)
	}
}

func TestDefaultProjectDir_returnsPathRejectedByValidate_whenHomeIsUnset(t *testing.T) {
	// Given: HOME が空の環境
	t.Setenv("HOME", "")

	// When: 既定 dir を解決する
	got := agentsecrets.DefaultProjectDir("cursor")

	// Then: Validate が実行時 error として弾ける形で返る
	if err := (agentsecrets.EnvWrapper{ProjectDir: got}).Validate(); err == nil {
		t.Fatalf("Validate() = nil for %q, want error", got)
	}
}
