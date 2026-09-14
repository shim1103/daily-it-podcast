package r2_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/r2"
)

const (
	testAccessKeyID     = "r2-test-access-key-id-real-value"
	testSecretAccessKey = "r2-test-secret-access-key-real-value"
	testAccountID       = "r2-test-account-id-real-value"
	testBucket          = "r2-test-bucket-real-value"
)

func testCredentials() r2.Credentials {
	return r2.Credentials{
		AccessKeyID:     testAccessKeyID,
		SecretAccessKey: testSecretAccessKey,
	}
}

func testEndpoint() r2.Endpoint {
	return r2.Endpoint{AccountID: testAccountID, Bucket: testBucket}
}

type stubCall struct {
	Method      string
	Path        string
	Host        string
	Body        string
	ContentType string
	Auth        string
}

type stubResp struct {
	Status int
	Body   string
	Err    error
}

// seqRoundTripper は境界 I/O なしで request を記録し、応答を先頭から消費する。
type seqRoundTripper struct {
	mu    sync.Mutex
	calls []stubCall
	resps []stubResp
}

func (rt *seqRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		_ = req.Body.Close()
	}
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.calls = append(rt.calls, stubCall{
		Method:      req.Method,
		Path:        req.URL.EscapedPath(),
		Host:        req.URL.Host,
		Body:        string(body),
		ContentType: req.Header.Get("Content-Type"),
		Auth:        req.Header.Get("Authorization"),
	})
	if len(rt.resps) == 0 {
		return nil, fmt.Errorf("seqRoundTripper: no response left for %s %s", req.Method, req.URL.String())
	}
	res := rt.resps[0]
	rt.resps = rt.resps[1:]
	if res.Err != nil {
		return nil, res.Err
	}
	return &http.Response{
		StatusCode: res.Status,
		Body:       io.NopCloser(bytes.NewReader([]byte(res.Body))),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func newStubWriter(rt *seqRoundTripper) *r2.EpisodeWriter {
	return r2.NewEpisodeWriter(
		&http.Client{Transport: rt},
		testCredentials(),
		testEndpoint(),
	)
}

func assertNoSecretLeak(t *testing.T, msg string) {
	t.Helper()
	leaks := []string{testAccessKeyID, testSecretAccessKey, testAccountID, testBucket}
	for _, leak := range leaks {
		if strings.Contains(msg, leak) {
			t.Fatalf("Error() が secret/resource 実値を含む: %q", msg)
		}
	}
}

func TestWrite_retriesOnce_whenPutReturns5xxThenSucceeds(t *testing.T) {
	t.Parallel()

	// Given: json が 500→200、mp3 が 200
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusInternalServerError, Body: "fail"},
		{Status: http.StatusOK},
		{Status: http.StatusOK},
	}}
	w := newStubWriter(rt)

	// When
	err := w.Write(context.Background(), "ep-r5", []byte(`{}`), models.SpeechAudio{Content: []byte("a")})

	// Then: 有限 retry 後に成功。json path が 2 回
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if len(rt.calls) != 3 {
		t.Fatalf("calls = %d, want 3", len(rt.calls))
	}
	if !strings.HasSuffix(rt.calls[0].Path, "/ep-r5.json") || !strings.HasSuffix(rt.calls[1].Path, "/ep-r5.json") {
		t.Fatalf("retry paths unexpected: %q, %q", rt.calls[0].Path, rt.calls[1].Path)
	}
	if !strings.HasSuffix(rt.calls[2].Path, "/ep-r5.mp3") {
		t.Fatalf("mp3 path = %q", rt.calls[2].Path)
	}
}

func TestWrite_retriesOnce_whenPutReturns429ThenSucceeds(t *testing.T) {
	t.Parallel()

	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusTooManyRequests},
		{Status: http.StatusOK},
		{Status: http.StatusOK},
	}}
	w := newStubWriter(rt)

	err := w.Write(context.Background(), "ep-r429", []byte(`{}`), models.SpeechAudio{Content: []byte("a")})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if len(rt.calls) != 3 {
		t.Fatalf("calls = %d, want 3", len(rt.calls))
	}
}

func TestWrite_retriesOnce_whenNetworkFailsThenSucceeds(t *testing.T) {
	t.Parallel()

	rt := &seqRoundTripper{resps: []stubResp{
		{Err: errors.New("connection reset")},
		{Status: http.StatusOK},
		{Status: http.StatusOK},
	}}
	w := newStubWriter(rt)

	err := w.Write(context.Background(), "ep-net", []byte(`{}`), models.SpeechAudio{Content: []byte("a")})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if len(rt.calls) != 3 {
		t.Fatalf("calls = %d, want 3", len(rt.calls))
	}
}

func TestWrite_failsFastWithoutRetry_whenPutReturns4xxOtherThan429(t *testing.T) {
	t.Parallel()

	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusForbidden, Body: "denied"},
		{Status: http.StatusOK}, // 到達しない想定
	}}
	w := newStubWriter(rt)

	err := w.Write(context.Background(), "ep-4xx", []byte(`{}`), models.SpeechAudio{Content: []byte("a")})
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "r2:") {
		t.Fatalf("Error() = %q", infra.Error())
	}
	assertNoSecretLeak(t, err.Error())
	if strings.Contains(err.Error(), "ep-4xx") {
		t.Fatalf("Error() に object key 断片が含まれる: %q", err.Error())
	}
	if len(rt.calls) != 1 {
		t.Fatalf("calls = %d, want 1 (fail-fast)", len(rt.calls))
	}
}

func TestWrite_returnsInfrastructureErrorWithoutSecrets_when5xxExhaustsRetries(t *testing.T) {
	t.Parallel()

	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusBadGateway},
		{Status: http.StatusBadGateway},
	}}
	w := newStubWriter(rt)

	err := w.Write(context.Background(), "ep-ex", []byte(`{}`), models.SpeechAudio{Content: []byte("a")})
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	assertNoSecretLeak(t, err.Error())
	if strings.Contains(err.Error(), "ep-ex") {
		t.Fatalf("Error() に object key 断片が含まれる: %q", err.Error())
	}
	if len(rt.calls) != 2 {
		t.Fatalf("calls = %d, want 2 (finite retry)", len(rt.calls))
	}
}

func TestWrite_returnsInfrastructureError_whenClientNil(t *testing.T) {
	t.Parallel()

	w := r2.NewEpisodeWriter(nil, testCredentials(), testEndpoint())
	err := w.Write(context.Background(), "ep-1", []byte(`{}`), models.SpeechAudio{Content: []byte("a")})
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	assertNoSecretLeak(t, err.Error())
}

func TestWrite_returnsInfrastructureError_whenWriterNil(t *testing.T) {
	t.Parallel()

	var w *r2.EpisodeWriter
	err := w.Write(context.Background(), "ep-1", []byte(`{}`), models.SpeechAudio{Content: []byte("a")})
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "r2:") {
		t.Fatalf("Error() = %q", infra.Error())
	}
}
