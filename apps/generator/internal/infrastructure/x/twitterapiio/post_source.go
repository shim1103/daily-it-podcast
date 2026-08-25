package twitterapiio

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
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	xinfra "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x"
)

var _ port.ItemSource = (*PostSource)(nil)

type PostSource struct {
	client       secrettransport.Client
	apiKeySecret secrettransport.SecretRef
}

// NewPostSource は TwitterAPI.io 向け ItemSource を返す。
//
// @require client != nil
// @ensure 秘密値は保持しない。secret 名の知識は持たず、apiKeySecret の参照だけを保持する。
func NewPostSource(client secrettransport.Client, apiKeySecret secrettransport.SecretRef) *PostSource {
	return &PostSource{client: client, apiKeySecret: apiKeySecret}
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
		if !page.HasNextPage || page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	return out, nil
}

func (s *PostSource) fetchPage(ctx context.Context, userID, cursor string) (lastTweetsResponse, error) {
	res, err := s.client.Do(ctx, secrettransport.Request{
		Method:    http.MethodGet,
		TargetURL: lastTweetsURL(userID, cursor),
		Inject: secrettransport.Inject{
			Headers: []secrettransport.FieldInjection{{Field: apiKeyHeaderName, Secret: s.apiKeySecret}},
		},
	})
	if err != nil {
		return lastTweetsResponse{}, infraErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return lastTweetsResponse{}, infraErr("read_body", err)
	}
	if res.StatusCode != http.StatusOK {
		return lastTweetsResponse{}, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	}
	var page lastTweetsResponse
	if err := json.Unmarshal(body, &page); err != nil {
		return lastTweetsResponse{}, infraErr("decode", err)
	}
	if page.Status == "error" {
		return lastTweetsResponse{}, infraErr("vendor_status", fmt.Errorf("%s", page.Message))
	}
	return page, nil
}

const (
	lastTweetsEndpoint = "https://api.twitterapi.io/twitter/user/last_tweets"
	createdAtLayout    = "Mon Jan 02 15:04:05 -0700 2006"
	includeRepliesNo   = "false"
	apiKeyHeaderName   = "X-API-Key"
)

func lastTweetsURL(userID, cursor string) string {
	q := url.Values{"userId": []string{userID}, "includeReplies": []string{includeRepliesNo}}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	return lastTweetsEndpoint + "?" + q.Encode()
}

type lastTweetsResponse struct {
	Tweets      []rawTweet `json:"tweets"`
	HasNextPage bool       `json:"has_next_page"`
	NextCursor  string     `json:"next_cursor"`
	Status      string     `json:"status"`
	Message     string     `json:"message"`
}

type rawTweet struct {
	ID             string          `json:"id"`
	URL            string          `json:"url"`
	Text           string          `json:"text"`
	CreatedAt      string          `json:"createdAt"`
	IsReply        bool            `json:"isReply"`
	Author         rawAuthor       `json:"author"`
	Entities       rawEntities     `json:"entities"`
	QuotedTweet    json.RawMessage `json:"quoted_tweet"`
	RetweetedTweet json.RawMessage `json:"retweeted_tweet"`
}

type rawAuthor struct {
	ID       string `json:"id"`
	UserName string `json:"userName"`
}

type rawEntities struct {
	URLs []rawURL `json:"urls"`
}

type rawURL struct {
	ExpandedURL string `json:"expanded_url"`
}

func isOriginal(t rawTweet) bool {
	return !t.IsReply && !hasEmbeddedTweet(t.QuotedTweet) && !hasEmbeddedTweet(t.RetweetedTweet)
}

func hasEmbeddedTweet(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s != "" && s != "null"
}

func toSourceItem(t rawTweet, occurredAt time.Time) models.SourceItem {
	lines := []string{"item_id: " + t.ID, "actor_id: " + t.Author.ID}
	if t.Author.UserName != "" {
		lines = append(lines, "actor_name: "+t.Author.UserName)
	}
	lines = append(lines, "text: "+t.Text, "permalink: "+t.URL)
	urls := make([]string, 0, len(t.Entities.URLs))
	for _, item := range t.Entities.URLs {
		if item.ExpandedURL != "" {
			urls = append(urls, item.ExpandedURL)
		}
	}
	if len(urls) > 0 {
		lines = append(lines, "links: "+strings.Join(urls, " "))
	}
	return models.SourceItem{SourceID: xinfra.SourceID, OccurredAt: occurredAt, Context: strings.Join(lines, "\n")}
}
