package twitterapiio_test

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/twitterapiio"
)

// stubBindings は test 用の BindingResolver fake。
type stubBindings map[secrettransport.SecretRef]string

func (b stubBindings) ResolveSecret(ref secrettransport.SecretRef) (string, bool) {
	name, ok := b[ref]
	return name, ok
}

const twitterAPIIOTestSecretName = "TWITTERAPIIO_TEST_API_KEY"

// newTestSecretTransportClient は固定 upstream URL への接続を test TLS server へ差し替えた processenv.Client を返す。
// why: Adapter は本番 host を定数として持つため、DialTLSContext で接続先だけを test server へ redirect する。
func newTestSecretTransportClient(t *testing.T, server *httptest.Server, apiKeySecret secrettransport.SecretRef) secrettransport.Client {
	t.Helper()
	t.Setenv(twitterAPIIOTestSecretName, "twitterapiio-test-real-value")
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, server.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
	return processenv.NewClient(stubBindings{apiKeySecret: twitterAPIIOTestSecretName}, httpClient, nil)
}

func TestList_returnsSourceItems_withoutMediaContext(t *testing.T) {
	// Given: watch user と vendor response を返す upstream double
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"tweets":[{"id":"tweet-1","url":"https://x.example/tweet-1","text":"本文","createdAt":"Wed Aug 19 10:00:00 +0000 2026","author":{"id":"author-1","userName":"user_name"},"entities":{"urls":[{"expanded_url":"https://example.com"}]}}],"has_next_page":false,"status":"success"}`)
	}))
	t.Cleanup(server.Close)
	apiKeySecret := secrettransport.NewSecretRef()
	source := twitterapiio.NewPostSource(newTestSecretTransportClient(t, server, apiKeySecret), apiKeySecret)
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: media 行なしの SourceItem
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	wantOccurredAt := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	wantContext := strings.Join([]string{
		"item_id: tweet-1",
		"actor_id: author-1",
		"actor_name: user_name",
		"text: 本文",
		"permalink: https://x.example/tweet-1",
		"links: https://example.com",
	}, "\n")
	if len(got) != 1 {
		t.Fatalf("got = %+v, want one SourceItem", got)
	}
	if got[0].SourceID != x.SourceID ||
		!got[0].OccurredAt.Equal(wantOccurredAt) ||
		got[0].OccurredAt.Location() != time.UTC ||
		got[0].Context != wantContext {
		t.Fatalf("got[0] = %+v, want SourceID=%q OccurredAt=%v Context=%q", got[0], x.SourceID, wantOccurredAt, wantContext)
	}
	if strings.Contains(got[0].Context, "media:") {
		t.Fatalf("Context = %q, must not contain media", got[0].Context)
	}
}

func TestList_paginates_and_stops_atSince(t *testing.T) {
	// Given: 1ページ目は次ページを持ち、2ページ目は since より古い
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			_, _ = io.WriteString(w, `{"tweets":[{"id":"new","createdAt":"Wed Aug 19 10:00:00 +0000 2026"}],"has_next_page":true,"next_cursor":"next","status":"success"}`)
			return
		}
		_, _ = io.WriteString(w, `{"tweets":[{"id":"old","createdAt":"Tue Aug 18 10:00:00 +0000 2026"}],"has_next_page":false,"status":"success"}`)
	}))
	t.Cleanup(server.Close)
	apiKeySecret := secrettransport.NewSecretRef()
	source := twitterapiio.NewPostSource(newTestSecretTransportClient(t, server, apiKeySecret), apiKeySecret)

	// When: since を指定して List を呼ぶ
	got, err := source.List(context.Background(), time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC))

	// Then: 2ページ目まで取得し、古い要素は含めない
	if err != nil || len(got) != 1 || calls != 2 {
		t.Fatalf("got = %+v, err = %v, calls = %d", got, err, calls)
	}
}

func TestList_returnsInfrastructureError_whenClientNil(t *testing.T) {
	// Given: client が nil
	source := twitterapiio.NewPostSource(nil, secrettransport.NewSecretRef())
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: Infrastructure Error。Error / Unwrap が観測できる
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	var infra *twitterapiio.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *twitterapiio.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "twitterapiio:") {
		t.Fatalf("Error() = %q, want prefix twitterapiio:", infra.Error())
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() is nil")
	}
}

func TestList_returnsError_whenVendorStatusIsError(t *testing.T) {
	// Given: watch user と vendor error response を返す upstream double
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"status":"error","message":"失敗"}`)
	}))
	t.Cleanup(server.Close)
	apiKeySecret := secrettransport.NewSecretRef()
	source := twitterapiio.NewPostSource(newTestSecretTransportClient(t, server, apiKeySecret), apiKeySecret)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC))

	// Then: 部分結果なしで error
	if got != nil || err == nil {
		t.Fatalf("got = %+v, err = %v, want nil and error", got, err)
	}
}
