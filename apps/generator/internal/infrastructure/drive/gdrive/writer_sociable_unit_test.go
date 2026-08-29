package gdrive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

const testFolderID = "gdrive-test-folder-id"

// stubTokenSource は test 用の TokenSource fake。
type stubTokenSource struct {
	token string
	err   error
}

func (s stubTokenSource) Token(ctx context.Context) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.token, nil
}

// stubClientCall は stubRoundTripper が受けた1回分の request を、境界 I/O なしで記録する。
type stubClientCall struct {
	Method      string
	TargetURL   string
	Body        string
	ContentType string
}

// stubClientResponse は method・target・Content-Type 部分一致に応じて返す応答、または error を表す。
type stubClientResponse struct {
	MatchMethod      string
	MatchPath        string // TargetURL の部分文字列。空なら method のみで一致。
	MatchContentType string // Content-Type 部分文字列。空なら不問。
	Status           int
	Body             string
	Err              error
}

// stubRoundTripper は http.RoundTripper を境界 I/O なしで満たす直接 Stub。
// responses の先頭から順に、method・target が一致する最初の1件を使う。
type stubRoundTripper struct {
	responses []stubClientResponse
	calls     []stubClientCall
}

func (c *stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		_ = req.Body.Close()
	}
	contentType := req.Header.Get("Content-Type")
	c.calls = append(c.calls, stubClientCall{
		Method:      req.Method,
		TargetURL:   req.URL.String(),
		Body:        string(body),
		ContentType: contentType,
	})
	for _, res := range c.responses {
		if res.MatchMethod != "" && res.MatchMethod != req.Method {
			continue
		}
		if res.MatchPath != "" && !strings.Contains(req.URL.String(), res.MatchPath) {
			continue
		}
		if res.MatchContentType != "" && !strings.Contains(contentType, res.MatchContentType) {
			continue
		}
		if res.Err != nil {
			return nil, res.Err
		}
		return &http.Response{
			StatusCode: res.Status,
			Body:       io.NopCloser(bytes.NewReader([]byte(res.Body))),
			Header:     make(http.Header),
		}, nil
	}
	return nil, fmt.Errorf("stubRoundTripper: no response configured for method=%s target=%s", req.Method, req.URL.String())
}

func newStubWriter(rt *stubRoundTripper, tokens TokenSource) *EpisodeWriter {
	return NewRawEpisodeWriter(&http.Client{Transport: rt}, tokens, testFolderID)
}

func jsonBody(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return string(raw)
}

func listCalls(calls []stubClientCall) []stubClientCall {
	var out []stubClientCall
	for _, c := range calls {
		if c.Method == http.MethodGet {
			out = append(out, c)
		}
	}
	return out
}

func createCalls(calls []stubClientCall) []stubClientCall {
	var out []stubClientCall
	for _, c := range calls {
		if c.Method == http.MethodPost {
			out = append(out, c)
		}
	}
	return out
}

func uploadCalls(calls []stubClientCall) []stubClientCall {
	var out []stubClientCall
	for _, c := range calls {
		if c.Method == http.MethodPatch {
			out = append(out, c)
		}
	}
	return out
}

func TestWrite_updatesExistingFiles_whenSameNameListed(t *testing.T) {

	// Given: list が既存 file の id を返す stub。create は呼ばれない前提のため応答を用意しない
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, Status: http.StatusOK, Body: jsonBody(t, map[string]any{
				"files": []map[string]any{{"id": "existing-id", "name": "listed"}},
			})},
			{MatchMethod: http.MethodPatch, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"id": "updated"})},
		},
	}
	writer := newStubWriter(rt, stubTokenSource{token: "ya29.test-token"})

	// When: 同一 episodeId で Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: update のみで create は呼ばない
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := len(createCalls(rt.calls)); got != 0 {
		t.Fatalf("create calls = %d, want 0", got)
	}
	if got := len(uploadCalls(rt.calls)); got != 2 {
		t.Fatalf("upload calls = %d, want 2", got)
	}
	for _, c := range uploadCalls(rt.calls) {
		if !strings.Contains(c.TargetURL, "existing-id") {
			t.Fatalf("upload target = %q, want existing-id", c.TargetURL)
		}
	}
}

func TestWrite_returnsInfrastructureError_whenTokenSourceFails(t *testing.T) {

	// Given: TokenSource が error を返す
	rt := &stubRoundTripper{}
	writer := newStubWriter(rt, stubTokenSource{err: fmt.Errorf("refresh failed")})

	// When: Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: Infrastructure Error。Client は 0 回
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gdrive.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "gdrive:") {
		t.Fatalf("Error() = %q", infra.Error())
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() is nil")
	}
	if len(rt.calls) != 0 {
		t.Fatalf("unexpected client calls: %+v", rt.calls)
	}
}

func TestWrite_returnsInfrastructureErrorWithoutDelete_whenWAVUploadFailsAfterJSON(t *testing.T) {

	// Given: json 書込は成功し wav media upload だけ 500 を返す stub
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"files": []any{}})},
			{MatchMethod: http.MethodPost, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"id": "created-id"})},
			{MatchMethod: http.MethodPatch, MatchContentType: "audio", Status: http.StatusInternalServerError, Body: jsonBody(t, map[string]any{"error": "upload failed"})},
			{MatchMethod: http.MethodPatch, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"id": "uploaded-json"})},
		},
	}
	writer := newStubWriter(rt, stubTokenSource{token: "ya29.test-token"})

	// When: Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: 全体は Infrastructure Error。DELETE は呼ばない
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gdrive.Error", err, err)
	}
	for _, c := range rt.calls {
		if c.Method == http.MethodDelete {
			t.Fatalf("delete was called: %s", c.TargetURL)
		}
	}
	var createdJSON bool
	for _, c := range createCalls(rt.calls) {
		var meta struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal([]byte(c.Body), &meta)
		if meta.Name == "ep-1.json" {
			createdJSON = true
		}
	}
	if !createdJSON {
		t.Fatal("json create was not observed")
	}
}

