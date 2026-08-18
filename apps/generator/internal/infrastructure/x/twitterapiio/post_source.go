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
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

const (
	lastTweetsEndpoint = "https://api.twitterapi.io/twitter/user/last_tweets"
	createdAtLayout    = "Mon Jan 02 15:04:05 -0700 2006"
	includeRepliesNo   = "false"
	apiKeyHeaderName   = "X-API-Key"
)

var _ port.PostSource = (*PostSource)(nil)

type PostSource struct {
	client *agentsecrets.Client
}

// NewPostSource は TwitterAPI.io 向け PostSource を返す。
//
// @require client != nil
// @ensure 秘密値は保持しない。
func NewPostSource(client *agentsecrets.Client) *PostSource {
	return &PostSource{client: client}
}

func (s *PostSource) ListByUser(ctx context.Context, userID string, since time.Time) ([]models.Post, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("list_by_user", fmt.Errorf("client is nil"))
	}

	out := make([]models.Post, 0)
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
			if createdAt.Before(since) {
				// why: OpenAPI は created_at 降順。下限を下回ったら以降 page も古い。
				return out, nil
			}
			if !isOriginal(raw) {
				continue
			}
			out = append(out, toPost(raw, createdAt))
		}
		if !page.HasNextPage || page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	return out, nil
}

func (s *PostSource) fetchPage(ctx context.Context, userID, cursor string) (lastTweetsResponse, error) {
	target := lastTweetsURL(userID, cursor)
	// why: TwitterAPI.io 公式 auth は X-API-Key。Bearer は Authorization になり合わない。
	res, err := s.client.Do(ctx, agentsecrets.Request{
		Method:    http.MethodGet,
		TargetURL: target,
		Inject: agentsecrets.Inject{
			Headers: map[string]string{apiKeyHeaderName: secretnames.TwitterIOAPIKeyName},
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

func lastTweetsURL(userID, cursor string) string {
	q := url.Values{}
	q.Set("userId", userID)
	q.Set("includeReplies", includeRepliesNo)
	// why: 初回は cursor 空が正。空文字を query に付けると vendor が次 page と誤読しうる。
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
	ID        string      `json:"id"`
	URL       string      `json:"url"`
	Text      string      `json:"text"`
	CreatedAt string      `json:"createdAt"`
	IsReply   bool        `json:"isReply"`
	Author    rawAuthor   `json:"author"`
	Entities  rawEntities `json:"entities"`
	// why: 中身は捨て、null 以外の存在だけで引用・Repost と判定する。
	QuotedTweet    json.RawMessage `json:"quoted_tweet"`
	RetweetedTweet json.RawMessage `json:"retweeted_tweet"`
}

type rawAuthor struct {
	ID string `json:"id"`
}

type rawEntities struct {
	URLs []rawURL `json:"urls"`
}

type rawURL struct {
	ExpandedURL string `json:"expanded_url"`
}

func isOriginal(t rawTweet) bool {
	// why: includeReplies=false は Reply だけ落とす。引用・Repost は quoted_tweet / retweeted_tweet で除外する。
	return !t.IsReply && !hasEmbeddedTweet(t.QuotedTweet) && !hasEmbeddedTweet(t.RetweetedTweet)
}

func hasEmbeddedTweet(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s != "" && s != "null"
}

func toPost(t rawTweet, createdAt time.Time) models.Post {
	urls := make([]string, 0, len(t.Entities.URLs))
	for _, u := range t.Entities.URLs {
		if u.ExpandedURL == "" {
			continue
		}
		urls = append(urls, u.ExpandedURL)
	}
	return models.Post{
		ID:        t.ID,
		AuthorID:  t.Author.ID,
		Text:      t.Text,
		CreatedAt: createdAt,
		Permalink: t.URL,
		URLs:      urls,
		// why: OpenAPI の Tweet に media field が無い。
		Media: []models.Media{},
	}
}
