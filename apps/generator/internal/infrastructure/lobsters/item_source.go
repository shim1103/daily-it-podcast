package lobsters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

var _ port.ItemSource = (*ListItemSource)(nil)

// apiBaseURL は Lobsters の base。
const apiBaseURL = "https://lobste.rs"

// 取得上限（この file が契約として値を固定する）。
//
// why: podcast は 1 日 1 回だけ生成する。上限値で 1 run あたりの fetch 回数を有界にする。
const (
	// MaxStoriesScanned は結果 SourceItem に含める story 数の上限。
	// created_at >= since を満たした story を、hottest.json の先頭からこの件数まで集める。
	MaxStoriesScanned = 25
	// MaxCommentsPerStory は 1 story あたり取得する comment 数の上限。
	MaxCommentsPerStory = 8
	// CommentDepth は取得する comment 階層の深さ（top-level のみ）。
	CommentDepth = 1
)

// ListItemSource は Lobsters の hottest を ItemSource として返す Adapter。
type ListItemSource struct {
	client *http.Client
}

// NewListItemSource は Lobsters 向け ItemSource を返す。
//
// @require httpClient != nil
// @ensure 戻りは非 nil の *ListItemSource。vendor 固有型を露出しない。
func NewListItemSource(httpClient *http.Client) *ListItemSource {
	return &ListItemSource{client: httpClient}
}

// hottestSummary は hottest.json の 1 要素のうち Adapter が使う field だけを表す。
type hottestSummary struct {
	ShortID   string `json:"short_id"`
	CreatedAt string `json:"created_at"`
}

// storyDetail は /s/<short_id>.json のうち Adapter が使う field だけを表す。
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

// storyComment は story 詳細内 comments[] の 1 要素。
type storyComment struct {
	CommentPlain string `json:"comment_plain"`
	IsDeleted    bool   `json:"is_deleted"`
	IsModerated  bool   `json:"is_moderated"`
}

// List は since 以降に発生した Lobsters story を SourceItem slice で返す。
//
// @require since は OccurredAt の inclusive 下限。
// @ensure 各要素の SourceID は非空（= SourceID）。OccurredAt は UTC かつ since 以上。
// @ensure 結果は created_at >= since を満たす story のみ。最大 MaxStoriesScanned 件。
// @ensure 該当なしは空 slice（nil ではない）。
// @invariant vendor 固有型・監視対象一覧を露出しない。Context を key として解釈しない。
func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("list", fmt.Errorf("client is nil"))
	}

	summaries, err := s.fetchHottest(ctx)
	if err != nil {
		return nil, err
	}

	targets := filterSummariesInWindow(summaries, since, MaxStoriesScanned)

	out := make([]models.SourceItem, 0, len(targets))
	for _, summary := range targets {
		detail, occurredAt, ok := s.fetchStoryDetail(ctx, summary.ShortID, since)
		if !ok {
			continue
		}
		out = append(out, toSourceItem(detail, occurredAt))
	}
	return out, nil
}

// fetchHottest は hottest.json を取得して summary 列を返す。失敗は List ごと失敗させる。
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

// filterSummariesInWindow は created_at >= since の summary を先頭 limit 件まで返す。
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

// fetchStoryDetail は /s/<short_id>.json を取得し、window 内なら (detail, occurredAt, true) を返す。
// 個別 story の取得・decode 失敗・created_at parse 失敗・window 外は (zero, zero, false) を返す（List は続行）。
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

// getWithRetry は GET を実行し body を返す。
// client.Do error / 5xx は 1 回だけ即再試行（backoff なし）。4xx（429 含む）/ 読み取り失敗は即 return。
func (s *ListItemSource) getWithRetry(ctx context.Context, url, op string) ([]byte, error) {
	body, retryable, err := s.get(ctx, url)
	if err == nil {
		return body, nil
	}
	if !retryable {
		return nil, infraErr(op, err)
	}
	body, _, err = s.get(ctx, url)
	if err != nil {
		return nil, infraErr(op, err)
	}
	return body, nil
}

// get は GET を 1 回実行し body を返す。
// 2 つめの戻り値は再試行してよい失敗（client.Do error / 5xx）かどうか。
func (s *ListItemSource) get(ctx context.Context, url string) (body []byte, retryable bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = res.Body.Close() }()
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, false, err
	}
	if res.StatusCode >= 500 {
		return nil, true, fmt.Errorf("status %d", res.StatusCode)
	}
	if res.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("status %d", res.StatusCode)
	}
	return body, false, nil
}

// parseCreatedAt は offset 付き RFC3339 相当の created_at を UTC に変換する。
func parseCreatedAt(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// collectCommentBodies は deleted / moderated でない comment の comment_plain を先頭 MaxCommentsPerStory 件まで返す。
func collectCommentBodies(comments []storyComment) []string {
	bodies := make([]string, 0, MaxCommentsPerStory)
	for _, c := range comments {
		if c.IsDeleted || c.IsModerated {
			continue
		}
		if c.CommentPlain == "" {
			continue
		}
		bodies = append(bodies, c.CommentPlain)
		if len(bodies) >= MaxCommentsPerStory {
			break
		}
	}
	return bodies
}

// toSourceItem は story 詳細と検証済み occurredAt から SourceItem を組む。
func toSourceItem(detail storyDetail, occurredAt time.Time) models.SourceItem {
	lines := []string{
		"item_id: " + detail.ShortID,
	}
	if detail.SubmitterUser != "" {
		lines = append(lines, "actor_id: "+detail.SubmitterUser, "actor_name: "+detail.SubmitterUser)
	}
	if detail.Title != "" {
		lines = append(lines, "title: "+detail.Title)
	}

	texts := make([]string, 0, MaxCommentsPerStory+1)
	if detail.DescriptionPlain != "" {
		texts = append(texts, detail.DescriptionPlain)
	}
	texts = append(texts, collectCommentBodies(detail.Comments)...)
	if len(texts) > 0 {
		lines = append(lines, "text: "+strings.Join(texts, "\n"))
	}

	permalink := detail.ShortIDURL
	if permalink == "" {
		permalink = detail.CommentsURL
	}
	if permalink != "" {
		lines = append(lines, "permalink: "+permalink)
	}
	if detail.URL != "" {
		lines = append(lines, "links: "+detail.URL)
	}

	return models.SourceItem{
		SourceID:   SourceID,
		OccurredAt: occurredAt.UTC(),
		Context:    strings.Join(lines, "\n"),
	}
}
