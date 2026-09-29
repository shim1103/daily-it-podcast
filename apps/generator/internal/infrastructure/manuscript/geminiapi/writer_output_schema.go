package geminiapi

import "encoding/json"

// writerOutputResponseSchema は ManuscriptDraft wire（models.WriterOutput）の形を拘束する
// generateContent responseSchema。topic 件数・rune 数・日本語含有は Domain 検証のまま。
// why: Cursor Cloud Agents に structured 出力が無いため Gemini Adapter だけが持つ（Decision）。
var writerOutputResponseSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "title": { "type": "string" },
    "intro": { "type": "string" },
    "topics": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "title": { "type": "string" },
          "preface": { "type": "string" },
          "detail": { "type": "string" }
        },
        "required": ["title", "preface", "detail"]
      }
    },
    "closingSummary": { "type": "string" }
  },
  "required": ["title", "intro", "topics", "closingSummary"]
}`)

// writerOutputGenerationConfig は JSON のみ・schema 拘束付きの generationConfig。
var writerOutputGenerationConfig = generationConfig{
	ResponseMIMEType: "application/json",
	ResponseSchema:   writerOutputResponseSchema,
}

type generationConfig struct {
	ResponseMIMEType string          `json:"responseMimeType"`
	ResponseSchema   json.RawMessage `json:"responseSchema"`
}
