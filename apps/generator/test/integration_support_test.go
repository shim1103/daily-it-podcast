// Scope: Integration test 共通 support（Narrow / Broad 中立）
// 実物境界: なし（test double 組み立て helper のみ）
// Double: httptest TLS redirect・fake agent script・wire JSON fixture。
// @invariant dummy secret 実値は helper が error message へ出さない。
// @invariant 本番直撃の assert / client 組み立ては接続 cache suite（item_source_connection_cache_test.go）が所有する。
package test

import (
	"bytes"
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
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/fetch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/writeepisode"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/audio/ffmpeg"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/cloudwatch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/hackernews"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/lobsters"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/publickey"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/r2"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/techcrunch"
)

// retryReporterSpy は port.RetryReporter を満たし、Retry 呼び出しを記録する Spy。
// Narrow integration test 群（*_narrow_integration_test.go）が、upstream 5xx から
// 再試行を発生させる test と成功一発の test を同一 helper で共有するために使う。
type retryReporterSpy struct {
	calls int
}

func (s *retryReporterSpy) Retry(step string, attempt, max int, reason string) {
	s.calls++
}

const (
	broadIntegrationTopicCount = constants.DraftTopicCountTarget

	// integrationTTSFixedSegmentCount は TTS 束の固定 segment 数（greeting+intro 束 / closingSummary+farewell 束）。
	// SpeechTexts が topic+2 束を返すため（Decision 2026-09-02T13-55-00）。
	integrationTTSFixedSegmentCount = 2

	broadDummyCursorKey      = "broad-cursor-dummy-key-value"
	broadDummyGeminiKey      = "broad-gemini-dummy-key-value"
	broadDummyR2AccessKeyID  = "broad-r2-access-key-id-dummy-value"
	broadDummyR2SecretAccess = "broad-r2-secret-access-key-dummy-value"
	broadDummyR2AccountID    = "broad-r2-account-id-dummy-value"
	broadDummyR2Bucket       = "broad-r2-bucket-dummy-value"
	broadFixedEpisodeID      = "broad-ep-fixed-0001"
)

var broadDummySecrets = []string{
	broadDummyCursorKey,
	broadDummyGeminiKey,
	broadDummyR2AccessKeyID,
	broadDummyR2SecretAccess,
	broadDummyR2AccountID,
	broadDummyR2Bucket,
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
	gemini     http.HandlerFunc
	r2         http.HandlerFunc
	hackernews http.HandlerFunc
	lobsters   http.HandlerFunc
	publickey  http.HandlerFunc
	techcrunch http.HandlerFunc
	cloudwatch http.HandlerFunc
}

func newIntegrationTLSClient(t *testing.T, routes integrationTLSRoutes) *http.Client {
	t.Helper()
	servers := map[string]*httptest.Server{
		"generativelanguage.googleapis.com":                 httptest.NewTLSServer(routes.gemini),
		broadDummyR2AccountID + ".r2.cloudflarestorage.com": httptest.NewTLSServer(routes.r2),
		"hacker-news.firebaseio.com":                        httptest.NewTLSServer(routes.hackernews),
		"lobste.rs":                                         httptest.NewTLSServer(routes.lobsters),
		"www.publickey1.jp":                                 httptest.NewTLSServer(routes.publickey),
		"techcrunch.com":                                    httptest.NewTLSServer(routes.techcrunch),
		"cloud.watch.impress.co.jp":                         httptest.NewTLSServer(routes.cloudwatch),
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

// integrationR2Probe は Broad が R2 PutObject 呼び出し回数を観測するための counter。
type integrationR2Probe struct {
	putObject int64
}

// integrationR2SuccessHandler は List(空)→常に成功する PutObject を返す S3 互換 double。
// HasPair は毎回「未完成」を返す（List が空）ため、Broad は必ず Write 経路まで進む。
func integrationR2SuccessHandler(t *testing.T, probe *integrationR2Probe) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Query().Get("list-type") == "2":
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`+
				`<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">`+
				`<IsTruncated>false</IsTruncated></ListBucketResult>`)
		case r.Method == http.MethodPut:
			if probe != nil {
				probe.putObject++
			}
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected r2 request method=%s url=%s", r.Method, r.URL.String())
		}
	}
}

