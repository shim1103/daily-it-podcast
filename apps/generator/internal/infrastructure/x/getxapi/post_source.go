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
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

const (
	userTweetsEndpoint = "https://api.getxapi.com/twitter/user/tweets"
	createdAtLayout    = "Mon Jan 02 15:04:05 -0700 2006"
)

var _ port.PostSource = (*PostSource)(nil)

type PostSource struct {
	client *agentsecrets.Client
}

// NewPostSource は GetXAPI 向け PostSource を返す。
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
				// why: docs は順序未記載。新しい順と仮定し、下限を下回ったら以降 page も古い。
				return out, nil
			}
			if !isOriginal(raw) {
				continue
			}
			out = append(out, toPost(raw, createdAt))
		}
		if !page.HasMore || page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	return out, nil
}

func (s *PostSource) fetchPage(ctx context.Context, userID, cursor string) (userTweetsResponse, error) {
	target := userTweetsURL(userID, cursor)
	// why: 公式 auth は Authorization Bearer。X-API-Key header 注入は合わない。
	res, err := s.client.Do(ctx, agentsecrets.Request{
		Method:    http.MethodGet,
		TargetURL: target,
		Inject: agentsecrets.Inject{
			Bearer: secretnames.GetXAPIKeyName,
		},
	})
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

func userTweetsURL(userID, cursor string) string {
	q := url.Values{}
	q.Set("userId", userID)
	// why: 初回は cursor 空が正。空文字を query に付けると vendor が次 page と誤読しうる。
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
	ID        string      `json:"id"`
	URL       string      `json:"url"`
	Text      string      `json:"text"`
	CreatedAt string      `json:"createdAt"`
	IsReply   bool        `json:"isReply"`
	Author    rawAuthor   `json:"author"`
	Entities  rawEntities `json:"entities"`
	Media     []rawMedia  `json:"media"`
	// why: 中身は捨て、null 以外の存在だけで引用と判定する。
	QuotedTweet json.RawMessage `json:"quoted_tweet"`
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

// why: expanded_url は投稿 permalink。取得可能なメディア URL は url。
type rawMedia struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func isOriginal(t rawTweet) bool {
	// why: User Tweets docs sample に retweeted_tweet が無い。Posts tab の応答に従い field を持たない。
	return !t.IsReply && !hasEmbeddedTweet(t.QuotedTweet)
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
	media := make([]models.Media, 0, len(t.Media))
	for _, m := range t.Media {
		if m.URL == "" {
			continue
		}
		media = append(media, models.Media{Type: m.Type, URL: m.URL})
	}
	return models.Post{
		ID:        t.ID,
		AuthorID:  t.Author.ID,
		Text:      t.Text,
		CreatedAt: createdAt,
		Permalink: t.URL,
		URLs:      urls,
		Media:     media,
	}
}
