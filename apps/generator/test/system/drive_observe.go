//go:build system

package system

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/google/oauth"
)

// driveFile は System 観測用の Drive file 要約。
type driveFile struct {
	ID   string
	Name string
	Size int64
}

type driveObserver struct {
	client   *http.Client
	tokens   *oauth.TokenSource
	folderID string
}

func requireSystemEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		config.GetXAPIKeyEnv,
		config.CursorAPIKeyEnv,
		config.GeminiAPIKeyEnv,
		config.GoogleOAuthClientIDEnv,
		config.GoogleOAuthClientSecretEnv,
		config.GoogleOAuthRefreshTokenEnv,
		config.DriveFolderIDEnv,
	}
	var missing []string
	for _, key := range keys {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("System precondition: process env 不足: %s", strings.Join(missing, ", "))
	}
}

func newDriveObserver(t *testing.T) *driveObserver {
	t.Helper()
	client := &http.Client{}
	tokens := oauth.NewTokenSource(
		client,
		os.Getenv(config.GoogleOAuthClientIDEnv),
		os.Getenv(config.GoogleOAuthClientSecretEnv),
		os.Getenv(config.GoogleOAuthRefreshTokenEnv),
	)
	return &driveObserver{
		client:   client,
		tokens:   tokens,
		folderID: os.Getenv(config.DriveFolderIDEnv),
	}
}

func (o *driveObserver) listFolder(ctx context.Context) (map[string]driveFile, error) {
	token, err := o.tokens.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("token: %w", err)
	}
	out := make(map[string]driveFile)
	pageToken := ""
	for {
		q := url.Values{}
		q.Set("q", fmt.Sprintf("'%s' in parents and trashed = false", escapeDriveQuery(o.folderID)))
		q.Set("fields", "nextPageToken,files(id,name,size)")
		q.Set("pageSize", "1000")
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		res, err := o.do(ctx, http.MethodGet, gdrive.FilesURL+"?"+q.Encode(), token, "", nil)
		if err != nil {
			return nil, err
		}
		var parsed struct {
			NextPageToken string `json:"nextPageToken"`
			Files         []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Size string `json:"size"`
			} `json:"files"`
		}
		if err := decodeJSON(res, &parsed); err != nil {
			return nil, err
		}
		for _, f := range parsed.Files {
			var size int64
			if f.Size != "" {
				_, _ = fmt.Sscan(f.Size, &size)
			}
			out[f.Name] = driveFile{ID: f.ID, Name: f.Name, Size: size}
		}
		if parsed.NextPageToken == "" {
			break
		}
		pageToken = parsed.NextPageToken
	}
	return out, nil
}

func (o *driveObserver) download(ctx context.Context, fileID string) ([]byte, error) {
	token, err := o.tokens.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("token: %w", err)
	}
	target := gdrive.FilesURL + "/" + url.PathEscape(fileID) + "?alt=media"
	res, err := o.do(ctx, http.MethodGet, target, token, "", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download status %d", res.StatusCode)
	}
	return body, nil
}

func (o *driveObserver) delete(ctx context.Context, fileID string) error {
	token, err := o.tokens.Token(ctx)
	if err != nil {
		return fmt.Errorf("token: %w", err)
	}
	target := gdrive.FilesURL + "/" + url.PathEscape(fileID)
	res, err := o.do(ctx, http.MethodDelete, target, token, "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete status %d", res.StatusCode)
	}
	return nil
}

func (o *driveObserver) do(ctx context.Context, method, target, token, contentType string, body []byte) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return o.client.Do(req)
}

func decodeJSON(res *http.Response, dest any) error {
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("drive http status %d", res.StatusCode)
	}
	return json.Unmarshal(raw, dest)
}

func escapeDriveQuery(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	return strings.ReplaceAll(v, `'`, `\'`)
}

func addedNames(before, after map[string]driveFile) []driveFile {
	var out []driveFile
	for name, f := range after {
		if _, ok := before[name]; !ok {
			out = append(out, f)
		}
	}
	return out
}

func completeStem(added []driveFile) (stem string, jsonFile, wavFile driveFile, ok bool) {
	jsonByStem := map[string]driveFile{}
	wavByStem := map[string]driveFile{}
	for _, f := range added {
		switch {
		case strings.HasSuffix(f.Name, ".json"):
			jsonByStem[strings.TrimSuffix(f.Name, ".json")] = f
		case strings.HasSuffix(f.Name, ".wav"):
			wavByStem[strings.TrimSuffix(f.Name, ".wav")] = f
		}
	}
	for s, jf := range jsonByStem {
		if wf, hit := wavByStem[s]; hit {
			return s, jf, wf, true
		}
	}
	return "", driveFile{}, driveFile{}, false
}
