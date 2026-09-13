package techcrunch

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

var _ port.ItemSource = (*ListItemSource)(nil)

// ListItemSource は TechCrunch を ItemSource として返す Adapter（A stub）。
type ListItemSource struct {
	client *http.Client
}

// NewListItemSource は TechCrunch 向け ItemSource を返す。
//
// @require httpClient != nil
// @ensure 戻りは非 nil の *ListItemSource。vendor 固有型を露出しない。
func NewListItemSource(httpClient *http.Client) *ListItemSource {
	return &ListItemSource{client: httpClient}
}

// List は since 以降に発生した TechCrunch item を SourceItem slice で返す。
//
// @require since は OccurredAt の inclusive 下限。
// @ensure 該当なしは空 slice（nil ではない）。A stub は常に空を返す。
// @invariant vendor 固有型・監視対象一覧を露出しない。Summary / Detail / Discourse を key として解釈しない。
func (s *ListItemSource) List(_ context.Context, _ time.Time) ([]models.SourceItem, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("list", fmt.Errorf("client is nil"))
	}
	return []models.SourceItem{}, nil
}
