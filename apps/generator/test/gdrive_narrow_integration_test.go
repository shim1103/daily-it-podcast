// Scope: Narrow Integration
// 実物境界: gdrive.EpisodeWriter が processenv.Client 経由で送信する外向き HTTP request（test upstream server）
// Double: BindingResolver は Composition と同型の in-memory map（narrowBindings）。TokenSource は stub。本番 credential は使わない。
// @require dummy process environment（t.Setenv）に Folder ID の実値をセットする。upstream は controllable な test server。DialTLSContext で本番 host 宛先だけを test server へ redirect する。
// @ensure list→create→upload の成功 call sequence を 1 連 upstream が受ける。json/wav の stem が一致する。
// @ensure Authorization header に TokenSource の token が Bearer で乗る。
// @invariant error message・assertion 失敗文言に dummy Folder ID の実値を含めない。
package test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/processenv"
)

const gdriveNarrowFolderIDSecretName = "GDRIVE_NARROW_TEST_FOLDER_ID"

// gdriveNarrowTokenSource は test 用の gdrive.TokenSource fake。
type gdriveNarrowTokenSource struct {
	token string
}

func (s gdriveNarrowTokenSource) Token(ctx context.Context) (string, error) {
	return s.token, nil
}

type gdriveNarrowCall struct {
	TargetURL     string
	Host          string
	Method        string
	Body          string
	Authorization string
}

// newGDriveWriterWithProxy は本番 host（www.googleapis.com）への接続を test TLS server へ差し替えた
// gdrive.EpisodeWriter を返す。
// why: Adapter は FilesURL / UploadURL を定数として持つため、DialTLSContext で接続先だけを test server へ redirect する。
func newGDriveWriterWithProxy(t *testing.T, token string, handler http.HandlerFunc) (*gdrive.EpisodeWriter, *[]gdriveNarrowCall) {
	t.Helper()
	calls := &[]gdriveNarrowCall{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read upstream body: %v", err)
		}
		*calls = append(*calls, gdriveNarrowCall{
			TargetURL:     r.URL.String(),
			Host:          r.Host,
			Method:        r.Method,
			Body:          string(body),
			Authorization: r.Header.Get("Authorization"),
		})
		r.Body = io.NopCloser(bytes.NewReader(body))
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	t.Setenv(gdriveNarrowFolderIDSecretName, "gdrive-narrow-folder-id-real-value")
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, server.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
	folderIDSecret := secrettransport.NewSecretRef()
	client := processenv.NewClient(narrowBindings{folderIDSecret: gdriveNarrowFolderIDSecretName}, httpClient, nil)
	return gdrive.NewRawEpisodeWriter(client, gdriveNarrowTokenSource{token: token}, folderIDSecret), calls
}

func gdriveNarrowWriteJSONStatus(t *testing.T, w http.ResponseWriter, status int, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
}

func gdriveNarrowSucceedHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.String()
		method := r.Method
		switch {
		case method == http.MethodGet && strings.Contains(target, "/drive/v3/files"):
			gdriveNarrowWriteJSONStatus(t, w, http.StatusOK, map[string]any{"files": []any{}})
		case method == http.MethodPost && strings.Contains(target, "/drive/v3/files"):
			var meta struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&meta); err != nil {
				t.Fatalf("decode create metadata: %v", err)
			}
			gdriveNarrowWriteJSONStatus(t, w, http.StatusOK, map[string]any{"id": "created-" + meta.Name})
		case method == http.MethodPatch && strings.Contains(target, "/upload/drive/v3/files/"):
			gdriveNarrowWriteJSONStatus(t, w, http.StatusOK, map[string]any{"id": "uploaded"})
		default:
			t.Fatalf("unexpected request method=%s url=%s", method, target)
		}
	}
}

func gdriveNarrowListFailHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.String()
		method := r.Method
		switch {
		case method == http.MethodGet && strings.Contains(target, "/drive/v3/files"):
			w.WriteHeader(http.StatusInternalServerError)
		default:
			t.Fatalf("unexpected request method=%s url=%s", method, target)
		}
	}
}

