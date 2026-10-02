package hackernews

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

// why: podcast は 1 日 1 回だけ生成する。結果件数と同時 fetch 数を上限で抑え、
// id 列は最大 500 件なので 1 run あたりの fetch 回数も有界に収まる。
const (
	// MaxStoriesScanned は結果 SourceItem に含める story 数の上限。
	MaxStoriesScanned = 20
	// MaxCommentsPerStory は 1 story あたり取得する top-level comment 数の上限。
	MaxCommentsPerStory = 8
	// CommentDepth は取得する comment 階層の深さ（top-level のみ）。
	CommentDepth = 1
	// MaxConcurrentFetches は item/<id>.json への同時 fetch 数の上限。
	// why: Hacker News API 側の rate limit は非公開のため、安全側に抑えた値とする。
	MaxConcurrentFetches = 5
)

// ListItemSource は Hacker News の topstories を ItemSource として返す Adapter。
type ListItemSource struct {
	client   *http.Client
	maxItems int
}

// NewListItemSource は Hacker News 向け ItemSource を返す。
//
// @require httpClient != nil
// @ensure 戻りは非 nil の *ListItemSource。vendor 固有型を露出しない。
// @ensure maxItems <= 0 の場合、実効上限は MaxStoriesScanned にフォールバックする。
func NewListItemSource(httpClient *http.Client, maxItems int) *ListItemSource {
	return &ListItemSource{client: httpClient, maxItems: maxItems}
}

// List は since 以降に発生した Hacker News story を SourceItem slice で返す。
//
// @ensure 各要素の SourceID は SourceID。最大 min(maxItems, MaxStoriesScanned) 件（maxItems <= 0 は MaxStoriesScanned）。
// @ensure 結果の順序は保証しない。
// @ensure id 列の個別 fetch は最大 MaxConcurrentFetches 件まで同時実行してよい。
func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("list", fmt.Errorf("client is nil"))
	}

	ids, err := s.fetchTopStoryIDs(ctx)
	if err != nil {
		return nil, err
	}

	// why: docs/decisions/2026-09-23T17-21-29
	limit := s.effectiveMaxStories()
	slots, filled, err := s.fetchStoriesInWindow(ctx, ids, since)
	if err != nil {
		return nil, err
	}

	out := make([]models.SourceItem, 0, limit)
	for i := range slots {
		if !filled[i] {
			continue
		}
		out = append(out, slots[i])
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

const apiBaseURL = "https://hacker-news.firebaseio.com/v0"

const permalinkBaseURL = "https://news.ycombinator.com/item?id="

type hnItem struct {
	ID      int64   `json:"id"`
	Type    string  `json:"type"`
	By      string  `json:"by"`
	Time    int64   `json:"time"`
	Title   string  `json:"title"`
	Text    string  `json:"text"`
	URL     string  `json:"url"`
	Kids    []int64 `json:"kids"`
	Deleted bool    `json:"deleted"`
	Dead    bool    `json:"dead"`
}

// why: maxItems <= 0 は呼び出し側の指定漏れとみなし、MaxStoriesScanned を安全側の値として使う。
func (s *ListItemSource) effectiveMaxStories() int {
	if s.maxItems <= 0 || s.maxItems > MaxStoriesScanned {
		return MaxStoriesScanned
	}
	return s.maxItems
}

func (s *ListItemSource) fetchStoriesInWindow(ctx context.Context, ids []int64, since time.Time) ([]models.SourceItem, []bool, error) {
	slots := make([]models.SourceItem, len(ids))
	filled := make([]bool, len(ids))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(MaxConcurrentFetches)
	for i, id := range ids {
		g.Go(func() error {
			story, ok := s.fetchStoryInWindow(gctx, id, since)
			if !ok {
				return nil
			}
			comments := s.fetchTopLevelComments(gctx, story.Kids)
			slots[i] = toSourceItem(story, comments)
			filled[i] = true
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, nil, err
	}
	return slots, filled, nil
}

func (s *ListItemSource) fetchTopStoryIDs(ctx context.Context) ([]int64, error) {
	body, err := s.getWithRetry(ctx, apiBaseURL+"/topstories.json", "fetch_top_stories")
	if err != nil {
		return nil, err
	}
	var ids []int64
	if err := json.Unmarshal(body, &ids); err != nil {
		return nil, infraErr("decode_top_stories", err)
	}
	return ids, nil
}

func (s *ListItemSource) fetchStoryInWindow(ctx context.Context, id int64, since time.Time) (hnItem, bool) {
	item, err := s.fetchItem(ctx, id)
	if err != nil {
		return hnItem{}, false
	}
	if item.Type != "story" || item.Deleted || item.Dead {
		return hnItem{}, false
	}
	if item.Time < since.Unix() {
		return hnItem{}, false
	}
	return item, true
}

func (s *ListItemSource) fetchTopLevelComments(ctx context.Context, kids []int64) []string {
	limit := MaxCommentsPerStory
	if len(kids) < limit {
		limit = len(kids)
	}
	slots := s.fetchCommentBodies(ctx, kids[:limit])

	bodies := make([]string, 0, limit)
	for _, body := range slots {
		if body == "" {
			continue
		}
		bodies = append(bodies, body)
	}
	return bodies
}

func (s *ListItemSource) fetchCommentBodies(ctx context.Context, kids []int64) []string {
	slots := make([]string, len(kids))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(MaxConcurrentFetches)
	for i, kid := range kids {
		g.Go(func() error {
			item, err := s.fetchItem(gctx, kid)
			if err != nil {
				return nil
			}
			normalized := httpget.NormalizeHTML(item.Text)
			if normalized == "" {
				return nil
			}
			slots[i] = normalized
			return nil
		})
	}
	_ = g.Wait()
	return slots
}

func (s *ListItemSource) fetchItem(ctx context.Context, id int64) (hnItem, error) {
	url := fmt.Sprintf("%s/item/%d.json", apiBaseURL, id)
	body, err := s.getWithRetry(ctx, url, "fetch_item")
	if err != nil {
		return hnItem{}, err
	}
	var item hnItem
	if err := json.Unmarshal(body, &item); err != nil {
		return hnItem{}, infraErr("decode_item", err)
	}
	return item, nil
}

func (s *ListItemSource) getWithRetry(ctx context.Context, url, op string) ([]byte, error) {
	body, err := httpget.GetWithRetry(ctx, s.client, url)
	if err != nil {
		return nil, infraErr(op, err)
	}
	return body, nil
}

// why: docs/decisions/2026-09-13T17-14-10
func toSourceItem(story hnItem, commentBodies []string) models.SourceItem {
	detail := models.SourceBody{Text: httpget.NormalizeHTML(story.Text)}
	if story.URL != "" {
		detail.Links = []string{story.URL}
	}
	discourse := models.SourceBody{
		Text:  strings.Join(commentBodies, "\n"),
		Links: []string{fmt.Sprintf("%s%d", permalinkBaseURL, story.ID)},
	}
	metaLines := []string{fmt.Sprintf("item_id: %d", story.ID)}
	if story.By != "" {
		metaLines = append(metaLines, "actor_id: "+story.By, "actor_name: "+story.By)
	}
	return models.SourceItem{
		SourceID:   SourceID,
		OccurredAt: time.Unix(story.Time, 0).UTC(),
		Summary:    story.Title,
		Detail:     detail,
		Discourse:  discourse,
		Meta:       strings.Join(metaLines, "\n"),
	}
}
