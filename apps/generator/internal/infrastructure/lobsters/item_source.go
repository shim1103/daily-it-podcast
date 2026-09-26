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
	MaxStoriesScanned = 20
	// MaxCommentsPerStory は 1 story あたり取得する comment 数の上限。
	MaxCommentsPerStory = 8
	// CommentDepth は取得する comment 階層の深さ（top-level のみ）。
	CommentDepth = 1
)

// ListItemSource は Lobsters の hottest を ItemSource として返す Adapter。
type ListItemSource struct {
	client   *http.Client
	maxItems int
	retry    port.RetryReporter
}

// NewListItemSource は Lobsters 向け ItemSource を返す。
//
// @require httpClient != nil。retry != nil（Composition Root の結線責務）。
// @ensure 戻りは非 nil の *ListItemSource。vendor 固有型を露出しない。
// @ensure maxItems <= 0 の場合、実効上限は MaxStoriesScanned にフォールバックする。
func NewListItemSource(httpClient *http.Client, maxItems int, retry port.RetryReporter) *ListItemSource {
	return &ListItemSource{client: httpClient, maxItems: maxItems, retry: retry}
}

// effectiveMaxStories は min(maxItems, MaxStoriesScanned) を実効上限として返す。
// maxItems <= 0 は呼び出し側の指定漏れとみなし、MaxStoriesScanned を安全側の値として使う。
func (s *ListItemSource) effectiveMaxStories() int {
	if s.maxItems <= 0 || s.maxItems > MaxStoriesScanned {
		return MaxStoriesScanned
	}
	return s.maxItems
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
// @ensure 結果は created_at >= since を満たす story のみ。最大 min(maxItems, MaxStoriesScanned) 件（maxItems <= 0 は MaxStoriesScanned）。
// @ensure 該当なしは空 slice（nil ではない）。
// @invariant vendor 固有型・監視対象一覧を露出しない。Summary / Detail / Discourse を key として解釈しない。
func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("list", fmt.Errorf("client is nil"))
	}

	summaries, err := s.fetchHottest(ctx)
	if err != nil {
		return nil, err
	}

	targets := filterSummariesInWindow(summaries, since, s.effectiveMaxStories())

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

// getWithRetry は httpget へ委譲し、失敗を Adapter の infraErr で包む。
func (s *ListItemSource) getWithRetry(ctx context.Context, url, op string) ([]byte, error) {
	body, err := httpget.GetWithRetry(ctx, s.client, url, s.retry, op)
	if err != nil {
		return nil, infraErr(op, err)
	}
	return body, nil
}

// parseCreatedAt は offset 付き RFC3339 相当の created_at を UTC に変換する。
func parseCreatedAt(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// collectCommentBodies は deleted / moderated でない comment_plain を先頭 MaxCommentsPerStory 件まで返す。
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

// toSourceItem は story 詳細と検証済み occurredAt から SourceItem を組む。
// 写像の方針は Decision 2026-09-13T17-14-10。本 func が lobsters 写像の正本。
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
