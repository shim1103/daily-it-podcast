// Scope: Integration test 共通 support（Narrow / Broad 中立）
// 実物境界: なし（test double 組み立て helper のみ）
// Double: httptest TLS redirect・fake agent script・wire JSON fixture
// @invariant dummy secret 実値は helper が error message へ出さない。
package test

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/google/oauth"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorcli"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/getxapi"
)

const (
	broadIntegrationTopicCount = constants.DraftTopicCountMin

	// integrationTTSFixedSegmentCount は TTS 束の固定 segment 数（greeting+intro 束 / closingSummary+farewell 束）。
	// SpeechTexts が topic+2 束を返すため（Decision 2026-09-02T13-55-00）。
	integrationTTSFixedSegmentCount = 2

	broadDummyGetXAPIKey        = "broad-getxapi-dummy-key-value"
	broadDummyCursorKey         = "broad-cursor-dummy-key-value"
	broadDummyGeminiKey         = "broad-gemini-dummy-key-value"
	broadDummyOAuthClientID     = "broad-oauth-client-id-dummy-value"
	broadDummyOAuthClientSecret = "broad-oauth-client-secret-dummy-value"
	broadDummyOAuthRefreshToken = "broad-oauth-refresh-token-dummy-value"
	broadDummyDriveFolderID     = "broad-drive-folder-id-dummy-value"
	broadDummyOAuthAccessToken  = "ya29.broad-access-token-dummy-value"
	broadFixedEpisodeID         = "broad-ep-fixed-0001"
)

var broadDummySecrets = []string{
	broadDummyGetXAPIKey,
	broadDummyCursorKey,
	broadDummyGeminiKey,
	broadDummyOAuthClientID,
	broadDummyOAuthClientSecret,
	broadDummyOAuthRefreshToken,
	broadDummyDriveFolderID,
	broadDummyOAuthAccessToken,
}

// integrationTestDisplayLocation は Broad / Integration test 用の固定 JST Location。
var integrationTestDisplayLocation = time.FixedZone("JST", 9*3600)

// integrationTestFixedNow は Broad Integration の Run 引数に渡す固定時刻。
var integrationTestFixedNow = time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC)

func broadFixedEpisodeIDFunc() string { return broadFixedEpisodeID }

// compositeItemSource は composition.newCompositeItemSource と同型の合成 port.ItemSource。
// why: composition は test から import できないため、同型の写しを test 内に置く。
type compositeItemSource []port.ItemSource

func (c compositeItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	merged := make([]models.SourceItem, 0)
	for _, source := range c {
		items, err := source.List(ctx, since)
		if err != nil {
			return nil, err
		}
		merged = append(merged, items...)
	}
	return merged, nil
}

// integrationWireTopic は buildIntegrationWireJSON が組む topic 1 件分の素材。
type integrationWireTopic struct {
	Title   string `json:"title"`
	Preface string `json:"preface"`
	Detail  string `json:"detail"`
}

func integrationJaRunes(n int) string {
	return strings.Repeat("あ", n)
}

func integrationJaField(n int) string {
	return integrationJaRunes(n) + string(constants.DraftSentenceSuffixRune)
}

func integrationTotalNarrationRunes(introLen, closingLen, prefaceLen, detailLen, topicCount int) int {
	return introLen + closingLen + topicCount*(prefaceLen+detailLen)
}

