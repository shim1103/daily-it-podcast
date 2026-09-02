//go:build system

// Scope: System（Gemini 以外 full）
// 実物: GetX API / Cursor CLI / Google OAuth / Google Drive を実 secret の production Adapter で通す。
// Double: SpeechSynthesizer だけ無音 WAV を返す test fake（Gemini API は 1 リクエストも発行しない）。
// 目的: Gemini 無料枠 quota を一切消費せず、GetX fetch → Cursor draft → OAuth → Drive 書込／読取／削除の
//
//	実到達と所要時間を 1 回で測る（Decision 2026-09-02T16-57-00）。
//
// @require process env に GETX_API_KEY / CURSOR_API_KEY / GOOGLE_OAUTH_CLIENT_ID / GOOGLE_OAUTH_CLIENT_SECRET /
//
//	GOOGLE_OAUTH_REFRESH_TOKEN / DRIVE_FOLDER_ID がある（どれか欠けたら Skip）。Cursor CLI の `agent` が PATH で
//	解決できる（無ければ Skip）。GEMINI_API_KEY は要らない。Fetch 窓内に SourceItem ≥1。DRIVE_FOLDER_ID は test 専用 folder。
//
// @ensure ProduceEpisode.Run が nil を返す。実行前後 diff の 1 stem について `{stem}.json`+`{stem}.wav` が非空で揃い、
//
//	JSON の episodeId が stem と一致する。本 run の成果物は cleanup で実削除する。各区間の経過秒を Logf で出す。
//
// @invariant local に secret を置かない。本番 folder を使わない。Composition を触らない。Gemini へ 1 リクエストも出さない。
// @invariant Domain 定数（Draft* / CharsPerSecond / SegmentSilenceSec / ModelID 等）を変更しない。
package system

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/google/oauth"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorcli"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/getxapi"
)

// geminiExcludedSpeechFake は port.SpeechSynthesizer を満たす test fake。
// 各セグメントで固定 1.5 秒の 24kHz/16bit/mono 無音 WAV を返す。Gemini へは 1 リクエストも出さない。
type geminiExcludedSpeechFake struct {
	calls int
}

func (f *geminiExcludedSpeechFake) Synthesize(_ context.Context, _ string) (models.SpeechAudio, error) {
	f.calls++
	return models.SpeechAudio{Content: silentWAV(1.5)}, nil
}

