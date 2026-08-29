package build

import "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"

// ComposeBrief は Fetch 結果から TextWriter へ渡す brief 平文 1 本を組み立てる。
// 固定 Prompt は entities/constants.TextWriterBriefPrompt。本 func は埋め込みのみ。
//
// @require items は Fetch 成功後の slice（0 件は呼び出し側が Domain Error にする）。
// @ensure 戻りは trim 後に非空の brief 平文 1 本。
// @ensure constants.TextWriterBriefPrompt の {{SOURCES}} {{JSON_EXAMPLE}} と数値 placeholder を置換して完成させる。
// @ensure 数値 placeholder は manuscript_draft_limits 定数を embedManuscriptDraftLimits で埋める。{{SOURCES}} は各 item の SourceID・OccurredAt・Context を平文列挙（窓幅説明なし）。{{JSON_EXAMPLE}} は models.WriterOutput から生成。
// @ensure OpeningGreeting / ClosingFarewell は含めない。
// @invariant Prompt 散文を本 package に hardcode しない。Context を structured parse しない。
func ComposeBrief(items []models.SourceItem) string {
	panic("compose brief: contract stub; logic is C")
}
