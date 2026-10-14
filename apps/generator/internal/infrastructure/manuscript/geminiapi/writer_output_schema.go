package geminiapi

import (
	"encoding/json"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// why: schema は WriterOutput の json tag から生成する。additionalProperties を含む JSON Schema は
// OpenAPI subset の responseSchema が拒否するため、models.WriterOutputSchema の載せ先は responseJsonSchema にする。
var writerOutputGenerationConfig = generationConfig{
	ResponseMIMEType:   "application/json",
	ResponseJSONSchema: json.RawMessage(models.WriterOutputSchema),
}

type generationConfig struct {
	ResponseMIMEType   string          `json:"responseMimeType"`
	ResponseJSONSchema json.RawMessage `json:"responseJsonSchema"`
}