func TestWrite_returnsInfrastructureError_whenClientNil(t *testing.T) {
	t.Parallel()

	// Given: client が nil
	writer := NewRawEpisodeWriter(nil, stubTokenSource{token: "ya29.test-token"}, testFolderID)

	// When: Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gdrive.Error", err, err)
	}
}

func TestWrite_returnsInfrastructureError_whenWriterNil(t *testing.T) {
	t.Parallel()

	// Given: receiver が nil
	var writer *EpisodeWriter

	// When: Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gdrive.Error", err, err)
	}
}

func TestWrite_returnsInfrastructureError_whenListHTTPFails(t *testing.T) {

	// Given: files.list が 500
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, Status: http.StatusInternalServerError, Body: jsonBody(t, map[string]any{"error": "list failed"})},
		},
	}
	writer := newStubWriter(rt, stubTokenSource{token: "ya29.test-token"})

	// When: Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gdrive.Error", err, err)
	}
}

func TestWrite_returnsInfrastructureError_whenListBodyInvalid(t *testing.T) {

	// Given: files.list の body が JSON でない
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, Status: http.StatusOK, Body: "not-json"},
		},
	}
	writer := newStubWriter(rt, stubTokenSource{token: "ya29.test-token"})

	// When: Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gdrive.Error", err, err)
	}
}

func TestWrite_returnsInfrastructureError_whenCreateHTTPFails(t *testing.T) {

	// Given: metadata create が 500
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"files": []any{}})},
			{MatchMethod: http.MethodPost, Status: http.StatusInternalServerError, Body: jsonBody(t, map[string]any{"error": "create failed"})},
		},
	}
	writer := newStubWriter(rt, stubTokenSource{token: "ya29.test-token"})

	// When: Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gdrive.Error", err, err)
	}
}

func TestWrite_returnsInfrastructureError_whenCreateIDEmpty(t *testing.T) {

	// Given: create 応答の id が空
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"files": []any{}})},
			{MatchMethod: http.MethodPost, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"id": ""})},
		},
	}
	writer := newStubWriter(rt, stubTokenSource{token: "ya29.test-token"})

	// When: Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gdrive.Error", err, err)
	}
}

func TestWrite_returnsInfrastructureError_whenCreateBodyInvalid(t *testing.T) {

	// Given: create 応答が JSON でない
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"files": []any{}})},
			{MatchMethod: http.MethodPost, Status: http.StatusCreated, Body: "not-json"},
		},
	}
	writer := newStubWriter(rt, stubTokenSource{token: "ya29.test-token"})

	// When: Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gdrive.Error", err, err)
	}
}

func TestWrite_embedsFolderIDInParents_whenCreatingMetadata(t *testing.T) {

	// Given: 空一覧と create/upload 成功を返す stub
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"files": []any{}})},
			{MatchMethod: http.MethodPost, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"id": "created-id"})},
			{MatchMethod: http.MethodPatch, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"id": "uploaded"})},
		},
	}
	writer := newStubWriter(rt, stubTokenSource{token: "ya29.test-token"})

	// When: Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: create metadata の Parents に folder ID が直接入る
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	for _, c := range createCalls(rt.calls) {
		var meta struct {
			Parents []string `json:"parents"`
		}
		if err := json.Unmarshal([]byte(c.Body), &meta); err != nil {
			t.Fatalf("unmarshal create metadata: %v", err)
		}
		if len(meta.Parents) != 1 || meta.Parents[0] != testFolderID {
			t.Fatalf("parents = %#v, want [%q]", meta.Parents, testFolderID)
		}
	}
}

func TestWrite_escapesQuoteInListQuery_whenEpisodeIDContainsQuote(t *testing.T) {

	// Given: episodeId に単引用符を含む
	const episodeID = "ep'1"
	rt := &stubRoundTripper{
		responses: []stubClientResponse{
			{MatchMethod: http.MethodGet, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"files": []any{}})},
			{MatchMethod: http.MethodPost, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"id": "created-id"})},
			{MatchMethod: http.MethodPatch, Status: http.StatusOK, Body: jsonBody(t, map[string]any{"id": "uploaded"})},
		},
	}
	writer := newStubWriter(rt, stubTokenSource{token: "ya29.test-token"})

	// When: Write する
	err := writer.Write(context.Background(), episodeID, []byte(`{"episodeId":"ep'1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: list q の値が escape されている
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	var sawEscaped bool
	for _, c := range listCalls(rt.calls) {
		u, parseErr := url.Parse(c.TargetURL)
		if parseErr != nil {
			t.Fatalf("parse target: %v", parseErr)
		}
		q := u.Query().Get("q")
		if strings.Contains(q, `name = 'ep\'1.json'`) || strings.Contains(q, `name = 'ep\'1.wav'`) {
			sawEscaped = true
		}
	}
	if !sawEscaped {
		t.Fatal("escaped quote was not observed in list q")
	}
}
