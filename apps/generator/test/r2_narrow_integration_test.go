// Scope: Narrow Integration
// 実物境界: r2.EpisodeWriter / r2.CompletedEpisodeLookup が標準 *http.Client で送信する外向き HTTPS PutObject / ListObjectsV2 / GetObject（test upstream server）
// Double: 本番 R2 peer は使わない。DialTLSContext で Account ID 由来 host 宛先だけを test server へ redirect する（httptest 経路。experimental local S3 gate は C1 依存で別 Verification）。
// @require dummy credential / Account ID / bucket を Adapter へ直接渡す。upstream は controllable な test server。
// @ensure path-style Host / フル path・SigV4 Authorization と、成功系代表（Writer: json→mp3 順・MIME・upsert。Lookup: List→Get→date 一致判定）を観測できる。
// @ensure 5xx / network の有限 retry と、その他 4xx の fail-fast を実 *http.Client + TLS double 経由で観測できる。
// @invariant error message に bucket・key・Account ID・Access Key・secret 実値を含めない。
package test

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/r2"
)

const (
	r2NarrowAccessKeyID     = "r2-narrow-access-key-id-real-value"
	r2NarrowSecretAccessKey = "r2-narrow-secret-access-key-real-value"
	r2NarrowAccountID       = "r2-narrow-account-id-real-value"
	r2NarrowBucket          = "r2-narrow-bucket-real-value"
)

type r2NarrowCall struct {
	Method      string
	Path        string
	Host        string
	Body        string
	ContentType string
	Auth        string
}

func newR2WriterWithProxy(t *testing.T, handler http.HandlerFunc) (*r2.EpisodeWriter, *[]r2NarrowCall) {
	t.Helper()
	return newR2WriterWithProxyDial(t, handler, nil)
}

// newR2WriterWithProxyDial は DialTLSContext を差し替え可能にする。
// dial が nil なら test TLS server へ常時接続する。
func newR2WriterWithProxyDial(t *testing.T, handler http.HandlerFunc, dial func(ctx context.Context, network, addr string) (net.Conn, error)) (*r2.EpisodeWriter, *[]r2NarrowCall) {
	t.Helper()
	calls := &[]r2NarrowCall{}
	var mu sync.Mutex
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read upstream body: %v", err)
		}
		mu.Lock()
		*calls = append(*calls, r2NarrowCall{
			Method:      r.Method,
			Path:        r.URL.EscapedPath(),
			Host:        r.Host,
			Body:        string(body),
			ContentType: r.Header.Get("Content-Type"),
			Auth:        r.Header.Get("Authorization"),
		})
		mu.Unlock()
		r.Body = io.NopCloser(bytes.NewReader(body))
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	if dial == nil {
		dial = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return tls.Dial(network, server.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // why: test server 自己署名を信頼する。
		}
	} else {
		inner := dial
		dial = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return inner(ctx, network, server.Listener.Addr().String())
		}
	}
	httpClient := &http.Client{
		Transport: &http.Transport{DialTLSContext: dial},
	}
	w := r2.NewEpisodeWriter(
		httpClient,
		r2NarrowAccessKeyID,
		r2NarrowSecretAccessKey,
		r2NarrowAccountID,
		r2NarrowBucket,
	)
	return w, calls
}

// newR2LookupWithProxy は httptest TLS double を刺した r2.CompletedEpisodeLookup を返す。
// Writer 用 newR2WriterWithProxy と同型の DialTLSContext redirect を使う。
func newR2LookupWithProxy(t *testing.T, handler http.HandlerFunc) (*r2.CompletedEpisodeLookup, *[]r2NarrowCall) {
	t.Helper()
	calls := &[]r2NarrowCall{}
	var mu sync.Mutex
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read upstream body: %v", err)
		}
		mu.Lock()
		*calls = append(*calls, r2NarrowCall{
			Method: r.Method,
			Path:   r.URL.EscapedPath(),
			Host:   r.Host,
			Auth:   r.Header.Get("Authorization"),
		})
		mu.Unlock()
		r.Body = io.NopCloser(bytes.NewReader(body))
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return tls.Dial(network, server.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // why: test server 自己署名を信頼する。
	}
	httpClient := &http.Client{
		Transport: &http.Transport{DialTLSContext: dial},
	}
	l := r2.NewCompletedEpisodeLookup(
		httpClient,
		r2NarrowAccessKeyID,
		r2NarrowSecretAccessKey,
		r2NarrowAccountID,
		r2NarrowBucket,
	)
	return l, calls
}

func assertR2NarrowNoSecretLeak(t *testing.T, msg string) {
	t.Helper()
	leaks := []string{r2NarrowAccessKeyID, r2NarrowSecretAccessKey, r2NarrowAccountID, r2NarrowBucket}
	for _, leak := range leaks {
		if strings.Contains(msg, leak) {
			t.Fatalf("Error() が secret/resource 実値を含む: %q", msg)
		}
	}
}

