package geminiapi

import (
	"encoding/json"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// writerOutputGenerationConfig は models.WriterOutputSchema を responseSchema に載せる。
// why: WriterOutput 形の正本は entities/models（json.Unmarshal 型と同 dir）。
// Adapter 内へ schema JSON を手写ししない（DRY）。contracts/ は完成 manuscript 用。
var writerOutputGenerationConfig = generationConfig{
	ResponseMIMEType: "application/json",
	ResponseSchema:   json.RawMessage(models.WriterOutputSchema),
}

type generationConfig struct {
	ResponseMIMEType string          `json:"responseMimeType"`
	ResponseSchema   json.RawMessage `json:"responseSchema"`
}
