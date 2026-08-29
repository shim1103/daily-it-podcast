package getxapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	xinfra "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x"
)

var _ port.ItemSource = (*PostSource)(nil)

type PostSource struct {
	client *http.Client
	apiKey string
}

// NewPostSource は GetXAPI 向け ItemSource を返す。
//
// @require httpClient != nil
// @ensure apiKey は Bearer 認証にだけ使い、保存元の知識は持たない。
func NewPostSource(httpClient *http.Client, apiKey string) *PostSource {
	return &PostSource{client: httpClient, apiKey: apiKey}
}

func (s *PostSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("list", fmt.Errorf("client is nil"))
	}
	out := make([]models.SourceItem, 0)
	for _, userID := range xinfra.WatchUserIDs {
		items, err := s.listByUser(ctx, userID, since)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

func (s *PostSource) listByUser(ctx context.Context, userID string, since time.Time) ([]models.SourceItem, error) {
	out := make([]models.SourceItem, 0)
	cursor := ""
	for {
		page, err := s.fetchPage(ctx, userID, cursor)
		if err != nil {
			return nil, err
		}
		for _, raw := range page.Tweets {
			createdAt, err := time.Parse(createdAtLayout, raw.CreatedAt)
			if err != nil {
				return nil, infraErr("parse_created_at", err)
			}
			createdAt = createdAt.UTC()
			if createdAt.Before(since) {
				return out, nil
			}
			if !isOriginal(raw) {
				continue
			}
			out = append(out, toSourceItem(raw, createdAt))
		}
		if !page.HasMore || page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	return out, nil
}

func (s *PostSource) fetchPage(ctx context.Context, userID, cursor string) (userTweetsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userTweetsURL(userID, cursor), nil)
	if err != nil {
		return userTweetsResponse{}, infraErr("build_request", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	res, err := s.client.Do(req)
	if err != nil {
		return userTweetsResponse{}, infraErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return userTweetsResponse{}, infraErr("read_body", err)
	}
	if res.StatusCode != http.StatusOK {
		return userTweetsResponse{}, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	}
	var page userTweetsResponse
	if err := json.Unmarshal(body, &page); err != nil {
		return userTweetsResponse{}, infraErr("decode", err)
	}
	return page, nil
}

const (
	userTweetsEndpoint = "https://api.getxapi.com/twitter/user/tweets"
	createdAtLayout    = "Mon Jan 02 15:04:05 -0700 2006"
)

func userTweetsURL(userID, cursor string) string {
	q := url.Values{"userId": []string{userID}}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	return userTweetsEndpoint + "?" + q.Encode()
}

type userTweetsResponse struct {
	Tweets     []rawTweet `json:"tweets"`
	HasMore    bool       `json:"has_more"`
	NextCursor string     `json:"next_cursor"`
}

type rawTweet struct {
	ID          string          `json:"id"`
	URL         string          `json:"url"`
	Text        string          `json:"text"`
	CreatedAt   string          `json:"createdAt"`
	IsReply     bool            `json:"isReply"`
	Author      rawAuthor       `json:"author"`
	Entities    rawEntities     `json:"entities"`
	Media       []rawMedia      `json:"media"`
	QuotedTweet json.RawMessage `json:"quoted_tweet"`
}

type rawAuthor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type rawEntities struct {
	URLs []rawURL `json:"urls"`
}

type rawURL struct {
	ExpandedURL string `json:"expanded_url"`
}

type rawMedia struct {
	URL string `json:"url"`
}

func isOriginal(t rawTweet) bool {
	return !t.IsReply && !hasEmbeddedTweet(t.QuotedTweet)
}

func hasEmbeddedTweet(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s != "" && s != "null"
}

func toSourceItem(t rawTweet, occurredAt time.Time) models.SourceItem {
	lines := []string{
		"item_id: " + t.ID,
		"actor_id: " + t.Author.ID,
	}
	actorName := t.Author.Name
	if actorName == "" {
		actorName = t.Author.DisplayName
	}
	if actorName != "" {
		lines = append(lines, "actor_name: "+actorName)
	}
	lines = append(lines, "text: "+t.Text, "permalink: "+t.URL)
	urls := expandedURLs(t.Entities.URLs)
	if len(urls) > 0 {
		lines = append(lines, "links: "+strings.Join(urls, " "))
	}
	media := make([]string, 0, len(t.Media))
	for _, item := range t.Media {
		if item.URL != "" {
			media = append(media, item.URL)
		}
	}
	if len(media) > 0 {
		lines = append(lines, "media: "+strings.Join(media, " "))
	}
	return models.SourceItem{SourceID: xinfra.SourceID, OccurredAt: occurredAt, Context: strings.Join(lines, "\n")}
}

func expandedURLs(raw []rawURL) []string {
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if item.ExpandedURL != "" {
			out = append(out, item.ExpandedURL)
		}
	}
	return out
}
