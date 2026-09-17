package build

import (
	_ "embed"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

//go:embed writer_output_example.json
var writerOutputExampleJSON string

// ComposeBrief は Fetch 結果から TextWriter へ渡す brief 平文 1 本を組み立てる。
// 固定 Prompt は entities/constants.TextWriterBriefPrompt。期待 topic 数は
// constants.DraftTopicCountTarget（本番固定値）を使う。本 func は埋め込みのみ。
//
// @require items は Fetch 成功後の slice。
// @ensure len(items) > 0 のとき戻りは trim 後に非空の brief 平文 1 本。
// @ensure len(items) == 0 のとき ("", Domain Error Op = no_source_items) を返す。
// @ensure constants.TextWriterBriefPrompt の {{SOURCES}} {{JSON_EXAMPLE}} と数値 placeholder を置換して完成させる。
// @ensure 数値 placeholder は manuscript_draft_limits 定数を embedManuscriptDraftLimits で埋める。{{SOURCES}} は各 item の SourceID・OccurredAt・Summary・Detail・Discourse・Meta を平文列挙（窓幅説明なし）。{{JSON_EXAMPLE}} は writer_output_example.json を読込・検証して埋める。
// @ensure OpeningGreeting / ClosingFarewell は含めない。
// @invariant Prompt 散文を本 package に hardcode しない。Summary / Detail / Discourse / Meta を structured parse しない。
func ComposeBrief(items []models.SourceItem) (string, error) {
	return ComposeBriefWithTemplate(items, constants.TextWriterBriefPrompt, constants.DraftTopicCountTarget)
}

// ComposeBriefWithTemplate は brief template 文字列と期待 topic 数を引数で受け取る
// ComposeBrief の一般版。rate 計測（system && ratemeasure）が prompt variant を
// 差し替えて A/B するための注入口（Decision 2026-09-03T14-47-00）。
// topicCount は system-test 等が任意の topic 数で埋め込み・example 検証を行うための引数
// （Decision 2026-09-16T19-53-19）。
//
// @require items は Fetch 成功後の slice。template は brief prompt template（parse しない）。topicCount >= 0。
// @ensure template の {{…_MIN}} 等の数値 placeholder を topicCount に基づく manuscript_draft_limits 定数で、
//
//	{{SOURCES}} を items の平文列挙で、{{JSON_EXAMPLE}} を writer_output_example.json（topicCount で検証済み）で埋める。
//
// @ensure ComposeBriefWithTemplate(items, constants.TextWriterBriefPrompt, constants.DraftTopicCountTarget) は ComposeBrief(items) と同一出力。
// @ensure len(items) == 0 のとき ("", Domain Error Op = no_source_items) を返す。
// @ensure writer_output_example.json が ManuscriptDraftFromWriterOutput に落ちるとき、その error をそのまま返す。
func ComposeBriefWithTemplate(items []models.SourceItem, template string, topicCount int) (string, error) {
	if len(items) == 0 {
		return "", domainerrors.DomainErr(domainerrors.OpNoSourceItems, nil)
	}

	brief := embedManuscriptDraftLimits(template, topicCount)
	brief = strings.Replace(brief, "{{SOURCES}}", formatSourceItems(items), 1)
	// why: {{JSON_EXAMPLE}} は writer_output_example.json という固定 topic 数
	// （constants.DraftTopicCountTarget 件）の embed fixture であり、topicCount の
	// 実行時値とは独立した「LLM へ見せる wire 形の例示」である。topicCount を渡すと
	// topicCount != DraftTopicCountTarget のとき fixture 自身の topic 数と矛盾し、
	// example が常に invalid 判定になる。
	jsonExample, err := loadWriterOutputExampleJSON(writerOutputExampleJSON, constants.DraftTopicCountTarget)
	if err != nil {
		return "", err
	}
	brief = strings.Replace(brief, "{{JSON_EXAMPLE}}", jsonExample, 1)
	return strings.TrimSpace(brief), nil
}

func formatSourceItems(items []models.SourceItem) string {
	var b strings.Builder
	for i, item := range items {
		if i > 0 {
			b.WriteByte('\n')
			b.WriteByte('\n')
		}
		b.WriteString("source_id: ")
		b.WriteString(item.SourceID)
		b.WriteString("\noccurred_at: ")
		b.WriteString(item.OccurredAt.UTC().Format(time.RFC3339))
		if item.Summary != "" {
			b.WriteString("\nsummary: ")
			b.WriteString(item.Summary)
		}
		appendSourceBody(&b, "detail", item.Detail)
		appendSourceBody(&b, "discourse", item.Discourse)
		if item.Meta != "" {
			b.WriteString("\nmeta: ")
			b.WriteString(item.Meta)
		}
	}
	return b.String()
}

func appendSourceBody(b *strings.Builder, label string, body models.SourceBody) {
	if body.Text != "" {
		b.WriteByte('\n')
		b.WriteString(label)
		b.WriteString(": ")
		b.WriteString(body.Text)
	}
	for _, link := range body.Links {
		if link == "" {
			continue
		}
		b.WriteByte('\n')
		b.WriteString(label)
		b.WriteString("_link: ")
		b.WriteString(link)
	}
}

// loadWriterOutputExampleJSON は {{JSON_EXAMPLE}} 用の WriterOutput JSON 平文を受け取り、
// ManuscriptDraftFromWriterOutput で正当性を検査してから返す。
//
// @require raw は WriterOutput 形の JSON 平文（前後空白可）。expectedTopicCount は raw が持つべき topic 数。
// @ensure 成功時は trim 後の raw を返す。
// @ensure 失敗時は ManuscriptDraftFromWriterOutput の error を wrap せずそのまま返す。
func loadWriterOutputExampleJSON(raw string, expectedTopicCount int) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if _, err := ManuscriptDraftFromWriterOutput(trimmed, expectedTopicCount); err != nil {
		return "", err
	}
	return trimmed, nil
}
