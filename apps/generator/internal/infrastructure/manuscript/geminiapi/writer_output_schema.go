package geminiapi

import (
	"encoding/json"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// writerOutputGenerationConfig は models.WriterOutputSchema を responseJsonSchema に載せる。
// why: schema は WriterOutput の json tag から生成する。手書き .schema.json を置かない。
//
//	生成する schema は JSON Schema（additionalProperties を含む）であり、OpenAPI 3.0 subset の
//	`responseSchema`（Schema object）は additionalProperties を受け付けず 400 INVALID_ARGUMENT になる
//	（generateContent 本番 System test run 36538248785）。JSON Schema は `responseJsonSchema` へ載せる。
//	両 field は排他で、いずれも responseMimeType=application/json が必須。
var writerOutputGenerationConfig = generationConfig{
	ResponseMIMEType:   "application/json",
	ResponseJSONSchema: json.RawMessage(models.WriterOutputSchema),
}

type generationConfig struct {
	ResponseMIMEType   string          `json:"responseMimeType"`
	ResponseJSONSchema json.RawMessage `json:"responseJsonSchema"`
}
