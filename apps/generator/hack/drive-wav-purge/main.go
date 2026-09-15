package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/google/oauth"
)

// main は Drive folder の .wav を消し、完成ペア（json+mp3）だけを確認する使い捨て入口である。
//
// @require process environment に Drive OAuth 4 key がある。
// @ensure wav 削除後に完成ペアのみ。Cursor / Gemini を読まない。
func main() {
	clientID := strings.TrimSpace(os.Getenv(config.GoogleOAuthClientIDEnv))
	clientSecret := strings.TrimSpace(os.Getenv(config.GoogleOAuthClientSecretEnv))
	refreshToken := strings.TrimSpace(os.Getenv(config.GoogleOAuthRefreshTokenEnv))
	folderID := strings.TrimSpace(os.Getenv(config.DriveFolderIDEnv))
	if clientID == "" || clientSecret == "" || refreshToken == "" || folderID == "" {
		fmt.Fprintf(os.Stderr, "drive-wav-purge: Drive OAuth / folder の env が不足（%s %s %s %s）\n",
			config.GoogleOAuthClientIDEnv,
			config.GoogleOAuthClientSecretEnv,
			config.GoogleOAuthRefreshTokenEnv,
			config.DriveFolderIDEnv,
		)
		os.Exit(1)
	}

	httpClient := &http.Client{Timeout: 2 * time.Minute}
	tokens := oauth.NewTokenSource(httpClient, clientID, clientSecret, refreshToken)
	c := &purgeClient{
		http:     httpClient,
		tokens:   tokens,
		folderID: folderID,
		filesURL: gdrive.FilesURL,
	}
	if err := c.purgeWavAndVerify(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "drive-wav-purge: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("drive-wav-purge: ok（wav 削除済み・完成ペアのみ）")
}

type tokenSource interface {
	Token(ctx context.Context) (string, error)
}

type purgeClient struct {
	http     *http.Client
	tokens   tokenSource
	folderID string
	filesURL string
}

func (c *purgeClient) purgeWavAndVerify(ctx context.Context) error {
	token, err := c.tokens.Token(ctx)
	if err != nil {
		return fmt.Errorf("token: %w", err)
	}
	before, err := c.listFolderFiles(ctx, token)
	if err != nil {
		return err
	}
	for _, id := range wavIDs(before) {
		if err := c.deleteFile(ctx, token, id); err != nil {
			return err
		}
	}
	after, err := c.listFolderFiles(ctx, token)
	if err != nil {
		return err
	}
	return assertCompletedSetsOnly(after)
}

func (c *purgeClient) listFolderFiles(ctx context.Context, token string) ([]listedFile, error) {
	var out []listedFile
	pageToken := ""
	for {
		q := url.Values{}
		q.Set("q", "'"+escapeDriveQueryValue(c.folderID)+"' in parents and trashed = false")
		q.Set("fields", "nextPageToken,files(id,name)")
		q.Set("pageSize", "1000")
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		res, err := c.do(ctx, http.MethodGet, c.filesURL+"?"+q.Encode(), token)
		if err != nil {
			return nil, fmt.Errorf("list: %w", err)
		}
		var parsed struct {
			NextPageToken string `json:"nextPageToken"`
			Files         []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"files"`
		}
		if err := readJSONBody(res, &parsed, http.StatusOK); err != nil {
			return nil, fmt.Errorf("list: %w", err)
		}
		for _, f := range parsed.Files {
			out = append(out, listedFile{ID: f.ID, Name: f.Name})
		}
		if parsed.NextPageToken == "" {
			return out, nil
		}
		pageToken = parsed.NextPageToken
	}
}

func (c *purgeClient) deleteFile(ctx context.Context, token, fileID string) error {
	res, err := c.do(ctx, http.MethodDelete, c.filesURL+"/"+url.PathEscape(fileID), token)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete: status %d", res.StatusCode)
	}
	return nil
}

func (c *purgeClient) do(ctx context.Context, method, target, token string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return c.http.Do(req)
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
