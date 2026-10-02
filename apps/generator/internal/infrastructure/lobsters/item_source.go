package lobsters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/httpget"
	"golang.org/x/sync/errgroup"
)

var _ port.ItemSource = (*ListItemSource)(nil)

// why: podcast は 1 日 1 回だけ生成する。上限値で 1 run あたりの fetch 回数を有界にする。
const (
	// MaxStoriesScanned は結果 SourceItem に含める story 数の上限。
	MaxStoriesScanned = 20
	// MaxCommentsPerStory は 1 story あたり取得する comment 数の上限。
	MaxCommentsPerStory = 8
	// CommentDepth は取得する comment 階層の深さ（top-level のみ）。
	CommentDepth = 1
	// MaxConcurrentFetches は /s/<short_id>.json への同時 fetch 数の上限。
	// why: Lobsters API 側の rate limit は非公開のため、安全側に抑えた値とする。
	MaxConcurrentFetches = 5
)

// ListItemSource は Lobsters の hottest を ItemSource として返す Adapter。
type ListItemSource struct {
	client   *http.Client
	maxItems int
}

// NewListItemSource は Lobsters 向け ItemSource を返す。
//
// @require httpClient != nil
// @ensure 戻りは非 nil の *ListItemSource。vendor 固有型を露出しない。
// @ensure maxItems <= 0 の場合、実効上限は MaxStoriesScanned にフォールバックする。
func NewListItemSource(httpClient *http.Client, maxItems int) *ListItemSource {
	return &ListItemSource{client: httpClient, maxItems: maxItems}
}

// List は since 以降に発生した Lobsters story を SourceItem slice で返す。
//
// @ensure 各要素の SourceID は SourceID。最大 min(maxItems, MaxStoriesScanned) 件（maxItems <= 0 は MaxStoriesScanned）。
// @ensure 結果の順序は保証しない。
// @ensure targets の個別 fetch は最大 MaxConcurrentFetches 件まで同時実行してよい。
func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("list", fmt.Errorf("client is nil"))
	}

	summaries, err := s.fetchHottest(ctx)
	if err != nil {
		return nil, err
	}

	targets := filterSummariesInWindow(summaries, since, s.effectiveMaxStories())

	items, ok, err := s.fetchStoryDetails(ctx, targets, since)
	if err != nil {
		return nil, err
	}

	out := make([]models.SourceItem, 0, len(targets))
	for i := range ok {
		if ok[i] {
			out = append(out, items[i])
		}
	}
	return out, nil
}

const apiBaseURL = "https://lobste.rs"

type hottestSummary struct {
	ShortID   string `json:"short_id"`
	CreatedAt string `json:"created_at"`
}

type storyDetail struct {
	ShortID          string         `json:"short_id"`
	SubmitterUser    string         `json:"submitter_user"`
	Title            string         `json:"title"`
	DescriptionPlain string         `json:"description_plain"`
	URL              string         `json:"url"`
	ShortIDURL       string         `json:"short_id_url"`
	CommentsURL      string         `json:"comments_url"`
	CreatedAt        string         `json:"created_at"`
	Comments         []storyComment `json:"comments"`
}

type storyComment struct {
	CommentPlain string `json:"comment_plain"`
	IsDeleted    bool   `json:"is_deleted"`
	IsModerated  bool   `json:"is_moderated"`
}

// why: maxItems <= 0 は呼び出し側の指定漏れとみなし、MaxStoriesScanned を安全側の値として使う。
func (s *ListItemSource) effectiveMaxStories() int {
	if s.maxItems <= 0 || s.maxItems > MaxStoriesScanned {
		return MaxStoriesScanned
	}
	return s.maxItems
}

func (s *ListItemSource) fetchStoryDetails(ctx context.Context, targets []hottestSummary, since time.Time) ([]models.SourceItem, []bool, error) {
	items := make([]models.SourceItem, len(targets))
	ok := make([]bool, len(targets))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(MaxConcurrentFetches)
	for i := range targets {
		g.Go(func() error {
			summary := targets[i]
			detail, occurredAt, fetched := s.fetchStoryDetail(gctx, summary.ShortID, since)
			if !fetched {
				return nil
			}
			items[i] = toSourceItem(detail, occurredAt)
			ok[i] = true
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, nil, err
	}
	return items, ok, nil
}

func (s *ListItemSource) fetchHottest(ctx context.Context) ([]hottestSummary, error) {
	body, err := s.getWithRetry(ctx, apiBaseURL+"/hottest.json", "fetch_hottest")
	if err != nil {
		return nil, err
	}
	var summaries []hottestSummary
	if err := json.Unmarshal(body, &summaries); err != nil {
		return nil, infraErr("decode_hottest", err)
	}
	return summaries, nil
}

func filterSummariesInWindow(summaries []hottestSummary, since time.Time, limit int) []hottestSummary {
	out := make([]hottestSummary, 0, limit)
	for _, summary := range summaries {
		createdAt, err := parseCreatedAt(summary.CreatedAt)
		if err != nil || createdAt.Before(since) {
			continue
		}
		out = append(out, summary)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func (s *ListItemSource) fetchStoryDetail(ctx context.Context, shortID string, since time.Time) (storyDetail, time.Time, bool) {
	url := fmt.Sprintf("%s/s/%s.json", apiBaseURL, shortID)
	body, err := s.getWithRetry(ctx, url, "fetch_story")
	if err != nil {
		return storyDetail{}, time.Time{}, false
	}
	var detail storyDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		return storyDetail{}, time.Time{}, false
	}
	createdAt, err := parseCreatedAt(detail.CreatedAt)
	if err != nil || createdAt.Before(since) {
		return storyDetail{}, time.Time{}, false
	}
	return detail, createdAt, true
}

func (s *ListItemSource) getWithRetry(ctx context.Context, url, op string) ([]byte, error) {
	body, err := httpget.GetWithRetry(ctx, s.client, url)
	if err != nil {
		return nil, infraErr(op, err)
	}
	return body, nil
}

func parseCreatedAt(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func collectCommentBodies(comments []storyComment) []string {
	bodies := make([]string, 0, MaxCommentsPerStory)
	for _, c := range comments {
		if c.IsDeleted || c.IsModerated || c.CommentPlain == "" {
			continue
		}
		bodies = append(bodies, c.CommentPlain)
		if len(bodies) >= MaxCommentsPerStory {
			break
		}
	}
	return bodies
}

// why: docs/decisions/2026-09-13T17-14-10
func toSourceItem(detail storyDetail, occurredAt time.Time) models.SourceItem {
	dBody := models.SourceBody{Text: detail.DescriptionPlain}
	if detail.URL != "" {
		dBody.Links = []string{detail.URL}
	}
	discourse := models.SourceBody{Text: strings.Join(collectCommentBodies(detail.Comments), "\n")}
	if detail.ShortIDURL != "" {
		discourse.Links = []string{detail.ShortIDURL}
	}
	metaLines := []string{"item_id: " + detail.ShortID}
	if detail.SubmitterUser != "" {
		metaLines = append(metaLines, "actor_id: "+detail.SubmitterUser, "actor_name: "+detail.SubmitterUser)
	}
	return models.SourceItem{
		SourceID:   SourceID,
		OccurredAt: occurredAt.UTC(),
		Summary:    detail.Title,
		Detail:     dBody,
		Discourse:  discourse,
		Meta:       strings.Join(metaLines, "\n"),
	}
}
