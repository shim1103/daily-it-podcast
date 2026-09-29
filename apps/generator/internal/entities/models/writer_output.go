package models

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
)

// WriterOutputSchema は WriterOutput の json tag から組み立てた Gemini responseSchema 用 JSON。
// 手書き .schema.json は置かない。wire 形の正本は WriterOutput 型。
var WriterOutputSchema = mustWriterOutputSchemaJSON()

// WriterOutputExampleJSON は brief prompt の {{JSON_EXAMPLE}} 用 fixture（go:embed）。
//
//go:embed writer_output_example.json
var WriterOutputExampleJSON string

// WriterOutput は TextWriter 成功戻り string の JSON wire 形。
// json.Unmarshal の正本。Domain validation 後に ManuscriptDraft へ写す。
// Decision: docs/decisions/2026-08-29T15-00-00-docs-produce-episode-run-spec-writer-output-json-wire.md
type WriterOutput struct {
	Title         string              `json:"title"`
	OpeningIntro  string              `json:"openingIntro"`
	Topics        []WriterOutputTopic `json:"topics"`
	EndingSummary string              `json:"endingSummary"`
}

// WriterOutputTopic は WriterOutput 内の 1 トピック分。
type WriterOutputTopic struct {
	Title   string `json:"title"`
	Preface string `json:"preface"`
	Detail  string `json:"detail"`
}

func mustWriterOutputSchemaJSON() []byte {
	schema := objectSchemaFrom(reflect.TypeOf(WriterOutput{}))
	b, err := json.Marshal(schema)
	if err != nil {
		panic(fmt.Sprintf("WriterOutput schema marshal: %v", err))
	}
	return b
}

// objectSchemaFrom は struct の json tag から Gemini responseSchema 向け object schema を組む。
// string / []T（T は struct）だけを扱う。Domain 制約（件数・rune・日本語）は載せない。
func objectSchemaFrom(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("objectSchemaFrom: want struct, got %s", t.Kind()))
	}

	properties := map[string]any{}
	required := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name, ok := jsonFieldName(f)
		if !ok {
			continue
		}
		properties[name] = propertySchemaFrom(f.Type)
		required = append(required, name)
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             required,
		"properties":           properties,
	}
}

func propertySchemaFrom(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Slice:
		return map[string]any{
			"type":  "array",
			"items": objectSchemaFrom(t.Elem()),
		}
	case reflect.Struct:
		return objectSchemaFrom(t)
	default:
		panic(fmt.Sprintf("propertySchemaFrom: unsupported kind %s", t.Kind()))
	}
}

func jsonFieldName(f reflect.StructField) (string, bool) {
	tag := f.Tag.Get("json")
	if tag == "" || tag == "-" {
		return "", false
	}
	name := tag
	if i := indexByte(tag, ','); i >= 0 {
		name = tag[:i]
	}
	if name == "" || name == "-" {
		return "", false
	}
	return name, true
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
