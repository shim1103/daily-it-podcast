package hackernews

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

var _ port.ItemSource = (*ListItemSource)(nil)

// apiBaseURL は Hacker News Firebase API v0 の base。
const apiBaseURL = "https://hacker-news.firebaseio.com/v0"

// 取得上限（この file が契約として値を固定する）。
//
// why: podcast は 1 日 1 回だけ生成する。上限件数に達したら走査を打ち切るので、
// 1 run あたりの fetch 回数も有界に収まる。
const (
	// MaxStoriesScanned は結果 SourceItem に含める story 数の上限。
	// time >= since を満たした story を、topstories.json の id 列の先頭からこの件数まで集める。
	//
	// topstories.json の id をこの件数まで見る、という意味ではない。topstories.json は
	// time 順でなくランキング順に並ぶため、window 内 story を探して id 列を走査する。
	// window 内 story が上限未満の日は id 列を最後まで走査するが、id 列は最大 500 件なので
	// fetch 回数は min(len(ids), 探索必要数) で有界。
	//
	// 名前が「Scanned」だと誤解を生むが、契約値の識別子変更は scope 外なので名前は変えない。
	MaxStoriesScanned = 30
	// MaxCommentsPerStory は 1 story あたり取得する top-level comment 数の上限。
	MaxCommentsPerStory = 8
	// CommentDepth は取得する comment 階層の深さ（top-level のみ）。
	CommentDepth = 1
)

// permalinkBaseURL は Context の permalink 行が指す Hacker News item ページ。
const permalinkBaseURL = "https://news.ycombinator.com/item?id="

// ListItemSource は Hacker News の topstories を ItemSource として返す Adapter。
type ListItemSource struct {
	client *http.Client
}

// NewListItemSource は Hacker News 向け ItemSource を返す。
//
// @require httpClient != nil
// @ensure 戻りは非 nil の *ListItemSource。vendor 固有型を露出しない。
func NewListItemSource(httpClient *http.Client) *ListItemSource {
	return &ListItemSource{client: httpClient}
}

// hnItem は Hacker News item/<id>.json のうち Adapter が使う field だけを表す。
// package 外へ露出しない unexported 型。
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

// List は since 以降に発生した Hacker News story を SourceItem slice で返す。
//
// @require since は OccurredAt の inclusive 下限。
// @ensure 各要素の SourceID は非空（= SourceID）。OccurredAt は UTC かつ since 以上。
// @ensure 結果は time >= since を満たす story のみ。最大 MaxStoriesScanned 件。
// @ensure 該当なしは空 slice（nil ではない）。
// @invariant vendor 固有型・監視対象一覧を露出しない。Context を key として解釈しない。
func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("list", fmt.Errorf("client is nil"))
	}

	ids, err := s.fetchTopStoryIDs(ctx)
	if err != nil {
		return nil, err
	}

	// why: topstories.json は time 順でなくランキング順。window 内 story を探して
	// id 列を先頭から走査し、結果が MaxStoriesScanned 件に達したら打ち切る。
	// 最悪ケース（window 内が上限未満）でも id 列は最大 500 件で有界。
	out := make([]models.SourceItem, 0, MaxStoriesScanned)
	for _, id := range ids {
		story, ok := s.fetchStoryInWindow(ctx, id, since)
		if !ok {
			continue
		}
		comments := s.fetchTopLevelComments(ctx, story.Kids)
		out = append(out, toSourceItem(story, comments))
		if len(out) >= MaxStoriesScanned {
			break
		}
	}
	return out, nil
}

// fetchTopStoryIDs は topstories.json を取得して id 列を返す。失敗は List ごと失敗させる。
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

// fetchStoryInWindow は item/<id>.json を取得し、story かつ window 内なら (item, true) を返す。
// 個別 item の取得・decode 失敗はその要素を落として (zero, false) を返す（List は続行）。
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

// fetchTopLevelComments は kids 先頭 MaxCommentsPerStory 件の comment 本文を取得順に返す。
// CommentDepth=1: comment の kids は辿らない。個別 comment の取得失敗はその要素を落とす。
func (s *ListItemSource) fetchTopLevelComments(ctx context.Context, kids []int64) []string {
	limit := MaxCommentsPerStory
	if len(kids) < limit {
		limit = len(kids)
	}
	bodies := make([]string, 0, limit)
	for _, kid := range kids[:limit] {
		item, err := s.fetchItem(ctx, kid)
		if err != nil {
			continue
		}
		normalized := normalizeHTML(item.Text)
		if normalized == "" {
			continue
		}
		bodies = append(bodies, normalized)
	}
	return bodies
}

// fetchItem は item/<id>.json を 1 件取得して decode する。
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
	// why: 2 回目は retryable を問わず打ち切る（再試行は 1 回だけ）。
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

// toSourceItem は story と取得済み comment 本文から SourceItem を組む。
func toSourceItem(story hnItem, commentBodies []string) models.SourceItem {
	lines := []string{
		fmt.Sprintf("item_id: %d", story.ID),
	}
	if story.By != "" {
		lines = append(lines, "actor_id: "+story.By, "actor_name: "+story.By)
	}
	if story.Title != "" {
		lines = append(lines, "title: "+story.Title)
	}

	texts := make([]string, 0, len(commentBodies)+1)
	if storyText := normalizeHTML(story.Text); storyText != "" {
		texts = append(texts, storyText)
	}
	texts = append(texts, commentBodies...)
	if len(texts) > 0 {
		lines = append(lines, "text: "+strings.Join(texts, "\n"))
	}

	lines = append(lines, fmt.Sprintf("permalink: %s%d", permalinkBaseURL, story.ID))
	if story.URL != "" {
		lines = append(lines, "links: "+story.URL)
	}

	return models.SourceItem{
		SourceID:   SourceID,
		OccurredAt: time.Unix(story.Time, 0).UTC(),
		Context:    strings.Join(lines, "\n"),
	}
}

// pTagReplacer は <p> 段落区切りを改行へ変換する。
// why: normalizeHTML は 1 run で最大数百回呼ばれる。Replacer 生成を 1 回へ巻き上げる。
var pTagReplacer = strings.NewReplacer("<p>", "\n", "<P>", "\n")

// normalizeHTML は HTML entity を unescape し、<p> を改行へ、他タグを除去する軽い正規化。
// token 効率が目的で full readability はしない。
func normalizeHTML(s string) string {
	if s == "" {
		return ""
	}
	s = pTagReplacer.Replace(s)
	s = stripTags(s)
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}

// stripTags は "<" と ">" で囲まれた token を除去する。
func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}