// silentWAV は sec 秒のゼロ埋め data を持つ 44 byte 標準 header の RIFF/WAVE を返す。
// PCM パラメータは Gemini Adapter の実出力（24kHz / 16bit / mono）に合わせる。
// build.WavDurationSec が正の秒数を返し、build.ConcatWAV の PCM パラメータ一致検査も通る。
func silentWAV(sec float64) []byte {
	const (
		sampleRate    = 24000
		channels      = 1
		bitsPerSample = 16
	)
	blockAlign := channels * bitsPerSample / 8
	byteRate := sampleRate * blockAlign
	dataLen := int(math.Round(sec*float64(sampleRate))) * blockAlign

	out := make([]byte, 44+dataLen)
	copy(out[0:4], "RIFF")
	binary.LittleEndian.PutUint32(out[4:8], uint32(36+dataLen))
	copy(out[8:12], "WAVE")
	copy(out[12:16], "fmt ")
	binary.LittleEndian.PutUint32(out[16:20], 16)
	binary.LittleEndian.PutUint16(out[20:22], 1) // PCM
	binary.LittleEndian.PutUint16(out[22:24], uint16(channels))
	binary.LittleEndian.PutUint32(out[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(out[28:32], uint32(byteRate))
	binary.LittleEndian.PutUint16(out[32:34], uint16(blockAlign))
	binary.LittleEndian.PutUint16(out[34:36], uint16(bitsPerSample))
	copy(out[36:40], "data")
	binary.LittleEndian.PutUint32(out[40:44], uint32(dataLen))
	// out[44:] はゼロ埋めのまま（無音）。
	return out
}

// geminiExcludedRequiredEnv は GEMINI_API_KEY を除いた必須 env（Decision 2026-09-02T16-57-00）。
var geminiExcludedRequiredEnv = []string{
	config.GetXAPIKeyEnv,
	config.CursorAPIKeyEnv,
	config.GoogleOAuthClientIDEnv,
	config.GoogleOAuthClientSecretEnv,
	config.GoogleOAuthRefreshTokenEnv,
	config.DriveFolderIDEnv,
}

// requireGeminiExcludedEnv は必須 env と Cursor CLI `agent` の PATH 解決を確認し、欠けていれば Skip する。
func requireGeminiExcludedEnv(t *testing.T) {
	t.Helper()
	var missing []string
	for _, key := range geminiExcludedRequiredEnv {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		t.Skipf("System precondition: process env 不足: %s（Gemini 以外 full を skip）", strings.Join(missing, ", "))
	}
	if _, err := exec.LookPath(cursorcli.BinaryName); err != nil {
		t.Skipf("System precondition: Cursor CLI %q が PATH で解決できない（Gemini 以外 full を skip）: %v", cursorcli.BinaryName, err)
	}
}

func displayLocationJST(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(constants.DisplayTimeZone)
	if err != nil {
		t.Fatalf("time.LoadLocation(%q): %v", constants.DisplayTimeZone, err)
	}
	return loc
}

func TestProduceEpisodeSystem_geminiExcluded_realBoundariesExceptSpeech_writesJsonAndWavPair(t *testing.T) {
	// Given: GEMINI を除く実 secret と test Drive folder の一覧 snapshot
	requireGeminiExcludedEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	obs := newDriveObserver(t)

	beforeStart := time.Now()
	before, err := obs.listFolder(ctx)
	if err != nil {
		t.Fatalf("Drive list (before): %v", err)
	}
	t.Logf("Drive list(before): %d files, %.2fs", len(before), time.Since(beforeStart).Seconds())

	// 同日ペアが既にあると Run は Fetch 前 skip で nil を返す（Decision 2026-08-30T23-30-00）。暗黙 skip にしない。
	if _, _, _, ok := completeStem(mapValues(before)); ok {
		t.Fatalf("test Drive folder に既存の完成ペアが残っている。当日ペアだと Run が Fetch 前 skip する。手動掃除が要る")
	}

	var createdIDs []string
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cleanCancel()
		for _, id := range createdIDs {
			delStart := time.Now()
			if delErr := obs.delete(cleanCtx, id); delErr != nil {
				t.Errorf("Drive cleanup id=%s: %v", id, delErr)
				continue
			}
			t.Logf("Drive delete id=%s: %.2fs", id, time.Since(delStart).Seconds())
		}
	})

	// Given: GetX / Cursor / OAuth / Drive は実 Adapter、Speech だけ無音 WAV fake
	httpClient := &http.Client{Timeout: 30 * time.Second}

	getxSource := getxapi.NewPostSource(httpClient, os.Getenv(config.GetXAPIKeyEnv))
	fetch := application.NewFetchSourceItems(getxSource)

	cursorFactory := processenv.NewSecretEnvLauncherFactory(os.Getenv(config.CursorAPIKeyEnv), os.LookupEnv)
	textWriter := cursorcli.NewTextWriter(cursorFactory)

	tokens := oauth.NewTokenSource(
		httpClient,
		os.Getenv(config.GoogleOAuthClientIDEnv),
		os.Getenv(config.GoogleOAuthClientSecretEnv),
		os.Getenv(config.GoogleOAuthRefreshTokenEnv),
	)
	folderID := os.Getenv(config.DriveFolderIDEnv)
	lookup := gdrive.NewCompletedEpisodeLookup(httpClient, tokens, folderID)
	rawWriter := gdrive.NewRawEpisodeWriter(httpClient, tokens, folderID)
	writeEpisode := application.NewWriteEpisode(rawWriter)

	speechFake := &geminiExcludedSpeechFake{}

	episodeIDPrefix := fmt.Sprintf("sys-nogemini-%d", time.Now().UnixNano())
	newEpisodeID := func() string { return episodeIDPrefix }

	uc := application.NewProduceEpisode(
		fetch,
		lookup,
		textWriter,
		speechFake,
		writeEpisode,
		newEpisodeID,
		displayLocationJST(t),
	)

	// When: Gemini を 0 リクエストで Run を 1 回通す
	runStart := time.Now()
	runErr := uc.Run(ctx, time.Now())
	runElapsed := time.Since(runStart)
	t.Logf("ProduceEpisode.Run 全体: %.2fs（Speech fake 呼び出し %d 回 / Gemini 0 リクエスト）", runElapsed.Seconds(), speechFake.calls)

	// Then: Run が nil。新規 stem の json+wav が契約どおり
	if runErr != nil {
		t.Fatalf("ProduceEpisode.Run: %v", runErr)
	}
	if speechFake.calls == 0 {
		t.Fatalf("Speech fake が 1 度も呼ばれていない（Fetch 前 skip か途中 error の疑い）")
	}

	afterStart := time.Now()
	after, err := obs.listFolder(ctx)
	if err != nil {
		t.Fatalf("Drive list (after): %v", err)
	}
	t.Logf("Drive list(after): %d files, %.2fs", len(after), time.Since(afterStart).Seconds())

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

	downloadStart := time.Now()
	raw, err := obs.download(ctx, jsonFile.ID)
	if err != nil {
		t.Fatalf("download %s: %v", jsonFile.Name, err)
	}
	t.Logf("Drive download %s: %.2fs", jsonFile.Name, time.Since(downloadStart).Seconds())

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

// mapValues は driveFile map を slice へ落とす（completeStem は slice を取る）。
func mapValues(m map[string]driveFile) []driveFile {
	out := make([]driveFile, 0, len(m))
	for _, f := range m {
		out = append(out, f)
	}
	return out
}