type broadProduceEpisodeConfig struct {
	cursorFail   bool
	geminiFailAt int  // 1-origin。0 なら失敗しない
	emptySources bool // true なら 5 源すべてが 0 件を返す
}

// 各情報源の success / empty handler。時刻は integrationTestFixedNow を使う。
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

func integrationPublickeySuccessHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	published := integrationTestFixedNow.UTC().Format(time.RFC3339)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/atom.xml" {
			http.Error(w, "unexpected publickey path", http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, fmt.Sprintf(
			`<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
				`<feed xmlns="http://www.w3.org/2005/Atom">`+"\n"+
				`<entry><title>Broad Publickey 記事</title>`+
				`<link rel="alternate" href="https://www.publickey1.jp/blog/broad.html"/>`+
				`<id>tag:www.publickey1.jp,2026://2.broad</id>`+
				`<published>%s</published>`+
				`<summary>要約</summary>`+
				`<content type="html">本文</content>`+
				`<author><name>broad</name></author>`+
				`</entry></feed>`,
			published,
		))
	}
}

func integrationPublickeyEmptyHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/atom.xml" {
			http.Error(w, "unexpected publickey path", http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
			`<feed xmlns="http://www.w3.org/2005/Atom"><title>Publickey</title></feed>`)
	}
}

func integrationTechCrunchSuccessHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	pubDate := integrationTestFixedNow.UTC().Format(time.RFC1123Z)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/feed/" && r.URL.Path != "/feed" {
			http.Error(w, "unexpected techcrunch path", http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, fmt.Sprintf(
			`<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
				`<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/"><channel>`+
				`<item><title>Broad TechCrunch 記事</title>`+
				`<link>https://techcrunch.com/broad/</link>`+
				`<description>説明</description>`+
				`<pubDate>%s</pubDate>`+
				`<guid isPermaLink="false">https://techcrunch.com/?p=broad</guid>`+
				`<dc:creator>broad-author</dc:creator>`+
				`</item></channel></rss>`,
			pubDate,
		))
	}
}

func integrationTechCrunchEmptyHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/feed/" && r.URL.Path != "/feed" {
			http.Error(w, "unexpected techcrunch path", http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
			`<rss version="2.0"><channel><title>TechCrunch</title></channel></rss>`)
	}
}

func integrationCloudWatchSuccessHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	date := integrationTestFixedNow.UTC().Format(time.RFC3339)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/rss/1.0/clw/feed.rdf" {
			http.Error(w, "unexpected cloudwatch path", http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, fmt.Sprintf(
			`<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
				`<rdf:RDF xmlns="http://purl.org/rss/1.0/" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns:dc="http://purl.org/dc/elements/1.1/">`+
				`<item rdf:about="https://cloud.watch.impress.co.jp/docs/news/broad.html?ref=rss">`+
				`<title>Broad クラウド Watch 記事</title>`+
				`<link>https://cloud.watch.impress.co.jp/docs/news/broad.html</link>`+
				`<dc:date>%s</dc:date>`+
				`<description>説明</description>`+
				`</item></rdf:RDF>`,
			date,
		))
	}
}

func integrationCloudWatchEmptyHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/rss/1.0/clw/feed.rdf" {
			http.Error(w, "unexpected cloudwatch path", http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
			`<rdf:RDF xmlns="http://purl.org/rss/1.0/" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">`+
			`<channel rdf:about="https://cloud.watch.impress.co.jp/data/rss/1.0/clw/feed.rdf">`+
			`<title>クラウド Watch</title></channel></rdf:RDF>`)
	}
}

type broadProduceEpisodeHarness struct {
	uc          *application.ProduceEpisode
	textWriter  *broadTextWriter
	geminiPosts atomic.Int32
	r2Uploads   *integrationR2Probe
	// logBuf は progress reporter (delivery.LogWriter) の結線を Broad で観測するための出力先。
	logBuf *bytes.Buffer
}

// broadTextWriter は Broad が Application の停止条件を観測するための port.TextWriter double。
type broadTextWriter struct {
	fragment string
	fail     bool
	calls    atomic.Int32
}

