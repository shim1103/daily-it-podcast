package gdrive

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/processenv"
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

// stubBindings は test 用の BindingResolver fake。
type stubBindings map[secrettransport.SecretRef]string

func (b stubBindings) ResolveSecret(ref secrettransport.SecretRef) (string, bool) {
	name, ok := b[ref]
	return name, ok
}

const driveTestFolderIDSecretName = "GDRIVE_TEST_FOLDER_ID"

type driveCall struct {
	TargetURL     string
	Host          string
	Method        string
	Body          string
	ContentType   string
	Authorization string
}

type driveProbe struct {
	Calls []driveCall
}

// newWriterWithProxy は本番 host（www.googleapis.com）への接続を test TLS server へ差し替えた EpisodeWriter を返す。
// why: Adapter は FilesURL / UploadURL を定数として持つため、DialTLSContext で接続先だけを test server へ redirect する。
func newWriterWithProxy(t *testing.T, tokens TokenSource, handler http.HandlerFunc) (*EpisodeWriter, *driveProbe) {
	t.Helper()
	probe := &driveProbe{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read upstream body: %v", err)
		}
		probe.Calls = append(probe.Calls, driveCall{
			TargetURL:     r.URL.String(),
			Host:          r.Host,
			Method:        r.Method,
			Body:          string(body),
			ContentType:   r.Header.Get("Content-Type"),
			Authorization: r.Header.Get("Authorization"),
		})
		r.Body = io.NopCloser(bytes.NewReader(body))
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	t.Setenv(driveTestFolderIDSecretName, "gdrive-test-folder-id-real-value")
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, server.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
	folderIDSecret := secrettransport.NewSecretRef()
	client := processenv.NewClient(stubBindings{folderIDSecret: driveTestFolderIDSecretName}, httpClient, nil)
	return NewRawEpisodeWriter(client, tokens, folderIDSecret), probe
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

func isListCall(c driveCall) bool {
	u, err := url.Parse(c.TargetURL)
	if err != nil {
		return false
	}
	return c.Method == http.MethodGet && c.Host == "www.googleapis.com" && u.Path == "/drive/v3/files"
}

func isMetadataCreate(c driveCall) bool {
	u, err := url.Parse(c.TargetURL)
	if err != nil {
		return false
	}
	return c.Method == http.MethodPost && c.Host == "www.googleapis.com" && u.Path == "/drive/v3/files"
}

func isMediaUpload(c driveCall) bool {
	u, err := url.Parse(c.TargetURL)
	if err != nil {
		return false
	}
	return c.Method == http.MethodPatch &&
		c.Host == "www.googleapis.com" &&
		strings.HasPrefix(u.Path, "/upload/drive/v3/files/")
}

func succeedCreateHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.String()
		method := r.Method
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
		t.Fatal("upstream was not called")
	}

	var jsonName, wavName string
	var parentInjected int
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
		}
		if isMetadataCreate(c) {
			var meta struct {
				Name    string   `json:"name"`
				Parents []string `json:"parents"`
			}
			if err := json.Unmarshal([]byte(c.Body), &meta); err != nil {
				t.Fatalf("create metadata: %v body=%s", err, c.Body)
			}
			if len(meta.Parents) != 1 || meta.Parents[0] != "gdrive-test-folder-id-real-value" {
				t.Fatalf("parents = %#v, want folder id real value", meta.Parents)
			}
			parentInjected++
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
	if parentInjected != 2 {
		t.Fatalf("parent inject count = %d, want 2", parentInjected)
	}
	if jsonName != episodeID+".json" {
		t.Fatalf("json name = %q", jsonName)
	}
	if wavName != episodeID+".wav" {
		t.Fatalf("wav name = %q", wavName)
	}
}

func TestWrite_updatesExistingFiles_whenSameNameListed(t *testing.T) {

	// Given: list が同一名の既存 file を返す stub
	const episodeID = "ep-1"
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, probe := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.String()
		method := r.Method
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

func TestWrite_returnsInfrastructureError_whenTokenSourceFails(t *testing.T) {

	// Given: TokenSource が error を返す
	tokens := stubTokenSource{err: fmt.Errorf("refresh failed")}
	writer, probe := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("upstream must not be called")
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

	// Given: json 書込は成功し wav media upload だけ 500
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, probe := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.String()
		method := r.Method
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
	writer := NewRawEpisodeWriter(nil, stubTokenSource{token: "ya29.test-token"}, secrettransport.NewSecretRef())

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
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, _ := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
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

	// Given: files.list の body が JSON でない
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, _ := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
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

	// Given: metadata create が 500
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, _ := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
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

	// Given: create 応答の id が空
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, _ := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
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

	// Given: create 応答が JSON でない
	tokens := stubTokenSource{token: "ya29.test-token"}
	writer, _ := newWriterWithProxy(t, tokens, func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
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
