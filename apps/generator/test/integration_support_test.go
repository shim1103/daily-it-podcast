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
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/google/oauth"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/hackernews"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/itmedia"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/lobsters"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
)

const (
	broadIntegrationTopicCount = constants.DraftTopicCountMin

	// integrationTTSFixedSegmentCount は TTS 順の固定 segment 数（Greeting + Intro + Summary + Farewell）。
	integrationTTSFixedSegmentCount = 4

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
	return integrationTTSFixedSegmentCount + 2*topicCount
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

func minimalIntegrationGeminiPCM() []byte {
	const sampleCount = 2400
	return make([]byte, sampleCount*2)
}

func writeIntegrationGeminiAudioResponse(t *testing.T, w http.ResponseWriter, pcm []byte) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	body, err := json.Marshal(map[string]any{
		"output_audio": map[string]any{
			"data": base64.StdEncoding.EncodeToString(pcm),
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
	gemini     http.HandlerFunc
	oauth      http.HandlerFunc
	gdrive     http.HandlerFunc
	hackernews http.HandlerFunc
	lobsters   http.HandlerFunc
	itmedia    http.HandlerFunc
}

func newIntegrationTLSClient(t *testing.T, routes integrationTLSRoutes) *http.Client {
	t.Helper()
	servers := map[string]*httptest.Server{
		"generativelanguage.googleapis.com": httptest.NewTLSServer(routes.gemini),
		"oauth2.googleapis.com":             httptest.NewTLSServer(routes.oauth),
		"www.googleapis.com":                httptest.NewTLSServer(routes.gdrive),
		"hacker-news.firebaseio.com":        httptest.NewTLSServer(routes.hackernews),
		"lobste.rs":                         httptest.NewTLSServer(routes.lobsters),
		"rss.itmedia.co.jp":                 httptest.NewTLSServer(routes.itmedia),
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
	cursorFail   bool
	geminiFailAt int  // 1-origin。0 なら失敗しない
	emptySources bool // true なら 3 情報源すべてが 0 件を返す
}

// 3 情報源の success / empty handler。時刻は integrationTestFixedNow を使う。
// FetchSourceItems は since = now - FetchWindow(24h) を渡すため、now 自体は必ず since 以上になる。
// Broad は「SourceItem が 1 件以上ある」ことだけを要求する。
func integrationHackerNewsSuccessHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	storyUnix := integrationTestFixedNow.Unix()
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/topstories.json"):
			_, _ = io.WriteString(w, "[9001]")
		case strings.HasSuffix(r.URL.Path, "/item/9001.json"):
			_, _ = io.WriteString(w, fmt.Sprintf(
				`{"id":9001,"type":"story","time":%d,"title":"Broad HackerNews 記事"}`,
				storyUnix,
			))
		default:
			http.Error(w, "unexpected hackernews path", http.StatusNotFound)
		}
	}
}

func integrationHackerNewsEmptyHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/topstories.json") {
			_, _ = io.WriteString(w, "[]")
			return
		}
		http.Error(w, "unexpected hackernews path", http.StatusNotFound)
	}
}

func integrationLobstersSuccessHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	createdAt := integrationTestFixedNow.Format(time.RFC3339)
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/hottest.json"):
			_, _ = io.WriteString(w, fmt.Sprintf(`[{"short_id":"broad1","created_at":%q}]`, createdAt))
		case strings.HasSuffix(r.URL.Path, "/s/broad1.json"):
			_, _ = io.WriteString(w, fmt.Sprintf(
				`{"short_id":"broad1","title":"Broad Lobsters 記事","created_at":%q}`,
				createdAt,
			))
		default:
			http.Error(w, "unexpected lobsters path", http.StatusNotFound)
		}
	}
}

func integrationLobstersEmptyHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/hottest.json") {
			_, _ = io.WriteString(w, "[]")
			return
		}
		http.Error(w, "unexpected lobsters path", http.StatusNotFound)
	}
}

func integrationITmediaSuccessHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	pubDate := integrationTestFixedNow.Format(time.RFC1123Z)
	body := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
		`<rss version="2.0">` + "\n<channel>\n" +
		fmt.Sprintf(
			"<item><title>%s</title><link>%s</link><description>%s</description><pubDate>%s</pubDate></item>\n",
			"Broad ITmedia 記事", "https://www.itmedia.co.jp/news/articles/broad.html", "本文", pubDate,
		) +
		"</channel>\n</rss>\n"
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/rss/2.0/news_bursts.xml") {
			_, _ = io.WriteString(w, body)
			return
		}
		http.Error(w, "unexpected itmedia path", http.StatusNotFound)
	}
}

func integrationITmediaEmptyHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	body := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
		`<rss version="2.0">` + "\n<channel>\n</channel>\n</rss>\n"
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/rss/2.0/news_bursts.xml") {
			_, _ = io.WriteString(w, body)
			return
		}
		http.Error(w, "unexpected itmedia path", http.StatusNotFound)
	}
}

type broadProduceEpisodeHarness struct {
	uc            *application.ProduceEpisode
	textWriter    *broadTextWriter
	geminiPosts   atomic.Int32
	gdriveUploads *integrationGDriveProbe
}

// broadTextWriter は Broad が Application の停止条件を観測するための port.TextWriter double。
type broadTextWriter struct {
	fragment string
	fail     bool
	calls    atomic.Int32
}

func (w *broadTextWriter) Write(_ context.Context, _ string) (string, error) {
	w.calls.Add(1)
	if w.fail {
		return "", errors.New("broad text writer failure")
	}
	return w.fragment, nil
}

func assertBroadDownstreamCalls(t *testing.T, h *broadProduceEpisodeHarness, wantCursor, wantGemini, wantUpload int) {
	t.Helper()
	if wantCursor >= 0 {
		if got := int(h.textWriter.calls.Load()); got != wantCursor {
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

	wireJSON := buildIntegrationWireJSON(broadIntegrationTopicCount)
	gdriveProbe := &integrationGDriveProbe{}

	h := &broadProduceEpisodeHarness{
		gdriveUploads: gdriveProbe,
		textWriter:    &broadTextWriter{fragment: wireJSON, fail: cfg.cursorFail},
	}
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

	hackernewsHandler := integrationHackerNewsSuccessHandler(t)
	lobstersHandler := integrationLobstersSuccessHandler(t)
	itmediaHandler := integrationITmediaSuccessHandler(t)
	if cfg.emptySources {
		hackernewsHandler = integrationHackerNewsEmptyHandler(t)
		lobstersHandler = integrationLobstersEmptyHandler(t)
		itmediaHandler = integrationITmediaEmptyHandler(t)
	}

	httpClient := newIntegrationTLSClient(t, integrationTLSRoutes{
		gemini:     http.HandlerFunc(geminiHandler),
		oauth:      http.HandlerFunc(integrationOAuthSuccessHandler),
		gdrive:     integrationGDriveSuccessHandler(t, gdriveProbe),
		hackernews: hackernewsHandler,
		lobsters:   lobstersHandler,
		itmedia:    itmediaHandler,
	})

	// 3 情報源（HackerNews → Lobsters → ITmedia）の upstream double を composite ItemSource へ結線する。
	// 登録順は composition.newProduceEpisode と同順。真外部は TLS redirect で double 済み。
	fetch := application.NewFetchSourceItems(compositeItemSource{
		hackernews.NewListItemSource(httpClient),
		lobsters.NewListItemSource(httpClient),
		itmedia.NewListItemSource(httpClient),
	})
	speech := gemini.NewSpeechSynthesizer(httpClient, broadDummyGeminiKey)
	tokens := oauth.NewTokenSource(httpClient, broadDummyOAuthClientID, broadDummyOAuthClientSecret, broadDummyOAuthRefreshToken)
	rawWriter := gdrive.NewRawEpisodeWriter(httpClient, tokens, broadDummyDriveFolderID)
	writeEpisode := application.NewWriteEpisode(rawWriter)

	h.uc = application.NewProduceEpisode(
		fetch,
		h.textWriter,
		speech,
		writeEpisode,
		broadFixedEpisodeIDFunc,
		integrationTestDisplayLocation,
	)
	return h
}
