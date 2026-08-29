package build_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

func assertDomainOp(t *testing.T, err error, wantOp string) {
	t.Helper()
	var domErr *domainerrors.Error
	if !errors.As(err, &domErr) {
		t.Fatalf("error type %T (%v), want *domainerrors.Error", err, err)
	}
	if domErr.Op != wantOp {
		t.Fatalf("Op = %q, want %q", domErr.Op, wantOp)
	}
}

func TestComposeBrief_returnsDomainErrorWithNoSourceItemsOp_whenItemsLengthZero(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		items []models.SourceItem
	}{
		{name: "nil slice", items: nil},
		{name: "empty slice", items: []models.SourceItem{}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given: 0 件の items
			items := tc.items

			// When: brief を組み立てる
			got, err := build.ComposeBrief(items)

			// Then: 空 brief と Domain Error（Op = no_source_items）
			if got != "" {
				t.Fatalf("brief = %q, want empty", got)
			}
			assertDomainOp(t, err, domainerrors.OpNoSourceItems)
		})
	}
}

func TestComposeBrief_returnsTrimmedBriefWithoutPlaceholders_whenSingleItemGiven(t *testing.T) {
	t.Parallel()

	// Given: 代表 1 件の SourceItem
	items := []models.SourceItem{
		{
			SourceID:   "x-api",
			OccurredAt: time.Date(2024, 12, 10, 10, 0, 0, 0, time.UTC),
			Context:    "item_id: tweet-1",
		},
	}

	// When: brief を組み立てる
	got, err := build.ComposeBrief(items)

	// Then: trim 後に非空で placeholder が残らず、Prompt 骨格と有効な JSON example を含む
	if err != nil {
		t.Fatalf("ComposeBrief: %v", err)
	}
	if strings.TrimSpace(got) != got {
		t.Fatalf("ComposeBrief: 戻りが trim されていない")
	}
	if got == "" {
		t.Fatal("ComposeBrief: 戻りが空")
	}
	if strings.Contains(got, "{{") || strings.Contains(got, "}}") {
		t.Fatalf("ComposeBrief: placeholder が残っている: %q", got)
	}
	if !strings.Contains(got, "あなたは IT ニュースを一人で解説する podcast ナレーターです。") {
		t.Fatal("ComposeBrief: TextWriterBriefPrompt 本文が欠けている")
	}
	jsonExample := extractJSONExample(t, got)
	var wire models.WriterOutput
	if err := json.Unmarshal([]byte(jsonExample), &wire); err != nil {
		t.Fatalf("JSON example の Unmarshal: %v\njson: %s", err, jsonExample)
	}
	if wire.Title == "" || wire.Intro == "" || wire.ClosingSummary == "" {
		t.Fatalf("WriterOutput の主要 field が空: %+v", wire)
	}
	if len(wire.Topics) == 0 {
		t.Fatal("WriterOutput.Topics が空")
	}
	for i, topic := range wire.Topics {
		if topic.Title == "" || topic.Preface == "" || topic.Detail == "" {
			t.Fatalf("topics[%d] に空 field がある: %+v", i, topic)
		}
	}
}

func TestComposeBrief_embedsSourcesPlainText_whenItemsGiven(t *testing.T) {
	t.Parallel()

	occurredUTC := time.Date(2024, 12, 10, 10, 0, 0, 0, time.UTC)
	occurredJST := time.Date(2024, 12, 10, 11, 30, 0, 0, time.FixedZone("JST", 9*60*60))

	cases := []struct {
		name        string
		items       []models.SourceItem
		wantSources string
	}{
		{
			name: "embedsSingleSource_whenOneItemGiven",
			items: []models.SourceItem{
				{
					SourceID:   "x-api",
					OccurredAt: occurredUTC,
					Context:    "item_id: tweet-1",
				},
			},
			wantSources: strings.Join([]string{
				"source_id: x-api",
				"occurred_at: 2024-12-10T10:00:00Z",
				"item_id: tweet-1",
			}, "\n"),
		},
		{
			name: "embedsMultipleSourcesSeparatedByBlankLine_whenTwoItemsGiven",
			items: []models.SourceItem{
				{
					SourceID:   "x-api",
					OccurredAt: occurredUTC,
					Context:    "item_id: tweet-1\nbody: hello",
				},
				{
					SourceID:   "rss-feed",
					OccurredAt: occurredJST,
					Context:    "item_id: article-9",
				},
			},
			wantSources: strings.Join([]string{
				"source_id: x-api",
				"occurred_at: 2024-12-10T10:00:00Z",
				"item_id: tweet-1",
				"body: hello",
				"",
				"source_id: rss-feed",
				"occurred_at: 2024-12-10T02:30:00Z",
				"item_id: article-9",
			}, "\n"),
		},
		{
			name: "normalizesOccurredAtToUTC_whenOccurredAtHasNonUTCZone",
			items: []models.SourceItem{
				{
					SourceID:   "rss-feed",
					OccurredAt: occurredJST,
					Context:    "item_id: article-9",
				},
			},
			wantSources: strings.Join([]string{
				"source_id: rss-feed",
				"occurred_at: 2024-12-10T02:30:00Z",
				"item_id: article-9",
			}, "\n"),
		},
		{
			name: "embedsSourceWithoutContextLines_whenContextEmpty",
			items: []models.SourceItem{
				{
					SourceID:   "empty-ctx",
					OccurredAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					Context:    "",
				},
			},
			wantSources: strings.Join([]string{
				"source_id: empty-ctx",
				"occurred_at: 2024-01-01T00:00:00Z",
			}, "\n"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given: SourceItem slice
			items := tc.items

			// When: brief を組み立てる
			got, err := build.ComposeBrief(items)

			// Then: 期待する sources 平文が埋まる
			if err != nil {
				t.Fatalf("ComposeBrief: %v", err)
			}
			if !strings.Contains(got, tc.wantSources) {
				t.Fatalf("ComposeBrief: sources 列挙が不正\nwant substring:\n%s\ngot:\n%s", tc.wantSources, got)
			}
		})
	}
}

func extractJSONExample(t *testing.T, brief string) string {
	t.Helper()
	const marker = "# Example shape"
	idx := strings.Index(brief, marker)
	if idx < 0 {
		t.Fatal("brief に # Example shape が無い")
	}
	rest := brief[idx+len(marker):]
	brace := strings.Index(rest, "{")
	if brace < 0 {
		t.Fatal("JSON example の開始 { が無い")
	}
	return strings.TrimSpace(rest[brace:])
}
