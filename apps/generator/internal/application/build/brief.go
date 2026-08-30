package build

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// ComposeBrief は Fetch 結果から TextWriter へ渡す brief 平文 1 本を組み立てる。
// 固定 Prompt は entities/constants.TextWriterBriefPrompt。本 func は埋め込みのみ。
//
// @require items は Fetch 成功後の slice。
// @ensure len(items) > 0 のとき戻りは trim 後に非空の brief 平文 1 本。
// @ensure len(items) == 0 のとき ("", Domain Error Op = no_source_items) を返す。
// @ensure constants.TextWriterBriefPrompt の {{SOURCES}} {{JSON_EXAMPLE}} と数値 placeholder を置換して完成させる。
// @ensure 数値 placeholder は manuscript_draft_limits 定数を embedManuscriptDraftLimits で埋める。{{SOURCES}} は各 item の SourceID・OccurredAt・Context を平文列挙（窓幅説明なし）。{{JSON_EXAMPLE}} は models.WriterOutput から生成。
// @ensure OpeningGreeting / ClosingFarewell は含めない。
// @invariant Prompt 散文を本 package に hardcode しない。Context を structured parse しない。
func ComposeBrief(items []models.SourceItem) (string, error) {
	if len(items) == 0 {
		return "", domainerrors.DomainErr(domainerrors.OpNoSourceItems, nil)
	}

	brief := embedManuscriptDraftLimits(constants.TextWriterBriefPrompt)
	brief = strings.Replace(brief, "{{SOURCES}}", formatSourceItems(items), 1)
	jsonExample, err := marshalWriterOutputExample()
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
		b.WriteByte('\n')
		b.WriteString(item.Context)
	}
	return b.String()
}

func marshalWriterOutputExample() (string, error) {
	example := models.WriterOutput{
		Title: "例: エピソードタイトル",
		Intro: "例: 導入文。",
		Topics: []models.WriterOutputTopic{
			{
				Title:   "例: トピック題名",
				Preface: "例: 前置き。",
				Detail:  "例: 本文。",
			},
		},
		ClosingSummary: "例: まとめ。",
	}
	data, err := json.Marshal(example)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
