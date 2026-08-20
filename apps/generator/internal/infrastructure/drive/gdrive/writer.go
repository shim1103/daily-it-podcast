package gdrive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

var _ port.EpisodeWriter = (*EpisodeWriter)(nil)

// TokenSource は Drive REST 呼び出しに使う access token を返す。
// OAuth refresh 等の取得手段は実装に閉じる。
type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

type EpisodeWriter struct {
	client *agentsecrets.Client
	tokens TokenSource
}

// NewEpisodeWriter は Google Drive 書込 Adapter を返す。
//
// @require client != nil。tokens != nil。
// @ensure 秘密値は保持しない。
func NewEpisodeWriter(client *agentsecrets.Client, tokens TokenSource) *EpisodeWriter {
	return &EpisodeWriter{client: client, tokens: tokens}
}

func (w *EpisodeWriter) Write(ctx context.Context, episodeID string, manuscript []byte, audio models.SpeechAudio) error {
	if w == nil || w.client == nil {
		return infraErr("write", fmt.Errorf("client is nil"))
	}
	if strings.TrimSpace(episodeID) == "" {
		return &domainerrors.EmptyEpisodeID{}
	}
	if len(audio.Content) == 0 {
		return &domainerrors.EmptyAudio{}
	}

	token, err := w.tokens.Token(ctx)
	if err != nil {
		return infraErr("token", err)
	}
	if err := w.putFile(ctx, token, episodeID+jsonExt, jsonMIME, manuscript); err != nil {
		return err
	}
	if err := w.putFile(ctx, token, episodeID+wavExt, wavMIME, audio.Content); err != nil {
		return err
	}
	return nil
}

func (w *EpisodeWriter) putFile(ctx context.Context, token, name, mime string, content []byte) error {
	fileID, err := w.findFileID(ctx, token, name)
	if err != nil {
		return err
	}
	if fileID == "" {
		fileID, err = w.createMetadata(ctx, token, name, mime)
		if err != nil {
			return err
		}
	}
	return w.uploadMedia(ctx, token, fileID, mime, content)
}

func (w *EpisodeWriter) findFileID(ctx context.Context, token, name string) (string, error) {
	q := url.Values{}
	q.Set("q", "name = '"+escapeDriveQueryValue(name)+"' and trashed = false")
	q.Set("fields", "files(id)")
	res, err := w.doDrive(ctx, http.MethodGet, FilesURL+"?"+q.Encode(), token, "", nil, agentsecrets.Inject{})
	if err != nil {
		return "", infraErr("list", err)
	}
	var parsed struct {
		Files []struct {
			ID string `json:"id"`
		} `json:"files"`
	}
	if err := readJSONBody(res, "list", &parsed, http.StatusOK); err != nil {
		return "", err
	}
	if len(parsed.Files) == 0 {
		return "", nil
	}
	return parsed.Files[0].ID, nil
}

type fileMetadata struct {
	Name     string   `json:"name"`
	MimeType string   `json:"mimeType"`
	Parents  []string `json:"parents"`
}

func (w *EpisodeWriter) createMetadata(ctx context.Context, token, name, mime string) (string, error) {
	// why: parents の値は AgentSecrets の Body inject が埋める。code に folder ID を置かない。
	meta, err := json.Marshal(fileMetadata{
		Name:     name,
		MimeType: mime,
		Parents:  []string{""},
	})
	if err != nil {
		return "", infraErr("create_marshal", err)
	}
	res, err := w.doDrive(ctx, http.MethodPost, FilesURL, token, jsonMIME, bytes.NewReader(meta), agentsecrets.Inject{
		Body: map[string]string{"parents.0": secretnames.DriveFolderIDName},
	})
	if err != nil {
		return "", infraErr("create", err)
	}
	var parsed struct {
		ID string `json:"id"`
	}
	if err := readJSONBody(res, "create", &parsed, http.StatusOK, http.StatusCreated); err != nil {
		return "", err
	}
	if strings.TrimSpace(parsed.ID) == "" {
		return "", infraErr("create_decode", fmt.Errorf("file id is empty"))
	}
	return parsed.ID, nil
}

func (w *EpisodeWriter) uploadMedia(ctx context.Context, token, fileID, mime string, content []byte) error {
	target := UploadURL + "/" + url.PathEscape(fileID) + "?uploadType=media"
	res, err := w.doDrive(ctx, http.MethodPatch, target, token, mime, bytes.NewReader(content), agentsecrets.Inject{})
	if err != nil {
		return infraErr("upload", err)
	}
	defer func() { _ = res.Body.Close() }()
	if _, err := io.Copy(io.Discard, res.Body); err != nil {
		return infraErr("upload_read", err)
	}
	if res.StatusCode != http.StatusOK {
		return infraErr("upload_http", fmt.Errorf("status %d", res.StatusCode))
	}
	return nil
}

func (w *EpisodeWriter) doDrive(ctx context.Context, method, target, token, contentType string, body io.Reader, inject agentsecrets.Inject) (*http.Response, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	if contentType != "" {
		headers["Content-Type"] = contentType
	}
	return w.client.Do(ctx, agentsecrets.Request{
		Method:             method,
		TargetURL:          target,
		Body:               body,
		PassthroughHeaders: headers,
		Inject:             inject,
	})
}

func readJSONBody(res *http.Response, op string, dest any, okStatuses ...int) error {
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return infraErr(op+"_read", err)
	}
	ok := false
	for _, status := range okStatuses {
		if res.StatusCode == status {
			ok = true
			break
		}
	}
	if !ok {
		return infraErr(op+"_http", fmt.Errorf("status %d", res.StatusCode))
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return infraErr(op+"_decode", err)
	}
	return nil
}

func escapeDriveQueryValue(name string) string {
	name = strings.ReplaceAll(name, `\`, `\\`)
	return strings.ReplaceAll(name, `'`, `\'`)
}
