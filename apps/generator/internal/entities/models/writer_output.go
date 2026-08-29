package models

// WriterOutput は TextWriter 成功戻り string の JSON wire 形。
// json.Unmarshal の正本。Domain validation 後に ManuscriptDraft へ写す。
// Decision: docs/decisions/2026-08-29T15-00-00-docs-produce-episode-run-spec-writer-output-json-wire.md
type WriterOutput struct {
	Title          string              `json:"title"`
	Intro          string              `json:"intro"`
	Topics         []WriterOutputTopic `json:"topics"`
	ClosingSummary string              `json:"closingSummary"`
}

// WriterOutputTopic は WriterOutput 内の 1 トピック分。
type WriterOutputTopic struct {
	Title   string `json:"title"`
	Preface string `json:"preface"`
	Detail  string `json:"detail"`
}
