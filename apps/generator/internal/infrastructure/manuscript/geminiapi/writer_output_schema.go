package geminiapi

import (
	"encoding/json"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// writerOutputGenerationConfig は models.WriterOutputSchema を responseSchema に載せる。
// why: schema は WriterOutput の json tag から生成する。手書き .schema.json を置かない。
var writerOutputGenerationConfig = generationConfig{
	ResponseMIMEType: "application/json",
	ResponseSchema:   json.RawMessage(models.WriterOutputSchema),
}

type generationConfig struct {
	ResponseMIMEType string          `json:"responseMimeType"`
	ResponseSchema   json.RawMessage `json:"responseSchema"`
}