func TestR2EpisodeWriter_putsJSONThenMP3WithMIMEAndUpsert_whenUpstreamSucceeds(t *testing.T) {
	// Given: 常に 200 の S3 互換 double
	writer, calls := newR2WriterWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ms := []byte(`{"episodeId":"narrow-ep-1"}`)
	audio := models.SpeechAudio{Content: []byte("narrow-mp3-body")}

	// When: 1 回目 Write のあと、同 episodeId で upsert Write
	if err := writer.Write(context.Background(), "narrow-ep-1", ms, audio); err != nil {
		t.Fatalf("Write#1: %v", err)
	}
	audio2 := models.SpeechAudio{Content: []byte("narrow-mp3-body-v2")}
	if err := writer.Write(context.Background(), "narrow-ep-1", []byte(`{"episodeId":"narrow-ep-1","n":2}`), audio2); err != nil {
		t.Fatalf("Write#2: %v", err)
	}

	// Then: json→mp3 順・MIME・Host・SigV4・同 key upsert
	if len(*calls) != 4 {
		t.Fatalf("calls = %d, want 4", len(*calls))
	}
	wantHost := r2NarrowAccountID + ".r2.cloudflarestorage.com"
	wantJSONPath := "/" + r2NarrowBucket + "/narrow-ep-1.json"
	wantMP3Path := "/" + r2NarrowBucket + "/narrow-ep-1.mp3"
	for i, c := range *calls {
		if c.Method != http.MethodPut {
			t.Fatalf("call[%d] method = %q, want PUT", i, c.Method)
		}
		if c.Host != wantHost {
			t.Fatalf("call[%d] host = %q, want %q", i, c.Host, wantHost)
		}
		if !strings.HasPrefix(c.Auth, "AWS4-HMAC-SHA256 ") {
			t.Fatalf("call[%d] Authorization missing SigV4", i)
		}
	}
	if (*calls)[0].Path != wantJSONPath || (*calls)[0].ContentType != "application/json" {
		t.Fatalf("json#1 path/MIME = %q / %q", (*calls)[0].Path, (*calls)[0].ContentType)
	}
	if (*calls)[1].Path != wantMP3Path || (*calls)[1].ContentType != "audio/mpeg" {
		t.Fatalf("mp3#1 path/MIME = %q / %q", (*calls)[1].Path, (*calls)[1].ContentType)
	}
	if (*calls)[1].Body != string(audio.Content) {
		t.Fatalf("mp3#1 body mismatch")
	}
	if (*calls)[2].Path != wantJSONPath || (*calls)[3].Path != wantMP3Path {
		t.Fatalf("upsert paths = %q, %q", (*calls)[2].Path, (*calls)[3].Path)
	}
	if (*calls)[3].Body != string(audio2.Content) {
		t.Fatalf("mp3 upsert body mismatch")
	}
}

func TestR2EpisodeWriter_returnsInfrastructureErrorAfterFiniteRetry_whenUpstreamKeeps5xx(t *testing.T) {
	// Given: PutObject が常に 502 を返す S3 互換 TLS double
	writer, calls := newR2WriterWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	})
	audio := models.SpeechAudio{Content: []byte("narrow-mp3-5xx")}

	// When: Write する
	err := writer.Write(context.Background(), "narrow-ep-5xx", []byte(`{"episodeId":"narrow-ep-5xx"}`), audio)

	// Then: *adaptererror.Error（r2: prefix）かつ json put が 2 回（有限 retry once）
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "r2:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "r2:")
	}
	assertR2NarrowNoSecretLeak(t, err.Error())
	if strings.Contains(err.Error(), "narrow-ep-5xx") {
		t.Fatalf("Error() に object key 断片が含まれる: %q", err.Error())
	}
	if len(*calls) != 2 {
		t.Fatalf("upstream received %d requests, want 2 (retry once on 5xx)", len(*calls))
	}
	wantJSONPath := "/" + r2NarrowBucket + "/narrow-ep-5xx.json"
	for i, c := range *calls {
		if c.Method != http.MethodPut || c.Path != wantJSONPath {
			t.Fatalf("call[%d] method/path = %q %q, want PUT %q", i, c.Method, c.Path, wantJSONPath)
		}
	}
}

