package httpget_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/httpget"
)

func TestGetWithRetry_returnsBody_whenFirstGetSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 1 回目で 200 と body "ok-body" を返す server
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		_, _ = io.WriteString(w, "ok-body")
	}))
	t.Cleanup(srv.Close)

	// When: GetWithRetry(ctx, client, url) を呼ぶ
	got, err := httpget.GetWithRetry(context.Background(), srv.Client(), srv.URL)

	// Then: body は "ok-body"、呼び出しは 1 回
	if err != nil {
		t.Fatalf("GetWithRetry() error = %v, want nil", err)
	}
	if string(got) != "ok-body" {
		t.Fatalf("body = %q, want %q", got, "ok-body")
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", calls.Load())
	}
}

func TestGetWithRetry_retriesOnce_whenStatus5xxThenSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 1 回目 503、2 回目 200 と body "recovered" を返す server
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = io.WriteString(w, "recovered")
	}))
	t.Cleanup(srv.Close)

	// When: GetWithRetry(ctx, client, url) を呼ぶ
	got, err := httpget.GetWithRetry(context.Background(), srv.Client(), srv.URL)

	// Then: body は "recovered"、呼び出しは 2 回（5xx で 1 回 retry）
	if err != nil {
		t.Fatalf("GetWithRetry() error = %v, want nil", err)
	}
	if string(got) != "recovered" {
		t.Fatalf("body = %q, want %q", got, "recovered")
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestGetWithRetry_stopsAfterOneRetry_whenStatus5xxPersists(t *testing.T) {
	t.Parallel()

	// Given: 常に 502 を返す server
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)

	// When: GetWithRetry(ctx, client, url) を呼ぶ
	got, err := httpget.GetWithRetry(context.Background(), srv.Client(), srv.URL)

	// Then: body は nil、error に 502、呼び出しは 2 回で打ち切り
	if got != nil {
		t.Fatalf("body = %q, want nil", got)
	}
	if err == nil {
		t.Fatal("error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Fatalf("error = %v, want status 502", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestGetWithRetry_doesNotRetry_whenStatus4xxIncluding429(t *testing.T) {
	t.Parallel()

	for _, status := range []int{http.StatusBadRequest, http.StatusNotFound, http.StatusTooManyRequests} {
		status := status
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			// Given: 指定 4xx（429 含む）を返す server
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				http.Error(w, "client error", status)
			}))
			t.Cleanup(srv.Close)

			// When: GetWithRetry(ctx, client, url) を呼ぶ
			got, err := httpget.GetWithRetry(context.Background(), srv.Client(), srv.URL)

			// Then: body は nil、error あり、呼び出しは 1 回（4xx は非 retry）
			if got != nil {
				t.Fatalf("body = %q, want nil", got)
			}
			if err == nil {
				t.Fatal("error = nil, want non-nil")
			}
			if calls.Load() != 1 {
				t.Fatalf("calls = %d, want 1 (no retry on 4xx)", calls.Load())
			}
		})
	}
}

func TestGetWithRetry_retriesOnce_whenTransportErrorThenSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 1 回目は transport error、2 回目は 200 と body "after-transport" を返す RoundTripper
	var calls atomic.Int32
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := calls.Add(1)
		if n == 1 {
			return nil, errors.New("connection reset")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("after-transport")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
	client := &http.Client{Transport: rt}

	// When: GetWithRetry(ctx, client, url) を呼ぶ
	got, err := httpget.GetWithRetry(context.Background(), client, "http://example.invalid/x")

	// Then: body は "after-transport"、呼び出しは 2 回
	if err != nil {
		t.Fatalf("GetWithRetry() error = %v, want nil", err)
	}
	if string(got) != "after-transport" {
		t.Fatalf("body = %q, want %q", got, "after-transport")
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestGetWithRetry_doesNotRetry_whenBodyReadFails(t *testing.T) {
	t.Parallel()

	// Given: 200 だが body 読み取りが失敗する RoundTripper
	var calls atomic.Int32
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(&errReader{err: errors.New("read failed")}),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
	client := &http.Client{Transport: rt}

	// When: GetWithRetry(ctx, client, url) を呼ぶ
	got, err := httpget.GetWithRetry(context.Background(), client, "http://example.invalid/x")

	// Then: body は nil、error あり、呼び出しは 1 回（body read 失敗は非 retry）
	if got != nil {
		t.Fatalf("body = %q, want nil", got)
	}
	if err == nil {
		t.Fatal("error = nil, want non-nil")
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1 (no retry on body read failure)", calls.Load())
	}
}

func TestGetWithRetry_returnsError_whenClientNil(t *testing.T) {
	t.Parallel()

	// Given: client が nil
	// When: GetWithRetry(ctx, nil, url) を呼ぶ
	got, err := httpget.GetWithRetry(context.Background(), nil, "http://example.invalid/x")

	// Then: body は nil、error あり
	if got != nil {
		t.Fatalf("body = %q, want nil", got)
	}
	if err == nil {
		t.Fatal("error = nil, want non-nil")
	}
}

func TestNormalizeHTML_unescapesEntitiesAndStripsTags_whenMarkupPresent(t *testing.T) {
	t.Parallel()

	// Given: <p>・他タグ・HTML entity を含む文字列
	raw := `最初の段落<p>次の段落 <a href="https://e.example">link</a> &amp; &lt;tag&gt;`

	// When: NormalizeHTML(raw) を呼ぶ
	got := httpget.NormalizeHTML(raw)

	// Then: 段落改行・タグ除去・entity unescape・trim した文字列
	want := "最初の段落\n次の段落 link & <tag>"
	if got != want {
		t.Fatalf("NormalizeHTML() = %q, want %q", got, want)
	}
}

func TestNormalizeHTML_returnsEmpty_whenInputEmpty(t *testing.T) {
	t.Parallel()

	// Given: 空文字
	raw := ""

	// When: NormalizeHTML("") を呼ぶ
	got := httpget.NormalizeHTML(raw)

	// Then: 空文字を返す
	if got != "" {
		t.Fatalf("NormalizeHTML(\"\") = %q, want empty", got)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type errReader struct {
	err error
}

func (r *errReader) Read([]byte) (int, error) {
	return 0, r.err
}
