package models_test

import (
	"encoding/json"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// TestWriterOutputSchema_listsJSONTagsFromWriterOutput は responseSchema が
// WriterOutput の json tag から組み立てられ、手書き schema file に依存しないことを固定する。
func TestWriterOutputSchema_listsJSONTagsFromWriterOutput(t *testing.T) {
	t.Parallel()

	// Given: models.WriterOutputSchema（型から生成）
	var schema struct {
		Type                 string         `json:"type"`
		AdditionalProperties bool           `json:"additionalProperties"`
		Required             []string       `json:"required"`
		Properties           map[string]any `json:"properties"`
	}

	// When: schema JSON を decode する
	if err := json.Unmarshal(models.WriterOutputSchema, &schema); err != nil {
		t.Fatalf("Unmarshal WriterOutputSchema: %v", err)
	}

	// Then: object・追加禁止・required / properties が WriterOutput wire field と一致
	if schema.Type != "object" {
		t.Fatalf("type = %q, want object", schema.Type)
	}
	if schema.AdditionalProperties {
		t.Fatal("additionalProperties = true, want false")
	}
	wantRequired := []string{"title", "openingIntro", "topics", "endingSummary"}
	if len(schema.Required) != len(wantRequired) {
		t.Fatalf("required = %v, want %v", schema.Required, wantRequired)
	}
	for i, name := range wantRequired {
		if schema.Required[i] != name {
			t.Fatalf("required[%d] = %q, want %q", i, schema.Required[i], name)
		}
		if _, ok := schema.Properties[name]; !ok {
			t.Fatalf("properties に %q が無い", name)
		}
	}
	topics, ok := schema.Properties["topics"].(map[string]any)
	if !ok {
		t.Fatalf("topics property type = %T, want object", schema.Properties["topics"])
	}
	if topics["type"] != "array" {
		t.Fatalf("topics.type = %v, want array", topics["type"])
	}
	items, ok := topics["items"].(map[string]any)
	if !ok {
		t.Fatalf("topics.items type = %T, want object", topics["items"])
	}
	itemRequired, ok := items["required"].([]any)
	if !ok {
		t.Fatalf("topics.items.required type = %T", items["required"])
	}
	for i, want := range []string{"title", "preface", "detail"} {
		if itemRequired[i] != want {
			t.Fatalf("topics.items.required[%d] = %v, want %q", i, itemRequired[i], want)
		}
	}
}

// TestWriterOutput_roundTripJSON_usesOpeningIntroAndEndingSummary は wire key 名を固定する。
func TestWriterOutput_roundTripJSON_usesOpeningIntroAndEndingSummary(t *testing.T) {
	t.Parallel()

	// Given: 新 field 名の WriterOutput
	in := models.WriterOutput{
		Title:         "題",
		OpeningIntro:  "導入。",
		Topics:        []models.WriterOutputTopic{{Title: "題", Preface: "前。", Detail: "本。"}},
		EndingSummary: "まとめ。",
	}

	// When: marshal → unmarshal
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out models.WriterOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	// Then: 値は一致し、旧 key（intro / closingSummary）は生 JSON に出ない
	if out.Title != in.Title || out.OpeningIntro != in.OpeningIntro || out.EndingSummary != in.EndingSummary {
		t.Fatalf("round-trip scalars = %+v, want %+v", out, in)
	}
	if len(out.Topics) != 1 || out.Topics[0] != in.Topics[0] {
		t.Fatalf("round-trip topics = %+v, want %+v", out.Topics, in.Topics)
	}
	var asMap map[string]any
	if err := json.Unmarshal(raw, &asMap); err != nil {
		t.Fatalf("Unmarshal map: %v", err)
	}
	for _, banned := range []string{"intro", "closingSummary"} {
		if _, ok := asMap[banned]; ok {
			t.Fatalf("wire に旧 key %q が残っている: %s", banned, raw)
		}
	}
	for _, want := range []string{"openingIntro", "endingSummary"} {
		if _, ok := asMap[want]; !ok {
			t.Fatalf("wire に %q が無い: %s", want, raw)
		}
	}
}