// buildIntegrationWireJSON は topicCount 件の valid wire JSON を返す。
// 各 field は min 長から始め、total range 下限を満たすまで detail→preface→intro→closing の順で延長する。
// range 網羅は Sociable Unit が所有する。
func buildIntegrationWireJSON(topicCount int) string {
	introLen := constants.DraftIntroMinLen
	closingLen := constants.DraftClosingMinLen
	prefaceLen := constants.DraftTopicPrefaceMinLen
	detailLen := constants.DraftTopicDetailMinLen

	for integrationTotalNarrationRunes(introLen, closingLen, prefaceLen, detailLen, topicCount) < constants.DraftTotalCharsMin {
		switch {
		case detailLen < constants.DraftTopicDetailMaxLen:
			detailLen++
		case prefaceLen < constants.DraftTopicPrefaceMaxLen:
			prefaceLen++
		case introLen < constants.DraftIntroMaxLen:
			introLen++
		case closingLen < constants.DraftClosingMaxLen:
			closingLen++
		default:
			panic("buildIntegrationWireJSON: total chars min を topicCount で満たせない")
		}
	}

	title := integrationJaRunes(constants.DraftTitleMinLen)
	intro := integrationJaField(introLen - 1)
	closing := integrationJaField(closingLen - 1)

	topics := make([]integrationWireTopic, topicCount)
	for i := 0; i < topicCount; i++ {
		topics[i] = integrationWireTopic{
			Title:   integrationJaRunes(constants.DraftTopicTitleMinLen),
			Preface: integrationJaField(prefaceLen - 1),
			Detail:  integrationJaField(detailLen - 1),
		}
	}
	doc := map[string]any{
		"title":          title,
		"intro":          intro,
		"topics":         topics,
		"closingSummary": closing,
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

func integrationSynthesizeCallCount(topicCount int) int {
	return integrationTTSFixedSegmentCount + topicCount
}

func assertIntegrationSecretsNotLeaked(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	msg := err.Error()
	for _, secret := range broadDummySecrets {
		if strings.Contains(msg, secret) {
			t.Fatalf("error message contains dummy secret value")
		}
	}
}

func integrationCallCount(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatalf("os.ReadFile(%q) error = %v, want nil or not exist", path, err)
	}
	if len(data) == 0 {
		return 0
	}
	return strings.Count(string(data), "\n")
}

type fakeAgentCLIConfig struct {
	recordStart  bool
	captureStdin bool
	countWrites  bool
}

type fakeAgentCLIResult struct {
	markerPath     string
	stdinSinkPath  string
	writeCountPath string
}

// installFakeAgentCLI は cursorcli.BinaryName 名の fake script を PATH 先頭へ置く。
// recordStart / captureStdin は Narrow 用、countWrites は Broad 用の Write 到達回数計測。
func installFakeAgentCLI(t *testing.T, scriptBody string, cfg fakeAgentCLIConfig) fakeAgentCLIResult {
	t.Helper()
	dir := t.TempDir()
	var result fakeAgentCLIResult
	preamble := []string{"#!/bin/sh"}
	if cfg.recordStart {
		result.markerPath = filepath.Join(dir, "started")
		preamble = append(preamble, "touch '"+result.markerPath+"'")
	}
	if cfg.captureStdin {
		result.stdinSinkPath = filepath.Join(dir, "stdin")
		preamble = append(preamble, "cat > '"+result.stdinSinkPath+"'")
	} else if cfg.countWrites {
		result.writeCountPath = filepath.Join(dir, "write_count")
		preamble = append(preamble, "echo 1 >> '"+result.writeCountPath+"'")
		preamble = append(preamble, "cat > /dev/null")
	}
	script := strings.Join(preamble, "\n") + "\n" + scriptBody + "\n"
	program := filepath.Join(dir, cursorcli.BinaryName)
	if err := os.WriteFile(program, []byte(script), 0o755); err != nil {
		t.Fatalf("os.WriteFile() error = %v, want nil", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return result
}

func cursorSuccessScriptBody(t *testing.T, dir, wireJSON string) string {
	t.Helper()
	envelope := map[string]any{
		"type":     "result",
		"subtype":  "success",
		"is_error": false,
		"result":   wireJSON,
	}
	raw, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v, want nil", err)
	}
	path := filepath.Join(dir, "envelope.json")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v, want nil", err)
	}
	return fmt.Sprintf("cat %q", path)
}

func minimalIntegrationGeminiPCM() []byte {
	// why: Adapter の最小尺閾値（0.5s）を超える長さ。これ未満だと極小 PCM として retry される。
	const sampleCount = 24000 // 1.0s 相当
	return make([]byte, sampleCount*2)
}

func writeIntegrationGeminiAudioResponse(t *testing.T, w http.ResponseWriter, pcm []byte) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	body, err := json.Marshal(map[string]any{
		"status": "completed",
		"steps": []map[string]any{
			{"content": []map[string]any{
				{"data": base64.StdEncoding.EncodeToString(pcm)},
			}},
		},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if _, err := w.Write(body); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

type integrationTLSRoutes struct {
	getxapi http.HandlerFunc
	gemini  http.HandlerFunc
	oauth   http.HandlerFunc
	gdrive  http.HandlerFunc
}

func newIntegrationTLSClient(t *testing.T, routes integrationTLSRoutes) *http.Client {
	t.Helper()
	servers := map[string]*httptest.Server{
		"api.getxapi.com":                   httptest.NewTLSServer(routes.getxapi),
		"generativelanguage.googleapis.com": httptest.NewTLSServer(routes.gemini),
		"oauth2.googleapis.com":             httptest.NewTLSServer(routes.oauth),
		"www.googleapis.com":                httptest.NewTLSServer(routes.gdrive),
	}
	for _, srv := range servers {
		t.Cleanup(srv.Close)
	}
	addrs := make(map[string]string, len(servers))
	for host, srv := range servers {
		addrs[host] = srv.Listener.Addr().String()
	}
	return &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(_ context.Context, network, addr string) (net.Conn, error) {
				host, _, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				target, ok := addrs[host]
				if !ok {
					return nil, fmt.Errorf("unexpected TLS host %q", host)
				}
				return tls.Dial(network, target, &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // why: test server 自己署名を信頼する。
			},
		},
	}
}

func integrationOAuthSuccessHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"access_token":"` + broadDummyOAuthAccessToken + `"}`))
}

func integrationGetXAPITweetResponse() string {
	return `{"tweets":[{"id":"broad-tweet-1","url":"https://x.example/broad-1","text":"Broad integration 本文","createdAt":"Wed Aug 30 10:00:00 +0000 2026","author":{"id":"author-broad","name":"Broad Author"},"entities":{"urls":[{"expanded_url":"https://example.com/broad"}]},"media":[{"url":"https://img.example/broad.jpg"}]}],"has_more":false}`
}

type integrationGDriveProbe struct {
	uploadPATCH int64
}

func integrationGDriveSuccessHandler(t *testing.T, probe *integrationGDriveProbe) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.String()
		switch {
		case r.Method == http.MethodGet && strings.Contains(target, "/drive/v3/files"):
			writeIntegrationJSONStatus(t, w, http.StatusOK, map[string]any{"files": []any{}})
		case r.Method == http.MethodPost && strings.Contains(target, "/drive/v3/files"):
			var meta struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&meta); err != nil {
				t.Fatalf("decode create metadata: %v", err)
			}
			writeIntegrationJSONStatus(t, w, http.StatusOK, map[string]any{"id": "created-" + meta.Name})
		case r.Method == http.MethodPatch && strings.Contains(target, "/upload/drive/v3/files/"):
			if probe != nil {
				probe.uploadPATCH++
			}
			writeIntegrationJSONStatus(t, w, http.StatusOK, map[string]any{"id": "uploaded"})
		default:
			t.Fatalf("unexpected gdrive request method=%s url=%s", r.Method, target)
		}
	}
}

func writeIntegrationJSONStatus(t *testing.T, w http.ResponseWriter, status int, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
}

type broadProduceEpisodeConfig struct {
	emptyGetXAPI bool
	cursorFail   bool
	geminiFailAt int // 1-origin。0 なら失敗しない
}

type broadProduceEpisodeHarness struct {
	uc               *application.ProduceEpisode
	cursorWriteCount string
	geminiPosts      atomic.Int32
	gdriveUploads    *integrationGDriveProbe
}

func assertBroadDownstreamCalls(t *testing.T, h *broadProduceEpisodeHarness, wantCursor, wantGemini, wantUpload int) {
	t.Helper()
	if wantCursor >= 0 {
		if got := integrationCallCount(t, h.cursorWriteCount); got != wantCursor {
			t.Fatalf("TextWriter calls = %d, want %d", got, wantCursor)
		}
	}
	if got := int(h.geminiPosts.Load()); got != wantGemini {
		t.Fatalf("Synthesize calls = %d, want %d", got, wantGemini)
	}
	if got := h.gdriveUploads.uploadPATCH; got != int64(wantUpload) {
		t.Fatalf("Drive upload calls = %d, want %d", got, wantUpload)
	}
}

func newBroadProduceEpisodeHarness(t *testing.T, cfg broadProduceEpisodeConfig) *broadProduceEpisodeHarness {
	t.Helper()

	previousWatch := x.WatchUserIDs
	x.WatchUserIDs = []string{"broad-user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previousWatch })

	wireJSON := buildIntegrationWireJSON(broadIntegrationTopicCount)
	agentDir := t.TempDir()

	var cursorBody string
	switch {
	case cfg.cursorFail:
		cursorBody = `printf '%s' 'broad cursor failure' 1>&2
exit 1`
	default:
		cursorBody = cursorSuccessScriptBody(t, agentDir, wireJSON)
	}
	fakeAgent := installFakeAgentCLI(t, cursorBody, fakeAgentCLIConfig{countWrites: true})

	gdriveProbe := &integrationGDriveProbe{}

	getxHandler := func(w http.ResponseWriter, _ *http.Request) {
		if cfg.emptyGetXAPI {
			_, _ = io.WriteString(w, `{"tweets":[],"has_more":false}`)
			return
		}
		_, _ = io.WriteString(w, integrationGetXAPITweetResponse())
	}

	h := &broadProduceEpisodeHarness{gdriveUploads: gdriveProbe}
	geminiHandler := func(w http.ResponseWriter, r *http.Request) {
		n := h.geminiPosts.Add(1)
		if cfg.geminiFailAt != 0 && int(n) == cfg.geminiFailAt {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"INVALID_ARGUMENT"}`))
			return
		}
		writeIntegrationGeminiAudioResponse(t, w, minimalIntegrationGeminiPCM())
	}

	httpClient := newIntegrationTLSClient(t, integrationTLSRoutes{
		getxapi: http.HandlerFunc(getxHandler),
		gemini:  http.HandlerFunc(geminiHandler),
		oauth:   http.HandlerFunc(integrationOAuthSuccessHandler),
		gdrive:  integrationGDriveSuccessHandler(t, gdriveProbe),
	})

	getxSource := getxapi.NewPostSource(httpClient, broadDummyGetXAPIKey)
	fetch := application.NewFetchSourceItems(compositeItemSource{getxSource})
	cursorFactory := processenv.NewSecretEnvLauncherFactory(broadDummyCursorKey, os.LookupEnv)
	textWriter := cursorcli.NewTextWriter(cursorFactory)
	speech := gemini.NewSpeechSynthesizer(httpClient, broadDummyGeminiKey)
	tokens := oauth.NewTokenSource(httpClient, broadDummyOAuthClientID, broadDummyOAuthClientSecret, broadDummyOAuthRefreshToken)
	lookup := gdrive.NewCompletedEpisodeLookup(httpClient, tokens, broadDummyDriveFolderID)
	rawWriter := gdrive.NewRawEpisodeWriter(httpClient, tokens, broadDummyDriveFolderID)
	writeEpisode := application.NewWriteEpisode(rawWriter)

	h.uc = application.NewProduceEpisode(
		fetch,
		lookup,
		textWriter,
		speech,
		writeEpisode,
		broadFixedEpisodeIDFunc,
		integrationTestDisplayLocation,
	)
	h.cursorWriteCount = fakeAgent.writeCountPath
	return h
}
