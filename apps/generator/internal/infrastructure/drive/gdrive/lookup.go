package gdrive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

var _ port.CompletedEpisodeLookup = (*CompletedEpisodeLookup)(nil)

// CompletedEpisodeLookup は Drive 上の完成ペア（同一 stem の json+wav）を表示 date で照会する。
type CompletedEpisodeLookup struct {
	client   *http.Client
	tokens   TokenSource
	folderID string
}

// NewCompletedEpisodeLookup は Google Drive を照会先とする CompletedEpisodeLookup を返す。
//
// @require httpClient != nil。tokens != nil。
// @ensure folderID は list query の parents にだけ使い、保存元の知識は持たない。
func NewCompletedEpisodeLookup(httpClient *http.Client, tokens TokenSource, folderID string) *CompletedEpisodeLookup {
	return &CompletedEpisodeLookup{client: httpClient, tokens: tokens, folderID: folderID}
}

// HasPair は所定 folder に date 一致の完成ペアがあるとき true を返す。
func (l *CompletedEpisodeLookup) HasPair(ctx context.Context, date string) (bool, error) {
	if l == nil || l.client == nil {
		return false, infraErr("has_pair", fmt.Errorf("client is nil"))
	}

	token, err := l.tokens.Token(ctx)
	if err != nil {
		return false, infraErr("token", err)
	}

	files, err := l.listFolderFiles(ctx, token)
	if err != nil {
		return false, err
	}

	wavStems := make(map[string]struct{})
	var jsonFiles []driveListedFile
	for _, f := range files {
		switch {
		case strings.HasSuffix(f.Name, jsonExt):
			jsonFiles = append(jsonFiles, f)
		case strings.HasSuffix(f.Name, wavExt):
			wavStems[strings.TrimSuffix(f.Name, wavExt)] = struct{}{}
		}
	}

	for _, jf := range jsonFiles {
		stem := strings.TrimSuffix(jf.Name, jsonExt)
		if _, ok := wavStems[stem]; !ok {
			continue
		}
		raw, err := l.downloadMedia(ctx, token, jf.ID)
		if err != nil {
			return false, err
		}
		episodeDate, ok := manuscriptDate(raw)
		if !ok {
			continue
		}
		if episodeDate == date {
			return true, nil
		}
	}
	return false, nil
}

type driveListedFile struct {
	ID   string
	Name string
}

func (l *CompletedEpisodeLookup) listFolderFiles(ctx context.Context, token string) ([]driveListedFile, error) {
	var out []driveListedFile
	pageToken := ""
	for {
		q := url.Values{}
		q.Set("q", "'"+escapeDriveQueryValue(l.folderID)+"' in parents and trashed = false")
		q.Set("fields", "nextPageToken,files(id,name)")
		q.Set("pageSize", "1000")
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		res, err := doDriveHTTP(ctx, l.client, http.MethodGet, FilesURL+"?"+q.Encode(), token, "", nil)
		if err != nil {
			return nil, infraErr("list", err)
		}
		var parsed struct {
			NextPageToken string `json:"nextPageToken"`
			Files         []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"files"`
		}
		if err := readJSONBody(res, "list", &parsed, http.StatusOK); err != nil {
			return nil, err
		}
		for _, f := range parsed.Files {
			out = append(out, driveListedFile{ID: f.ID, Name: f.Name})
		}
		if parsed.NextPageToken == "" {
			return out, nil
		}
		pageToken = parsed.NextPageToken
	}
}

func (l *CompletedEpisodeLookup) downloadMedia(ctx context.Context, token, fileID string) ([]byte, error) {
	target := FilesURL + "/" + url.PathEscape(fileID) + "?alt=media"
	res, err := doDriveHTTP(ctx, l.client, http.MethodGet, target, token, "", nil)
	if err != nil {
		return nil, infraErr("download", err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, infraErr("download_read", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, infraErr("download_http", fmt.Errorf("status %d", res.StatusCode))
	}
	return raw, nil
}

// manuscriptDate は原稿 JSON の date field だけを読む。schema 全体は検証しない。
func manuscriptDate(raw []byte) (string, bool) {
	var doc struct {
		Date string `json:"date"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "", false
	}
	date := strings.TrimSpace(doc.Date)
	if date == "" {
		return "", false
	}
	return date, true
}