func TestR2EpisodeWriter_retriesOnceThenSucceeds_whenDialFailsThenRecovers(t *testing.T) {
	// Given: 初回 DialTLS だけ失敗し、以降は TLS double へ到達する
	var dials atomic.Int32
	writer, calls := newR2WriterWithProxyDial(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, func(ctx context.Context, network, addr string) (net.Conn, error) {
		if dials.Add(1) == 1 {
			return nil, errors.New("connection reset")
		}
		return tls.Dial(network, addr, &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // why: test server 自己署名を信頼する。
	})
	audio := models.SpeechAudio{Content: []byte("narrow-mp3-net")}

	// When: Write する
	err := writer.Write(context.Background(), "narrow-ep-net", []byte(`{"episodeId":"narrow-ep-net"}`), audio)

	// Then: network 有限 retry 後に成功。upstream は json+mp3 の 2 PUT を受ける
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if dials.Load() < 2 {
		t.Fatalf("dials = %d, want >= 2 (retry after network fail)", dials.Load())
	}
	if len(*calls) != 2 {
		t.Fatalf("upstream received %d requests, want 2", len(*calls))
	}
	wantJSONPath := "/" + r2NarrowBucket + "/narrow-ep-net.json"
	wantMP3Path := "/" + r2NarrowBucket + "/narrow-ep-net.mp3"
	if (*calls)[0].Path != wantJSONPath || (*calls)[1].Path != wantMP3Path {
		t.Fatalf("paths = %q, %q", (*calls)[0].Path, (*calls)[1].Path)
	}
}

func TestR2EpisodeWriter_failsFastWithoutRetry_whenUpstreamReturns4xx(t *testing.T) {
	// Given: PutObject が常に 403 を返す S3 互換 TLS double（429 以外の 4xx）
	writer, calls := newR2WriterWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "denied", http.StatusForbidden)
	})
	audio := models.SpeechAudio{Content: []byte("narrow-mp3-4xx")}

	// When: Write する
	err := writer.Write(context.Background(), "narrow-ep-4xx", []byte(`{"episodeId":"narrow-ep-4xx"}`), audio)

	// Then: fail-fast で 1 回だけ PUT。*adaptererror.Error かつ secret 非露出
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "r2:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "r2:")
	}
	assertR2NarrowNoSecretLeak(t, err.Error())
	if strings.Contains(err.Error(), "narrow-ep-4xx") {
		t.Fatalf("Error() に object key 断片が含まれる: %q", err.Error())
	}
	if len(*calls) != 1 {
		t.Fatalf("upstream received %d requests, want 1 (fail-fast on 4xx)", len(*calls))
	}
}

func r2NarrowListObjectsV2XML(keys ...string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">`)
	for _, k := range keys {
		b.WriteString("<Contents><Key>")
		b.WriteString(k)
		b.WriteString("</Key></Contents>")
	}
	b.WriteString(`<IsTruncated>false</IsTruncated>`)
	b.WriteString(`</ListBucketResult>`)
	return b.String()
}

func TestR2CompletedEpisodeLookup_returnsTrue_whenUpstreamHasMatchingPair(t *testing.T) {
	// Given: List が同 stem の json+mp3 を返し、Get した json の date が照会日と一致する S3 互換 double
	const stem = "narrow-lookup-pair"
	lookup, calls := newR2LookupWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("list-type") == "2" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(r2NarrowListObjectsV2XML(stem+".json", stem+".mp3")))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"date":"2026-08-31","episodeId":"` + stem + `"}`))
	})

	// When: 一致 date で照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: true。List→Get の順で host / SigV4 が観測できる
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if !got {
		t.Fatal("HasPair = false, want true")
	}
	if len(*calls) != 2 {
		t.Fatalf("upstream received %d requests, want 2 (list + get)", len(*calls))
	}
	wantHost := r2NarrowAccountID + ".r2.cloudflarestorage.com"
	for i, c := range *calls {
		if c.Method != http.MethodGet {
			t.Fatalf("call[%d] method = %q, want GET", i, c.Method)
		}
		if c.Host != wantHost {
			t.Fatalf("call[%d] host = %q, want %q", i, c.Host, wantHost)
		}
		if !strings.HasPrefix(c.Auth, "AWS4-HMAC-SHA256 ") {
			t.Fatalf("call[%d] Authorization missing SigV4", i)
		}
	}
	wantGetPath := "/" + r2NarrowBucket + "/" + stem + ".json"
	if (*calls)[1].Path != wantGetPath {
		t.Fatalf("get path = %q, want %q", (*calls)[1].Path, wantGetPath)
	}
}

func TestR2CompletedEpisodeLookup_returnsInfrastructureErrorAfterFiniteRetry_whenUpstreamKeeps5xx(t *testing.T) {
	// Given: List が常に 502 を返す S3 互換 TLS double
	lookup, calls := newR2LookupWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	})

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: *adaptererror.Error（r2: prefix）かつ List が 2 回（有限 retry once）
	if err == nil {
		t.Fatal("expected error")
	}
	if got {
		t.Fatal("HasPair = true on error, want false")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "r2:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "r2:")
	}
	assertR2NarrowNoSecretLeak(t, err.Error())
	if len(*calls) != 2 {
		t.Fatalf("upstream received %d requests, want 2 (retry once on 5xx)", len(*calls))
	}
}

func TestR2CompletedEpisodeLookup_failsFastWithoutRetry_whenUpstreamReturns4xx(t *testing.T) {
	// Given: List が常に 403 を返す S3 互換 TLS double（429 以外の 4xx）
	lookup, calls := newR2LookupWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "denied", http.StatusForbidden)
	})

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: fail-fast で 1 回だけ List。*adaptererror.Error かつ secret 非露出
	if err == nil {
		t.Fatal("expected error")
	}
	if got {
		t.Fatal("HasPair = true on error, want false")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "r2:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "r2:")
	}
	assertR2NarrowNoSecretLeak(t, err.Error())
	if len(*calls) != 1 {
		t.Fatalf("upstream received %d requests, want 1 (fail-fast on 4xx)", len(*calls))
	}
}
