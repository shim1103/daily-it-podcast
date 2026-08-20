package gdrive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

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

type proxyCall struct {
	TargetURL     string
	Method        string
	Body          string
	BodyParents0  string
	ContentType   string
	Authorization string
}

type driveProbe struct {
	Calls []proxyCall
}

func newWriterWithProxy(t *testing.T, tokens TokenSource, handler http.HandlerFunc) (*EpisodeWriter, *driveProbe) {
	t.Helper()
	probe := &driveProbe{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read proxy body: %v", err)
		}
		probe.Calls = append(probe.Calls, proxyCall{
			TargetURL:     r.Header.Get("X-AS-Target-URL"),
			Method:        r.Header.Get("X-AS-Method"),
			Body:          string(body),
			BodyParents0:  r.Header.Get("X-AS-Inject-Body-parents-0"),
			ContentType:   r.Header.Get("Content-Type"),
			Authorization: r.Header.Get("Authorization"),
		})
		r.Body = io.NopCloser(bytes.NewReader(body))
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	client := &agentsecrets.Client{
		HTTP:     server.Client(),
		ProxyURL: server.URL,
	}
	return NewEpisodeWriter(client, tokens), probe
}

func writeJSONStatus(t *testing.T, w http.ResponseWriter, status int, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
}

func parseTarget(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse target: %v", err)
	}
	return u
}

func isListCall(c proxyCall) bool {
	u, err := url.Parse(c.TargetURL)
	if err != nil {
		return false
	}
	return c.Method == http.MethodGet && u.Host == "www.googleapis.com" && u.Path == "/drive/v3/files"
}

func isMetadataCreate(c proxyCall) bool {
	u, err := url.Parse(c.TargetURL)
	if err != nil {
		return false
	}
	return c.Method == http.MethodPost && u.Host == "www.googleapis.com" && u.Path == "/drive/v3/files"
}

func isMediaUpload(c proxyCall) bool {
	u, err := url.Parse(c.TargetURL)
	if err != nil {
		return false
	}
	return c.Method == http.MethodPatch &&
		u.Host == "www.googleapis.com" &&
		strings.HasPrefix(u.Path, "/upload/drive/v3/files/")
}

func succeedCreateHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("X-AS-Target-URL")
		method := r.Header.Get("X-AS-Method")
		switch {
		case method == http.MethodGet && strings.Contains(target, "/drive/v3/files"):
			writeJSONStatus(t, w, http.StatusOK, map[string]any{"files": []any{}})
		case method == http.MethodPost && strings.Contains(target, "/drive/v3/files"):
			var meta struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&meta); err != nil {
				t.Fatalf("decode create metadata: %v", err)
			}
			id := "created-" + meta.Name
			writeJSONStatus(t, w, http.StatusOK, map[string]any{"id": id})
		case method == http.MethodPatch && strings.Contains(target, "/upload/drive/v3/files/"):
			writeJSONStatus(t, w, http.StatusOK, map[string]any{"id": "uploaded"})
		default:
			t.Fatalf("unexpected request method=%s url=%s", method, target)
		}
	}
}

func TestWrite_uploadsJSONAndWAVWithSameStem_whenDriveSucceeds(t *testing.T) {
	t.Parallel()

	// Given: token stub と、Drive が空一覧と create 成功を返す stub
	const episodeID = "ep-1"
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, probe := newWriterWithProxy(t, tokens, succeedCreateHandler(t))
	audio := models.SpeechAudio{Content: []byte("RIFFWAV")}

	// When: 非空 manuscript と非空 WAV で Write する
	err := writer.Write(context.Background(), episodeID, []byte(`{"episodeId":"ep-1"}`), audio)

	// Then: 同一 stem の json と wav がフォルダ直下へ書かれる
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if len(probe.Calls) == 0 {
		t.Fatal("proxy was not called")
	}

	var jsonName, wavName string
	var bodyInjects int
	for _, c := range probe.Calls {
		if isListCall(c) {
			u := parseTarget(t, c.TargetURL)
			if got := u.Query().Get("fields"); got != "files(id)" {
				t.Fatalf("list fields = %q, want files(id)", got)
			}
			q := u.Query().Get("q")
			if strings.Contains(q, "in parents") {
				t.Fatalf("list q contains parents: %q", q)
			}
			if strings.Contains(q, secretnames.DriveFolderIDName) {
				t.Fatalf("list q contains folder key name: %q", q)
			}
		}
		if isMetadataCreate(c) {
			if c.BodyParents0 != secretnames.DriveFolderIDName {
				t.Fatalf("body inject parents.0 = %q, want %q", c.BodyParents0, secretnames.DriveFolderIDName)
			}
			bodyInjects++
			var meta struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal([]byte(c.Body), &meta); err != nil {
				t.Fatalf("create metadata: %v body=%s", err, c.Body)
			}
			switch {
			case strings.HasSuffix(meta.Name, ".json"):
				jsonName = meta.Name
			case strings.HasSuffix(meta.Name, ".wav"):
				wavName = meta.Name
			default:
				t.Fatalf("unexpected create name %q", meta.Name)
			}
		}
		if c.Authorization != "Bearer ya29.test-token" {
			t.Fatalf("Authorization = %q url=%s", c.Authorization, c.TargetURL)
		}
	}
	if bodyInjects != 2 {
		t.Fatalf("body inject count = %d, want 2", bodyInjects)
	}
	if jsonName != episodeID+".json" {
		t.Fatalf("json name = %q", jsonName)
	}
	if wavName != episodeID+".wav" {
		t.Fatalf("wav name = %q", wavName)
	}
}