func (w *broadTextWriter) Write(_ context.Context, _ string, buildFn func(string) (models.ManuscriptDraft, error)) (models.ManuscriptDraft, error) {
	w.calls.Add(1)
	if w.fail {
		return models.ManuscriptDraft{}, errors.New("broad text writer failure")
	}
	return buildFn(w.fragment)
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
	if got := h.r2Uploads.putObject; got != int64(wantUpload) {
		t.Fatalf("R2 upload calls = %d, want %d", got, wantUpload)
	}
}

func newBroadProduceEpisodeHarness(t *testing.T, cfg broadProduceEpisodeConfig) *broadProduceEpisodeHarness {
	t.Helper()

	wireJSON := buildIntegrationWireJSON(broadIntegrationTopicCount)
	r2Probe := &integrationR2Probe{}

	h := &broadProduceEpisodeHarness{
		r2Uploads:  r2Probe,
		textWriter: &broadTextWriter{fragment: wireJSON, fail: cfg.cursorFail},
		logBuf:     &bytes.Buffer{},
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
	publickeyHandler := integrationPublickeySuccessHandler(t)
	techcrunchHandler := integrationTechCrunchSuccessHandler(t)
	cloudwatchHandler := integrationCloudWatchSuccessHandler(t)
	if cfg.emptySources {
		hackernewsHandler = integrationHackerNewsEmptyHandler(t)
		lobstersHandler = integrationLobstersEmptyHandler(t)
		publickeyHandler = integrationPublickeyEmptyHandler(t)
		techcrunchHandler = integrationTechCrunchEmptyHandler(t)
		cloudwatchHandler = integrationCloudWatchEmptyHandler(t)
	}

	httpClient := newIntegrationTLSClient(t, integrationTLSRoutes{
		gemini:     http.HandlerFunc(geminiHandler),
		r2:         integrationR2SuccessHandler(t, r2Probe),
		hackernews: hackernewsHandler,
		lobsters:   lobstersHandler,
		publickey:  publickeyHandler,
		techcrunch: techcrunchHandler,
		cloudwatch: cloudwatchHandler,
	})

	// 5 情報源（HackerNews → Lobsters → Publickey → TechCrunch → クラウド Watch）。
	// 登録順は composition.newProduceEpisode と同順。真外部は TLS redirect で double 済み。
	fetchUC := fetch.NewFetchSourceItems(compositeItemSource{
		hackernews.NewListItemSource(httpClient, hackernews.MaxStoriesScanned, nil),
		lobsters.NewListItemSource(httpClient, lobsters.MaxStoriesScanned, nil),
		publickey.NewListItemSource(httpClient, publickey.MaxStoriesScanned, nil),
		techcrunch.NewListItemSource(httpClient, techcrunch.MaxStoriesScanned, nil),
		cloudwatch.NewListItemSource(httpClient, cloudwatch.MaxStoriesScanned, nil),
	})
	speech := gemini.NewSpeechSynthesizer(httpClient, broadDummyGeminiKey, gemini.TierFree, nil)
	lookup := r2.NewCompletedEpisodeLookup(httpClient, broadDummyR2AccessKeyID, broadDummyR2SecretAccess, broadDummyR2AccountID, broadDummyR2Bucket, nil)
	rawWriter := r2.NewEpisodeWriter(httpClient, broadDummyR2AccessKeyID, broadDummyR2SecretAccess, broadDummyR2AccountID, broadDummyR2Bucket, nil)
	writeEpisode := writeepisode.NewWriteEpisode(rawWriter)

	// progress reporter は production（composition.newProduceEpisode）と同型で delivery.LogWriter そのもの。
	// logBuf へ "generator: category=progress ..." を書かせる。
	logw := delivery.NewLogWriter(h.logBuf)
	h.uc = application.NewProduceEpisode(
		fetchUC,
		lookup,
		h.textWriter,
		speech,
		ffmpeg.NewEncoder(
			func(string) (string, error) { return "/fake/ffmpeg", nil },
			func(_ context.Context, _ string, _ []string, _ []byte) ([]byte, error) {
				return []byte("broad-fake-mp3"), nil
			},
		),
		writeEpisode,
		broadFixedEpisodeIDFunc,
		integrationTestDisplayLocation,
		logw,
		broadIntegrationTopicCount,
	)
	return h
}
