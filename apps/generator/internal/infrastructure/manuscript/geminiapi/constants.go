package geminiapi

import "time"

const (
	// ModelID は原稿生成に使う Gemini model の SSoT。
	// why: 同社 TTS で使う GEMINI_API_KEY を流用でき、Gemini の中で最も安い現行 Flash-Lite。
	// 版更新はこの 1 行だけで閉じる（runtime の model 存在確認はしない。docs は値を写さず geminiapi.ModelID を参照する）。
	ModelID = "gemini-3.1-flash-lite"
	// EndpointURLTemplate は generateContent の URL template。%s は ModelID。
	EndpointURLTemplate = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"
	// APIKeyHeader は API key を渡す HTTP header 名。
	// why: key を URL query に載せず header に載せる。error / log に URL を出しても key が漏れない。
	APIKeyHeader = "x-goog-api-key"
)

// MaxAttempts は Retry-After 付き 429 に対する最大試行数。無限 retry を防ぐ。cursorapi と同値。
// why: 原稿用 model の RPD は現状の 1 日 1 回 produce 運用では使い切りにくく、
// gemini（TTS）のように 1 episode で焼き切る動機が無い。Tier による差は付けない。
const MaxAttempts = 4

// TextWriterMaxAttempts は ManuscriptDraft 検証失敗（invalid-draft）時の Write 内部 retry 上限。
// LLM 出力の rune 数・topic 数の揺れを吸収する。429 用の MaxAttempts とは別物。
const TextWriterMaxAttempts = 5

// MaxRetryAfter は Retry-After header 由来の待ち時間の上限。
// why: 異常に長い Retry-After に待たされない。
const MaxRetryAfter = 30 * time.Second

// ResponseBufferBytes は generateContent 応答（原稿 JSON を含む）を読む上限。
// why: 想定最大原稿の UTF-8 byte 数に、応答 envelope（candidates / usageMetadata 等）の余裕を足す。
const ResponseBufferBytes = 1 << 20

// why: 1 Write は invalid-draft retry で複数回 fetch しうるので、TTS 側の httpCallTimeout より長く取る。
const textWriterHTTPTimeout = 10 * time.Minute