func TestWrite_updatesExistingFiles_whenSameNameListed(t *testing.T) {
	t.Parallel()

	// Given: list が同一名の既存 file を返す stub
	const episodeID = "ep-1"
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, probe := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("X-AS-Target-URL")
		method := r.Header.Get("X-AS-Method")
		switch {
		case method == http.MethodGet && strings.Contains(target, "/drive/v3/files"):
			u := parseTarget(t, target)
			q := u.Query().Get("q")
			id := "existing-json"
			if strings.Contains(q, ".wav") {
				id = "existing-wav"
			}
			writeJSONStatus(t, w, http.StatusOK, map[string]any{
				"files": []map[string]any{{"id": id, "name": "listed"}},
			})
		case method == http.MethodPatch && strings.Contains(target, "/upload/drive/v3/files/"):
			writeJSONStatus(t, w, http.StatusOK, map[string]any{"id": "updated"})
		default:
			t.Fatalf("unexpected request method=%s url=%s", method, target)
		}
	})

	// When: 同一 episodeId で Write する
	err := writer.Write(context.Background(), episodeID, []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: update のみで create は呼ばない
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	var updates []string
	for _, c := range probe.Calls {
		if isMetadataCreate(c) {
			t.Fatalf("create was called: %s body=%s", c.TargetURL, c.Body)
		}
		if isMediaUpload(c) {
			updates = append(updates, parseTarget(t, c.TargetURL).Path)
		}
	}
	if len(updates) != 2 {
		t.Fatalf("media uploads = %#v, want 2", updates)
	}
	joined := strings.Join(updates, ",")
	if !strings.Contains(joined, "existing-json") || !strings.Contains(joined, "existing-wav") {
		t.Fatalf("upload paths = %#v", updates)
	}
}

func TestWrite_returnsDomainErrorWithoutHTTP_whenEpisodeIDEmpty(t *testing.T) {
	t.Parallel()

	// Given: episodeID が空
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, probe := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("proxy must not be called")
	})

	// When: 空 episodeID で Write する
	err := writer.Write(context.Background(), "", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: Domain Error。HTTP は 0 回
	if err == nil {
		t.Fatal("expected error")
	}
	var domain *domainerrors.EmptyEpisodeID
	if !errors.As(err, &domain) {
		t.Fatalf("error type %T (%v), want *errors.EmptyEpisodeID", err, err)
	}
	if domain.Error() == "" {
		t.Fatal("Error() is empty")
	}
	if len(probe.Calls) != 0 {
		t.Fatalf("unexpected requests: %+v", probe.Calls)
	}
}

func TestWrite_returnsDomainErrorWithoutHTTP_whenWAVEmpty(t *testing.T) {
	t.Parallel()

	// Given: WAV Content が空
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, probe := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("proxy must not be called")
	})

	// When: 空 WAV で Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{})

	// Then: Domain Error。HTTP は 0 回
	if err == nil {
		t.Fatal("expected error")
	}
	var domain *domainerrors.EmptyAudio
	if !errors.As(err, &domain) {
		t.Fatalf("error type %T (%v), want *errors.EmptyAudio", err, err)
	}
	if domain.Error() == "" {
		t.Fatal("Error() is empty")
	}
	if len(probe.Calls) != 0 {
		t.Fatalf("unexpected requests: %+v", probe.Calls)
	}
}

