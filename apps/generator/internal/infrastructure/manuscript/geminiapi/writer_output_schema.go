package geminiapi

import (
	"encoding/json"

	"github.com/shim1103/daily-it-podcast/contracts"
)

// writerOutputGenerationConfig は contracts/writer-output.schema.json を responseSchema に載せる。
// why: WriterOutput 形の正本は contracts。Adapter 内へ schema JSON を手写ししない（DRY）。
var writerOutputGenerationConfig = generationConfig{
	ResponseMIMEType: "application/json",
	ResponseSchema:   json.RawMessage(contracts.WriterOutputSchema),
}

type generationConfig struct {
	ResponseMIMEType string          `json:"responseMimeType"`
	ResponseSchema   json.RawMessage `json:"responseSchema"`
}
