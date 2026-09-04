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
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

var _ port.EpisodeWriter = (*EpisodeWriter)(nil)

// TokenSource は Drive REST 呼び出しに使う access token を返す。
// OAuth refresh 等の取得手段は実装に閉じる。
type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

type EpisodeWriter struct {
	client   *http.Client
	tokens   TokenSource
	folderID string
}

// NewRawEpisodeWriter は validation 前の Google Drive 書込 Adapter を返す。
//
// @require httpClient != nil。tokens != nil。
// @ensure folderID は create metadata の Parents にだけ使い、保存元の知識は持たない。
func NewRawEpisodeWriter(httpClient *http.Client, tokens TokenSource, folderID string) *EpisodeWriter {
	return &EpisodeWriter{client: httpClient, tokens: tokens, folderID: folderID}
}

func (w *EpisodeWriter) Write(ctx context.Context, episodeID string, manuscript []byte, audio models.SpeechAudio) error {
	if w == nil || w.client == nil {
		return infraErr("write", fmt.Errorf("client is nil"))
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
	res, err := w.doDrive(ctx, http.MethodGet, FilesURL+"?"+q.Encode(), token, "", nil)
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
	// why: folder ID は Composition が渡した非 secret runtime config。metadata へ直接埋める。
	meta, err := json.Marshal(fileMetadata{
		Name:     name,
		MimeType: mime,
		Parents:  []string{w.folderID},
	})
	if err != nil {
		return "", infraErr("create_marshal", err)
	}
	res, err := w.doDrive(ctx, http.MethodPost, FilesURL, token, jsonMIME, meta)
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
	res, err := w.doDrive(ctx, http.MethodPatch, target, token, mime, content)
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

func (w *EpisodeWriter) doDrive(ctx context.Context, method, target, token, contentType string, body []byte) (*http.Response, error) {
	return doDriveHTTP(ctx, w.client, method, target, token, contentType, body)
}

func doDriveHTTP(ctx context.Context, client *http.Client, method, target, token, contentType string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(body))
	if err != nil {
		return nil, infraErr("build_request", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return client.Do(req)
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