func TestWrite_returnsInfrastructureError_whenTokenSourceFails(t *testing.T) {
	t.Parallel()

	// Given: TokenSource が error を返す
	tokens := stubTokenSource{err: fmt.Errorf("refresh failed")}
	writer, probe := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("proxy must not be called")
	})

	// When: Write する
	err := writer.Write(context.Background(), "ep-1", []byte(`{"episodeId":"ep-1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: Infrastructure Error。HTTP は 0 回
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
	if len(probe.Calls) != 0 {
		t.Fatalf("unexpected requests: %+v", probe.Calls)
	}
}

func TestWrite_returnsInfrastructureErrorWithoutDelete_whenWAVUploadFailsAfterJSON(t *testing.T) {
	t.Parallel()

	// Given: json 書込は成功し wav media upload だけ 500
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, probe := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("X-AS-Target-URL")
		method := r.Header.Get("X-AS-Method")
		switch {
		case method == http.MethodGet && strings.Contains(target, "/drive/v3/files"):
			writeJSONStatus(t, w, http.StatusOK, map[string]any{"files": []any{}})
		case method == http.MethodPost && strings.Contains(target, "/drive/v3/files"):
			writeJSONStatus(t, w, http.StatusOK, map[string]any{"id": "created-id"})
		case method == http.MethodPatch && strings.Contains(target, "/upload/drive/v3/files/"):
			if strings.Contains(r.Header.Get("Content-Type"), "audio") {
				writeJSONStatus(t, w, http.StatusInternalServerError, map[string]any{"error": "upload failed"})
				return
			}
			writeJSONStatus(t, w, http.StatusOK, map[string]any{"id": "uploaded-json"})
		default:
			t.Fatalf("unexpected request method=%s url=%s", method, target)
		}
	})

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
	var createdJSON bool
	for _, c := range probe.Calls {
		if c.Method == http.MethodDelete {
			t.Fatalf("delete was called: %s", c.TargetURL)
		}
		if isMetadataCreate(c) {
			var meta struct {
				Name string `json:"name"`
			}
			_ = json.Unmarshal([]byte(c.Body), &meta)
			if meta.Name == "ep-1.json" {
				createdJSON = true
			}
		}
	}
	if !createdJSON {
		t.Fatal("json create was not observed")
	}
}

func TestWrite_returnsInfrastructureError_whenClientNil(t *testing.T) {
	t.Parallel()

	// Given: client が nil
	writer := NewEpisodeWriter(nil, stubTokenSource{token: "ya29.test-token"})

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
	t.Parallel()

	// Given: files.list が 500
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, _ := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		method := r.Header.Get("X-AS-Method")
		switch method {
		case http.MethodGet:
			writeJSONStatus(t, w, http.StatusInternalServerError, map[string]any{"error": "list failed"})
		default:
			t.Fatalf("unexpected request method=%s", method)
		}
	})

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
	t.Parallel()

	// Given: files.list の body が JSON でない
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, _ := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		method := r.Header.Get("X-AS-Method")
		switch method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not-json"))
		default:
			t.Fatalf("unexpected request method=%s", method)
		}
	})

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
	t.Parallel()

	// Given: metadata create が 500
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, _ := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		method := r.Header.Get("X-AS-Method")
		switch method {
		case http.MethodGet:
			writeJSONStatus(t, w, http.StatusOK, map[string]any{"files": []any{}})
		case http.MethodPost:
			writeJSONStatus(t, w, http.StatusInternalServerError, map[string]any{"error": "create failed"})
		default:
			t.Fatalf("unexpected request method=%s", method)
		}
	})

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
	t.Parallel()

	// Given: create 応答の id が空
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, _ := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		method := r.Header.Get("X-AS-Method")
		switch method {
		case http.MethodGet:
			writeJSONStatus(t, w, http.StatusOK, map[string]any{"files": []any{}})
		case http.MethodPost:
			writeJSONStatus(t, w, http.StatusOK, map[string]any{"id": ""})
		default:
			t.Fatalf("unexpected request method=%s", method)
		}
	})

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
	t.Parallel()

	// Given: create 応答が JSON でない
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, _ := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		method := r.Header.Get("X-AS-Method")
		switch method {
		case http.MethodGet:
			writeJSONStatus(t, w, http.StatusOK, map[string]any{"files": []any{}})
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("not-json"))
		default:
			t.Fatalf("unexpected request method=%s", method)
		}
	})

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

func TestWrite_escapesQuoteInListQuery_whenEpisodeIDContainsQuote(t *testing.T) {
	t.Parallel()

	// Given: episodeId に単引用符を含む
	const episodeID = "ep'1"
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, probe := newWriterWithProxy(t, tokens, succeedCreateHandler(t))

	// When: Write する
	err := writer.Write(context.Background(), episodeID, []byte(`{"episodeId":"ep'1"}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: list q の値が escape されている
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	var sawEscaped bool
	for _, c := range probe.Calls {
		if !isListCall(c) {
			continue
		}
		q := parseTarget(t, c.TargetURL).Query().Get("q")
		if strings.Contains(q, `name = 'ep\'1.json'`) || strings.Contains(q, `name = 'ep\'1.wav'`) {
			sawEscaped = true
		}
	}
	if !sawEscaped {
		t.Fatal("escaped quote was not observed in list q")
	}
}
