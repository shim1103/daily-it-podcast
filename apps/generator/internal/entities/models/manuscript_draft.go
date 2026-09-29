package models

// ManuscriptDraft は完成 manuscript.schema.json ではない途中型。
// TextWriter の text 断片から ProduceEpisode（Builder）が解釈して作る。
// json tag は付けない（wire 正本は WriterOutput）。
type ManuscriptDraft struct {
	Title         string
	OpeningIntro  string
	Topics        []ManuscriptDraftTopic
	EndingSummary string
}

// ManuscriptDraftTopic は draft 内の 1 トピック分。
type ManuscriptDraftTopic struct {
	Title   string
	Preface string
	Detail  string
}