func TestGDriveEpisodeWriter_returnsErrorWithoutDummyValues_whenListReturns500(t *testing.T) {
	// Given: token stub と、Drive の files.list が 500 を返す test upstream
	const episodeID = "narrow-ep-err-1"
	const token = "ya29.narrow-test-token-real-value"
	writer, _ := newGDriveWriterWithProxy(t, token, gdriveNarrowListFailHandler(t))
	audio := models.SpeechAudio{Content: []byte("RIFFWAV")}

	// When: Write する
	err := writer.Write(context.Background(), episodeID, []byte(`{"episodeId":"narrow-ep-err-1"}`), audio)

	// Then: error が返り、error message に folder ID・token の実値を含まない
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *gdrive.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gdrive.Error", err, err)
	}
	msg := err.Error()
	if strings.Contains(msg, "gdrive-narrow-folder-id-real-value") {
		t.Fatalf("Error() = %q, must not contain folder ID real value", msg)
	}
	if strings.Contains(msg, token) {
		t.Fatalf("Error() = %q, must not contain token real value", msg)
	}
}

func TestGDriveEpisodeWriter_sendsListCreateUploadSequenceWithMatchingStem_whenDriveSucceeds(t *testing.T) {
	// Given: token stub と、Drive が空一覧と create/upload 成功を返す test upstream
	const episodeID = "narrow-ep-1"
	const token = "ya29.narrow-test-token"
	writer, calls := newGDriveWriterWithProxy(t, token, gdriveNarrowSucceedHandler(t))
	audio := models.SpeechAudio{Content: []byte("RIFFWAV")}

	// When: 非空 manuscript と非空 WAV で Write する
	err := writer.Write(context.Background(), episodeID, []byte(`{"episodeId":"narrow-ep-1"}`), audio)

	// Then: list→create→upload の呼び出しを json/wav それぞれ観測し、stem が一致する
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	var listCount, createCount, uploadCount int
	var jsonName, wavName string
	for _, c := range *calls {
		u, parseErr := url.Parse(c.TargetURL)
		if parseErr != nil {
			t.Fatalf("parse target: %v", parseErr)
		}
		if c.Host != "www.googleapis.com" {
			t.Fatalf("host = %q, want www.googleapis.com", c.Host)
		}
		if c.Authorization != "Bearer "+token {
			t.Fatalf("Authorization = %q, want Bearer token", c.Authorization)
		}
		switch {
		case c.Method == http.MethodGet && u.Path == "/drive/v3/files":
			listCount++
		case c.Method == http.MethodPost && u.Path == "/drive/v3/files":
			createCount++
			var meta struct {
				Name    string   `json:"name"`
				Parents []string `json:"parents"`
			}
			if err := json.Unmarshal([]byte(c.Body), &meta); err != nil {
				t.Fatalf("unmarshal create metadata: %v", err)
			}
			if len(meta.Parents) != 1 || meta.Parents[0] == "" {
				t.Fatalf("parents length = %d, want 1 non-empty entry", len(meta.Parents))
			}
			switch {
			case strings.HasSuffix(meta.Name, ".json"):
				jsonName = meta.Name
			case strings.HasSuffix(meta.Name, ".wav"):
				wavName = meta.Name
			default:
				t.Fatalf("unexpected create name %q", meta.Name)
			}
		case c.Method == http.MethodPatch && strings.HasPrefix(u.Path, "/upload/drive/v3/files/"):
			uploadCount++
		default:
			t.Fatalf("unexpected call method=%s path=%s", c.Method, u.Path)
		}
	}

	if listCount != 2 {
		t.Fatalf("list calls = %d, want 2", listCount)
	}
	if createCount != 2 {
		t.Fatalf("create calls = %d, want 2", createCount)
	}
	if uploadCount != 2 {
		t.Fatalf("upload calls = %d, want 2", uploadCount)
	}
	if jsonName != episodeID+".json" {
		t.Fatalf("json name = %q, want %q", jsonName, episodeID+".json")
	}
	if wavName != episodeID+".wav" {
		t.Fatalf("wav name = %q, want %q", wavName, episodeID+".wav")
	}
	wantStem := episodeID
	jsonStem := strings.TrimSuffix(jsonName, ".json")
	wavStem := strings.TrimSuffix(wavName, ".wav")
	if jsonStem != wantStem || wavStem != wantStem {
		t.Fatalf("stem mismatch: json=%q wav=%q, want %q", jsonStem, wavStem, wantStem)
	}
}
