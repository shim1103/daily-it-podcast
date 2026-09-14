package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
)

// mp3Uploader は TEST Drive folder へ mp3 1 本だけ書く（json は書かない）。
type mp3Uploader struct {
	http     *http.Client
	tokens   tokenSource
	folderID string
}

type tokenSource interface {
	Token(ctx context.Context) (string, error)
}

// PutMP3 は episodeID.mp3 を folder 直下へ create/upload する。
//
// @require episodeID・mp3 非空。Drive OAuth 済み。
// @ensure 同名があれば上書き upload。json は触らない。
func (u *mp3Uploader) PutMP3(ctx context.Context, episodeID string, mp3 []byte) error {
	if u == nil || u.http == nil || u.tokens == nil {
		return fmt.Errorf("uploader 未初期化")
	}
	episodeID = strings.TrimSpace(episodeID)
	if episodeID == "" {
		return fmt.Errorf("episodeID が空")
	}
	if len(mp3) == 0 {
		return fmt.Errorf("mp3 が空")
	}
	token, err := u.tokens.Token(ctx)
	if err != nil {
		return fmt.Errorf("token: %w", err)
	}
	name := episodeID + ".mp3"
	fileID, err := u.findFileID(ctx, token, name)
	if err != nil {
		return err
	}
	if fileID == "" {
		fileID, err = u.createMetadata(ctx, token, name)
		if err != nil {
			return err
		}
	}
	return u.uploadMedia(ctx, token, fileID, mp3)
}

func (u *mp3Uploader) findFileID(ctx context.Context, token, name string) (string, error) {
	q := url.Values{}
	q.Set("q", "name = '"+escapeDriveQueryValue(name)+"' and trashed = false")
	q.Set("fields", "files(id)")
	res, err := u.do(ctx, http.MethodGet, gdrive.FilesURL+"?"+q.Encode(), token, "", nil)
	if err != nil {
		return "", fmt.Errorf("list: %w", err)
	}
	var parsed struct {
		Files []struct {
			ID string `json:"id"`
		} `json:"files"`
	}
	if err := readJSONBody(res, &parsed, http.StatusOK); err != nil {
		return "", fmt.Errorf("list: %w", err)
	}
	if len(parsed.Files) == 0 {
		return "", nil
	}
	return parsed.Files[0].ID, nil
}

func (u *mp3Uploader) createMetadata(ctx context.Context, token, name string) (string, error) {
	meta, err := json.Marshal(map[string]any{
		"name":     name,
		"mimeType": "audio/mpeg",
		"parents":  []string{u.folderID},
	})
	if err != nil {
		return "", err
	}
	res, err := u.do(ctx, http.MethodPost, gdrive.FilesURL, token, "application/json", meta)
	if err != nil {
		return "", fmt.Errorf("create: %w", err)
	}
	var parsed struct {
		ID string `json:"id"`
	}
	if err := readJSONBody(res, &parsed, http.StatusOK, http.StatusCreated); err != nil {
		return "", fmt.Errorf("create: %w", err)
	}
	if strings.TrimSpace(parsed.ID) == "" {
		return "", fmt.Errorf("create: file id empty")
	}
	return parsed.ID, nil
}

func (u *mp3Uploader) uploadMedia(ctx context.Context, token, fileID string, content []byte) error {
	target := gdrive.UploadURL + "/" + url.PathEscape(fileID) + "?uploadType=media"
	res, err := u.do(ctx, http.MethodPatch, target, token, "audio/mpeg", content)
	if err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("upload: status %d", res.StatusCode)
	}
	return nil
}

func (u *mp3Uploader) do(ctx context.Context, method, target, token, contentType string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return u.http.Do(req)
}

func readJSONBody(res *http.Response, dest any, okStatuses ...int) error {
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	ok := false
	for _, status := range okStatuses {
		if res.StatusCode == status {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("status %d", res.StatusCode)
	}
	return json.Unmarshal(raw, dest)
}

func escapeDriveQueryValue(name string) string {
	name = strings.ReplaceAll(name, `\`, `\\`)
	return strings.ReplaceAll(name, `'`, `\'`)
}
