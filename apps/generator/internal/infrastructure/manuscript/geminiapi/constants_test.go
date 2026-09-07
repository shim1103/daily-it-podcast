package geminiapi

import (
	"fmt"
	"strings"
	"testing"
)

// Scope: 定数の不変条件
// why: Constraints「key は URL query に載せず header に載せる」を定数レベルで固定する。
// endpoint template が query（? / key=）を含むと、apiKey を query へ載せる実装事故を招く。

func TestEndpointURLTemplate_carriesNoQueryString(t *testing.T) {
	t.Parallel()

	// Given: ModelID を埋めた実 URL
	url := fmt.Sprintf(EndpointURLTemplate, ModelID)

	// Then: query 区切り "?" も "key=" も含まない
	if strings.Contains(url, "?") {
		t.Fatalf("URL %q contains a query string ('?')", url)
	}
	if strings.Contains(url, "key=") {
		t.Fatalf("URL %q carries a key query param", url)
	}
	// Then: generateContent の path で終わる
	if !strings.HasSuffix(url, ":generateContent") {
		t.Fatalf("URL %q does not end with :generateContent", url)
	}
}

func TestAPIKeyHeader_isTheGoogleAPIKeyHeader(t *testing.T) {
	t.Parallel()

	// why: key は必ずこの header に載る。定数の取り違えを固定する。
	if APIKeyHeader != "x-goog-api-key" {
		t.Fatalf("APIKeyHeader = %q, want %q", APIKeyHeader, "x-goog-api-key")
	}
}
