//go:build system

// Scope: System
// 実物: cmd/generator 入口 → GetX / Cursor CLI / Gemini / OAuth+Drive 出口
// Double: なし（test 専用 credential。実行場所は GHA）
// @require process env に config 契約の全 key がある。DRIVE_FOLDER_ID は test 専用 folder。Fetch 窓内に SourceItem ≥1。Cursor CLI の `agent` が PATH で解決できる。
// @ensure subprocess exit 0。実行前後 diff の 1 stem について `{stem}.json`+`{stem}.wav` が非空で揃い、JSON の episodeId が stem と一致する。本 run の成果物は cleanup する。
// @invariant local に secret を置かない。本番 credential / 本番 folder を使わない。下位 Scope の HTTP・配線・schema 全 field を再 assert しない。
package system

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestProduceEpisodeSystem_writesJsonAndWavPair_whenSubprocessSucceeds(t *testing.T) {
	// Given: System 用 process env と test Drive folder の一覧 snapshot
	requireSystemEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Minute)
	defer cancel()

	obs := newDriveObserver(t)
	before, err := obs.listFolder(ctx)
	if err != nil {
		t.Fatalf("Drive list (before): %v", err)
	}

	var createdIDs []string
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cleanCancel()
		for _, id := range createdIDs {
			if delErr := obs.delete(cleanCtx, id); delErr != nil {
				t.Errorf("Drive cleanup id=%s: %v", id, delErr)
			}
		}
	})

	// When: production と同じ cmd/generator 入口を subprocess で実行する
	moduleRoot := generatorModuleRoot(t)
	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/generator")
	cmd.Dir = moduleRoot
	cmd.Env = os.Environ()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	// Then: exit 0。新規 stem の json+wav が契約どおり。秘密値はログに出さない
	if runErr != nil {
		t.Fatalf("cmd/generator failed: %v\nstdout_bytes=%d stderr_bytes=%d", runErr, stdout.Len(), stderr.Len())
	}

	after, err := obs.listFolder(ctx)
	if err != nil {
		t.Fatalf("Drive list (after): %v", err)
	}
	added := addedNames(before, after)
	stem, jsonFile, wavFile, ok := completeStem(added)
	if !ok {
		t.Fatalf("新規 json+wav ペアなし: added=%d", len(added))
	}
	createdIDs = append(createdIDs, jsonFile.ID, wavFile.ID)

	if jsonFile.Size <= 0 {
		t.Fatalf("%s size = %d, want > 0", jsonFile.Name, jsonFile.Size)
	}
	if wavFile.Size <= 0 {
		t.Fatalf("%s size = %d, want > 0", wavFile.Name, wavFile.Size)
	}

	raw, err := obs.download(ctx, jsonFile.ID)
	if err != nil {
		t.Fatalf("download %s: %v", jsonFile.Name, err)
	}
	var manuscript struct {
		EpisodeID string `json:"episodeId"`
	}
	if err := json.Unmarshal(raw, &manuscript); err != nil {
		t.Fatalf("manuscript JSON decode: %v", err)
	}
	if manuscript.EpisodeID != stem {
		t.Fatalf("episodeId = %q, want stem %q", manuscript.EpisodeID, stem)
	}
}

func generatorModuleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller 失敗")
	}
	// test/system/*.go → apps/generator
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
